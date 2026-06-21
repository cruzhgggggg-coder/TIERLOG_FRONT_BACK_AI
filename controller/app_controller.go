package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"testing_go/auth"
	"testing_go/koneksi"
	"testing_go/middleware"
	"testing_go/models"
	"testing_go/realtime"
	"testing_go/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var WebSocketHub *realtime.Hub

type ClassificationItem struct {
	ID       uint64 `json:"id"`
	Category string `json:"category"`
}

func SetRealtimeHub(hub *realtime.Hub) {
	WebSocketHub = hub
}

func sanitizeUser(user *models.User) gin.H {
	response := gin.H{
		"id":                user.ID,
		"name":              user.Name,
		"email":             user.Email,
		"role":              user.Role,
		"openai_key":        auth.MaskAPIKey(auth.DecryptAPIKey(user.OpenAIKey)),
		"gemini_key":        auth.MaskAPIKey(auth.DecryptAPIKey(user.GeminiKey)),
		"anthropic_key":     auth.MaskAPIKey(auth.DecryptAPIKey(user.AnthropicKey)),
		"nvidia_key":        auth.MaskAPIKey(auth.DecryptAPIKey(user.NvidiaKey)),
		"groq_key":          auth.MaskAPIKey(auth.DecryptAPIKey(user.GroqKey)),
		"preferred_model":   user.PreferredModel,
		"is_gateway_active": user.IsGatewayActive,
		"created_at":        user.CreatedAt,
		"updated_at":        user.UpdatedAt,
	}

	if user.Student != nil {
		response["student"] = user.Student
	}
	if user.Lecturer != nil {
		response["lecturer"] = user.Lecturer
	}

	return response
}

func persistRefreshToken(c *gin.Context, userID uint64, refreshToken string, expiresAt time.Time) error {
	record := models.RefreshToken{
		UserID:    userID,
		TokenHash: auth.HashRefreshToken(refreshToken),
		ExpiresAt: expiresAt,
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: c.ClientIP(),
	}
	return koneksi.DB.Create(&record).Error
}

func authResponse(c *gin.Context, user *models.User) {
	bundle, refreshToken, refreshExpiry, err := auth.TokenBundle(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := persistRefreshToken(c, user.ID, refreshToken, refreshExpiry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":          sanitizeUser(user),
		"access_token":  bundle["access_token"],
		"token_type":    bundle["token_type"],
		"expires_at":    bundle["expires_at"],
		"refresh_token": refreshToken,
	})
}

func Register(c *gin.Context) {
	var req struct {
		Name         string          `json:"name" binding:"required"`
		Email        string          `json:"email" binding:"required,email"`
		Password     string          `json:"password" binding:"required,min=8"`
		Role         models.UserRole `json:"role" binding:"required"`
		NIM          string          `json:"nim"`
		Prodi        string          `json:"prodi"`
		ThesisTitle  string          `json:"thesis_title"`
		LecturerID   uint64          `json:"lecturer_id"`
		NIP          string          `json:"nip"`
		Faculty      string          `json:"faculty"`
		Keahlian     string          `json:"keahlian"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if utf8.RuneCountInString(req.Name) > maxInputNameLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Name must be ≤ %d characters", maxInputNameLength)})
		return
	}
	if len(req.Email) > maxInputEmailLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Email must be ≤ %d characters", maxInputEmailLength)})
		return
	}
	if req.ThesisTitle != "" && utf8.RuneCountInString(req.ThesisTitle) > maxInputContentLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Thesis title must be ≤ %d characters", maxInputContentLength)})
		return
	}

	if req.Role != models.RoleStudent && req.Role != models.RoleLecturer {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role must be student or lecturer"})
		return
	}
	if req.Role == models.RoleStudent && (req.NIM == "" || req.LecturerID == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Student requires a NIM and lecturer_id"})
		return
	}
	if req.Role == models.RoleLecturer && req.NIP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lecturer requires a NIP"})
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		Name:     req.Name,
		Email:    strings.ToLower(strings.TrimSpace(req.Email)),
		Password: hashedPassword,
		Role:     req.Role,
	}

	if err := koneksi.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		switch req.Role {
		case models.RoleStudent:
			student := models.Student{
				UserID:      user.ID,
				LecturerID:  req.LecturerID,
				NIM:         req.NIM,
				Name:        req.Name,
				Prodi:       req.Prodi,
				ThesisTitle: req.ThesisTitle,
			}
			if err := tx.Create(&student).Error; err != nil {
				return err
			}
			user.Student = &student
		case models.RoleLecturer:
			lecturer := models.Lecturer{
				UserID:   user.ID,
				NIP:      req.NIP,
				Name:     req.Name,
				Faculty:  req.Faculty,
				Keahlian: req.Keahlian,
			}
			if err := tx.Create(&lecturer).Error; err != nil {
				return err
			}
			user.Lecturer = &lecturer
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authResponse(c, &user)
}

func Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := koneksi.DB.Preload("Student").Preload("Lecturer").Where("email = ?", strings.ToLower(strings.TrimSpace(req.Email))).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect email or password"})
		return
	}

	if !auth.ComparePassword(user.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect email or password"})
		return
	}

	authResponse(c, &user)
}

func Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenHash := auth.HashRefreshToken(req.RefreshToken)
	var session models.RefreshToken
	if err := koneksi.DB.Where("token_hash = ?", tokenHash).First(&session).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	if session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has expired or has been revoked"})
		return
	}

	now := time.Now()
	session.RevokedAt = &now
	_ = koneksi.DB.Save(&session).Error

	var user models.User
	if err := koneksi.DB.Preload("Student").Preload("Lecturer").First(&user, session.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	authResponse(c, &user)
}

func Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenHash := auth.HashRefreshToken(req.RefreshToken)
	now := time.Now()
	koneksi.DB.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", &now)

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func Me(c *gin.Context) {
	user := middleware.CurrentUser(c)
	c.JSON(http.StatusOK, gin.H{"user": sanitizeUser(user)})
}

func UpdateProfile(c *gin.Context) {
	user := middleware.CurrentUser(c)

	var req struct {
		Name          string `json:"name" binding:"required"`
		Email         string `json:"email" binding:"required,email"`
		NIM           string `json:"nim"`
		Prodi         string `json:"prodi"`
		ThesisTitle   string `json:"thesis_title"`
		LecturerID    uint64 `json:"lecturer_id"`
		NIP           string `json:"nip"`
		Faculty       string `json:"faculty"`
		Keahlian      string `json:"keahlian"`
		AIConstraints string `json:"ai_constraints"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.Name = req.Name
	user.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if err := koneksi.DB.Save(user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user.Role == models.RoleStudent {
		var student models.Student
		if err := koneksi.DB.Where("user_id = ?", user.ID).First(&student).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Student profile not found"})
			return
		}
		student.Name = req.Name
		if req.NIM != "" {
			student.NIM = req.NIM
		}
		if req.LecturerID != 0 {
			student.LecturerID = req.LecturerID
		}
		student.Prodi = req.Prodi
		student.ThesisTitle = req.ThesisTitle
		if err := koneksi.DB.Save(&student).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user.Student = &student
	} else if user.Role == models.RoleLecturer {
		var lecturer models.Lecturer
		if err := koneksi.DB.Where("user_id = ?", user.ID).First(&lecturer).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Lecturer profile not found"})
			return
		}
		lecturer.Name = req.Name
		if req.NIP != "" {
			lecturer.NIP = req.NIP
		}
		lecturer.Faculty = req.Faculty
		lecturer.Keahlian = req.Keahlian
		lecturer.AIConstraints = req.AIConstraints
		if err := koneksi.DB.Save(&lecturer).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user.Lecturer = &lecturer
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully", "user": sanitizeUser(user)})
}

func UpdatePassword(c *gin.Context) {
	user := middleware.CurrentUser(c)

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		Password        string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !auth.ComparePassword(user.Password, req.CurrentPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password does not match"})
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.Password = hashedPassword
	if err := koneksi.DB.Save(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func UpdateAIGatewaySettingsV2(c *gin.Context) {
	user := middleware.CurrentUser(c)

	var req struct {
		OpenAIKey      *string `json:"openai_key"`
		GeminiKey      *string `json:"gemini_key"`
		AnthropicKey   *string `json:"anthropic_key"`
		NvidiaKey      *string `json:"nvidia_key"`
		GroqKey        *string `json:"groq_key"`
		PreferredModel string  `json:"preferred_model"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.OpenAIKey != nil {
		val := *req.OpenAIKey
		if val == "" {
			user.OpenAIKey = ""
		} else if !strings.Contains(val, "*") {
			user.OpenAIKey = auth.EncryptAPIKey(val)
		}
	}
	if req.GeminiKey != nil {
		val := *req.GeminiKey
		if val == "" {
			user.GeminiKey = ""
		} else if !strings.Contains(val, "*") {
			user.GeminiKey = auth.EncryptAPIKey(val)
		}
	}
	if req.AnthropicKey != nil {
		val := *req.AnthropicKey
		if val == "" {
			user.AnthropicKey = ""
		} else if !strings.Contains(val, "*") {
			user.AnthropicKey = auth.EncryptAPIKey(val)
		}
	}
	if req.NvidiaKey != nil {
		val := *req.NvidiaKey
		if val == "" {
			user.NvidiaKey = ""
		} else if !strings.Contains(val, "*") {
			user.NvidiaKey = auth.EncryptAPIKey(val)
		}
	}
	if req.GroqKey != nil {
		val := *req.GroqKey
		if val == "" {
			user.GroqKey = ""
		} else if !strings.Contains(val, "*") {
			user.GroqKey = auth.EncryptAPIKey(val)
		}
	}
	if req.PreferredModel != "" {
		user.PreferredModel = req.PreferredModel
	}

	if err := koneksi.DB.Save(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "AI Gateway settings updated successfully", "user": sanitizeUser(user)})
}

func RedeemGatewayCodeV2(c *gin.Context) {
	user := middleware.CurrentUser(c)

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var redeemCode models.RedeemCode
	if err := koneksi.DB.Where("code = ? AND is_used = ?", req.Code, false).First(&redeemCode).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Redeem code is invalid or has already been used"})
		return
	}

	redeemCode.IsUsed = true
	redeemCode.UsedBy = &user.ID
	user.IsGatewayActive = true

	if err := koneksi.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&redeemCode).Error; err != nil {
			return err
		}
		return tx.Save(user).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "AI Gateway activated successfully", "user": sanitizeUser(user)})
}

func queryScopeForUser(query *gorm.DB, user *models.User) *gorm.DB {
	switch user.Role {
	case models.RoleStudent:
		return query.Joins("JOIN students ON students.id = consultation_logs.student_id").Where("students.user_id = ?", user.ID)
	case models.RoleLecturer:
		return query.Joins("JOIN students ON students.id = consultation_logs.student_id").Where("students.lecturer_id = ?", user.Lecturer.ID)
	default:
		return query
	}
}

func DashboardStatsV2(c *gin.Context) {
	user := middleware.CurrentUser(c)

	var totalLogs int64
	var totalFeedback int64
	var majorFeedback int64
	var pendingFeedback int64
	var quests []models.FeedbackItem

	logQuery := queryScopeForUser(koneksi.DB.Model(&models.ConsultationLog{}), user)
	feedbackQuery := queryScopeForUser(koneksi.DB.Model(&models.FeedbackItem{}).Joins("JOIN consultation_logs ON consultation_logs.id = feedback_items.log_id"), user)

	logQuery.Count(&totalLogs)
	feedbackQuery.Count(&totalFeedback)
	feedbackQuery.Session(&gorm.Session{}).Where("feedback_items.category = ?", models.CategoryMajor).Count(&majorFeedback)
	feedbackQuery.Session(&gorm.Session{}).Where("feedback_items.status = ?", models.StatusPending).Count(&pendingFeedback)
	feedbackQuery.Session(&gorm.Session{}).Where("feedback_items.status != ?", models.StatusValidated).Order("feedback_items.created_at desc").Limit(5).Find(&quests)

	completionRate := 0
	if totalFeedback > 0 {
		completionRate = int(((totalFeedback - pendingFeedback) * 100) / totalFeedback)
	}

	response := gin.H{
		"total_consultations": totalLogs,
		"total_feedback":      totalFeedback,
		"pending_feedback":    pendingFeedback,
		"major_feedback":      majorFeedback,
		"completion_rate":     completionRate,
		"draft_count":         totalLogs,
		"upcoming_quests":     quests,
	}

	if user.Role == models.RoleStudent && user.Student != nil {
		var lecturer models.Lecturer
		if err := koneksi.DB.First(&lecturer, user.Student.LecturerID).Error; err == nil {
			response["lecturer_name"] = lecturer.Name
		}
	}
	if user.Role == models.RoleLecturer && user.Lecturer != nil {
		var studentCount int64
		koneksi.DB.Model(&models.Student{}).Where("lecturer_id = ?", user.Lecturer.ID).Count(&studentCount)
		response["student_count"] = studentCount
		response["validation_queue"] = pendingFeedback
	}

	c.JSON(http.StatusOK, response)
}

func accessibleLog(user *models.User, logID uint64) (*models.ConsultationLog, error) {
	var log models.ConsultationLog
	query := koneksi.DB.Preload("FeedbackItems").Preload("FeedbackItems.Comments").Preload("Student").Preload("Student.User").Preload("Student.Lecturer")

	switch user.Role {
	case models.RoleStudent:
		query = query.Joins("JOIN students ON students.id = consultation_logs.student_id").Where("consultation_logs.id = ? AND students.user_id = ?", logID, user.ID)
	case models.RoleLecturer:
		query = query.Joins("JOIN students ON students.id = consultation_logs.student_id").Where("consultation_logs.id = ? AND students.lecturer_id = ?", logID, user.Lecturer.ID)
	default:
		return nil, errors.New("unsupported role")
	}

	if err := query.First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func ConsultationListV2(c *gin.Context) {
	user := middleware.CurrentUser(c)
	var logs []models.ConsultationLog

	query := queryScopeForUser(
		koneksi.DB.
			Preload("FeedbackItems").
			Preload("FeedbackItems.Comments").
			Preload("RevisionAnnotations").
			Preload("Student").
			Preload("Student.User").
			Preload("Student.Lecturer"),
		user,
	)
	if err := query.Order("consultation_logs.created_at desc").Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func ArchiveListV2(c *gin.Context) {
	ConsultationListV2(c)
}

const (
	maxAudioUploadSize = 50 << 20  // 50 MB
	maxPaperUploadSize = 20 << 20  // 20 MB
	maxAnnotationSize  = 10 << 20  // 10 MB
	maxInputNameLength = 255
	maxInputEmailLength = 254
	maxInputContentLength = 5000
)

var (
	allowedAudioExts = map[string]bool{".mp3": true, ".wav": true, ".m4a": true, ".ogg": true, ".webm": true, ".mp4": true}
	allowedPaperExts = map[string]bool{".docx": true}
	allowedImageExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
	allowedAnnotExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true, ".docx": true}
	safeFilenameRe   = regexp.MustCompile(`[^a-zA-Z0-9_\-.]`)
)

func sanitizeFilename(name string) string {
	clean := safeFilenameRe.ReplaceAllString(name, "_")
	clean = filepath.Base(clean)
	if clean == "." || clean == ".." || clean == "" {
		clean = "unnamed"
	}
	if len(clean) > 100 {
		clean = clean[len(clean)-100:]
	}
	return clean
}

func validateFileUpload(file *multipart.FileHeader, maxSizes map[string]int64, allowedExts map[string]bool) error {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		return fmt.Errorf("file type '%s' is not allowed", ext)
	}
	if maxSize, ok := maxSizes[ext]; ok && file.Size > maxSize {
		return fmt.Errorf("file size %d exceeds maximum %d bytes", file.Size, maxSize)
	}
	return nil
}

func CreateConsultationV2(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user.Role != models.RoleStudent {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only students can create consultations"})
		return
	}

	var student models.Student
	if err := koneksi.DB.Where("user_id = ?", user.ID).First(&student).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student profile not found"})
		return
	}

	// ── Paper is ALWAYS required ──
	paperFile, err := c.FormFile("paper")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paper file (.docx) is required"})
		return
	}
	if paperFile.Size > maxPaperUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paper file exceeds 20 MB limit"})
		return
	}

	ext := strings.ToLower(filepath.Ext(paperFile.Filename))
	if !allowedPaperExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .docx files are accepted"})
		return
	}

	// ── Audio is OPTIONAL ──
	audioFile, _ := c.FormFile("audio") // err intentionally ignored
	if audioFile != nil && audioFile.Size > maxAudioUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio file exceeds 50 MB limit"})
		return
	}

	// ── Annotations are OPTIONAL ──
	var annotationFiles []*multipart.FileHeader
	if form, formErr := c.MultipartForm(); formErr == nil {
		annotationFiles = form.File["annotations"]
	}

	// ── Validation: at least one supplementary input required ──
	if audioFile == nil && len(annotationFiles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Please provide at least one: audio recording OR annotation images/notes",
		})
		return
	}

	timestamp := time.Now().UnixNano()

	// Save audio if provided
	var audioFilename string
	var audioPath string
	if audioFile != nil {
		audioFilename = fmt.Sprintf("%d_%s", timestamp, sanitizeFilename(audioFile.Filename))
		audioPath = filepath.Join("storage", "audio", audioFilename)
		if err := c.SaveUploadedFile(audioFile, audioPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save audio file"})
			return
		}
	}

	paperFilename := fmt.Sprintf("%d_%s", timestamp, sanitizeFilename(paperFile.Filename))
	paperPath := filepath.Join("storage", "paper", paperFilename)
	if err := c.SaveUploadedFile(paperFile, paperPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save paper file"})
		return
	}

	paperText, err := utils.ReadDocxText(paperPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract text from docx: " + err.Error()})
		return
	}

	var prevLog models.ConsultationLog
	var prevFeedbackStr string
	if err := koneksi.DB.Preload("FeedbackItems").Where("student_id = ?", student.ID).Order("created_at desc").First(&prevLog).Error; err == nil {
		var feedbackLines []string
		for _, item := range prevLog.FeedbackItems {
			feedbackLines = append(feedbackLines, fmt.Sprintf("- [%s] %s", item.Category, item.Content))
		}
		prevFeedbackStr = strings.Join(feedbackLines, "\n")
	}

	type annotationResult struct {
		filename     string
		fileType     models.AnnotationFileType
		extractedText string
	}

	var annotationResults []annotationResult
	var annotationSummary string
	if len(annotationFiles) > 0 {
		for i, fh := range annotationFiles {
			if fh.Size > maxAnnotationSize {
				continue
			}
			annExt := strings.ToLower(filepath.Ext(fh.Filename))
			if !allowedAnnotExts[annExt] {
				continue
			}
			filename := fmt.Sprintf("%d_annotation_%d%s", timestamp, i+1, annExt)
			savePath := filepath.Join("storage", "annotations", filename)
			if err := c.SaveUploadedFile(fh, savePath); err != nil {
				continue
			}
			var extractedText string
			var fileType models.AnnotationFileType
			if allowedImageExts[annExt] {
				fileType = models.AnnotationImage
				extractedText, _ = processAnnotationImage(savePath, user)
			} else if annExt == ".docx" {
				fileType = models.AnnotationDocx
				extractedText, _ = utils.ExtractDocxTrackChanges(savePath)
			} else {
				continue
			}
			annotationResults = append(annotationResults, annotationResult{
				filename:      filename,
				fileType:      fileType,
				extractedText: extractedText,
			})
			label := fmt.Sprintf("[Anotasi %d — %s]", i+1, fh.Filename)
			annotationSummary = annotationSummary + label + "\n" + extractedText + "\n\n---\n\n"
		}
		annotationSummary = strings.TrimRight(annotationSummary, "\n\n---\n\n")
		if annotationSummary != "" {
			prevFeedbackStr = prevFeedbackStr + "\n\nANOTASI REVISI DOSEN:\n" + annotationSummary
		}
	}

	feedbackItems, transcriptContent, err := AnalyzeAudioAndPaper(user.ID, audioPath, paperText, prevFeedbackStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI Processing failed: " + err.Error()})
		return
	}

	// Only write transcript file if there's content
	var transcriptFilename string
	if transcriptContent != "" {
		transcriptFilename = fmt.Sprintf("%d_transcript.txt", timestamp)
		transcriptPath := filepath.Join("storage", "transcript", transcriptFilename)
		if writeErr := os.WriteFile(transcriptPath, []byte(transcriptContent), 0644); writeErr != nil {
			fmt.Printf("[WARN] Failed to write transcript file: %v\n", writeErr)
		}
	}

	log := models.ConsultationLog{
		StudentID:          student.ID,
		AudioFilename:      audioFilename,
		TranscriptFilename: transcriptFilename,
		TranscriptText:     transcriptContent,
		PaperFilename:      paperFilename,
		FeedbackItems:      feedbackItems,
	}

	if err := koneksi.DB.Create(&log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	for _, ann := range annotationResults {
		record := models.RevisionAnnotation{
			ConsultationLogID: log.ID,
			Filename:          ann.filename,
			FileType:          ann.fileType,
			ExtractedText:     ann.extractedText,
		}
		if err := koneksi.DB.Create(&record).Error; err != nil {
			fmt.Printf("[WARN] Failed to save annotation record: %v\n", err)
		}
	}

	log.Student = &student
	c.JSON(http.StatusCreated, gin.H{"message": "Consultation created successfully", "data": log})
}

func ConsultationChatV2(c *gin.Context) {
	user := middleware.CurrentUser(c)
	var req struct {
		LogID uint64 `json:"log_id" binding:"required"`
		Query string `json:"query" binding:"required"`
		Model string `json:"model"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log, err := accessibleLog(user, req.LogID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied or consultation log not found"})
		return
	}

	// 1. Save user query to database
	userMsg := models.AIChatMessage{
		LogID:   log.ID,
		Role:    "user",
		Content: req.Query,
	}
	koneksi.DB.Create(&userMsg)

	if WebSocketHub != nil {
		WebSocketHub.Broadcast("consultation."+strconv.FormatUint(req.LogID, 10), "chat.message", gin.H{
			"id":         userMsg.ID,
			"log_id":     req.LogID,
			"role":       "user",
			"content":    req.Query,
			"created_at": userMsg.CreatedAt,
		})
	}

	response, err := GenerateRevisionAssistance(req.LogID, req.Query, req.Model)
	if err != nil {
		if strings.HasPrefix(err.Error(), "GUARDED:") {
			c.JSON(http.StatusForbidden, gin.H{"status": "guarded", "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. Save AI response to database
	aiMsg := models.AIChatMessage{
		LogID:   log.ID,
		Role:    "ai",
		Content: response,
	}
	koneksi.DB.Create(&aiMsg)

	if WebSocketHub != nil {
		WebSocketHub.Broadcast("consultation."+strconv.FormatUint(req.LogID, 10), "chat.message", gin.H{
			"id":         aiMsg.ID,
			"log_id":     req.LogID,
			"role":       "ai",
			"content":    response,
			"created_at": aiMsg.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "ai_response": response})
}

// GetAIChats fetches all persistent AI chats for a given log ID.
func GetAIChats(c *gin.Context) {
	user := middleware.CurrentUser(c)
	logIDStr := c.Param("id")
	logID, err := strconv.ParseUint(logIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	log, err := accessibleLog(user, logID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var messages []models.AIChatMessage
	if err := koneksi.DB.Where("log_id = ?", log.ID).Order("created_at asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}

func UpdateFeedbackStatusV2(c *gin.Context) {
	user := middleware.CurrentUser(c)
	id := c.Param("id")

	var req struct {
		Status  string `json:"status" binding:"required"`
		LogID   uint64 `json:"log_id"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if user.Role == models.RoleStudent {
		if req.Status != string(models.StatusFixed) && req.Status != string(models.StatusPending) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Students can only change status to Pending or Fixed"})
			return
		}
	} else if user.Role == models.RoleLecturer {
		if req.Status != string(models.StatusValidated) && req.Status != string(models.StatusPending) && req.Status != string(models.StatusRejected) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Lecturers can validate (Validated), return to Pending, or reject (Rejected)"})
			return
		}
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unknown role"})
		return
	}

	var feedback models.FeedbackItem
	if err := koneksi.DB.First(&feedback, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feedback item not found"})
		return
	}

	log, err := accessibleLog(user, feedback.ConsultationLogID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	feedback.Status = models.FeedbackStatus(req.Status)
	if err := koneksi.DB.Save(&feedback).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Status == string(models.StatusRejected) && req.Comment != "" {
		comment := models.FeedbackComment{
			FeedbackItemID: feedback.ID,
			SenderID:       user.ID,
			SenderRole:     string(user.Role),
			Content:        req.Comment,
		}
		if err := koneksi.DB.Create(&comment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Broadcast comment via WebSocket
		if WebSocketHub != nil {
			WebSocketHub.Broadcast("consultation."+strconv.FormatUint(log.ID, 10), "feedback.comment.new", gin.H{
				"id":               comment.ID,
				"feedback_item_id": comment.FeedbackItemID,
				"sender_id":        comment.SenderID,
				"sender_role":      comment.SenderRole,
				"content":          comment.Content,
				"created_at":       comment.CreatedAt,
			})
		}
	}

	payload := gin.H{
		"feedback_id":         feedback.ID,
		"log_id":              feedback.ConsultationLogID,
		"consultation_log_id": feedback.ConsultationLogID,
		"status":              feedback.Status,
		"updated_by_role":     user.Role,
	}

	if WebSocketHub != nil {
		WebSocketHub.Broadcast("consultation."+strconv.FormatUint(log.ID, 10), "feedback.status-updated", payload)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feedback status updated successfully", "data": payload})
}

func LecturerConsultationsV2(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user.Role != models.RoleLecturer || user.Lecturer == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only lecturers can access this data"})
		return
	}

	var logs []models.ConsultationLog
	if err := koneksi.DB.Preload("FeedbackItems").Preload("FeedbackItems.Comments").Preload("Student").Preload("Student.User").Preload("RevisionAnnotations").
		Joins("JOIN students ON students.id = consultation_logs.student_id").
		Where("students.lecturer_id = ?", user.Lecturer.ID).
		Order("consultation_logs.created_at desc").
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func LecturerStudentsV2(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user.Role != models.RoleLecturer || user.Lecturer == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only lecturers can access this data"})
		return
	}

	var students []models.Student
	if err := koneksi.DB.Preload("User").Where("lecturer_id = ?", user.Lecturer.ID).Find(&students).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": students})
}

// LecturerAddFeedbackV2 allows a lecturer to manually add a feedback item
// to a specific consultation log that belongs to one of their supervised students.
func LecturerAddFeedbackV2(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user.Role != models.RoleLecturer || user.Lecturer == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only lecturers can dispatch feedback"})
		return
	}

	logIDStr := c.Param("id")
	logID, err := strconv.ParseUint(logIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	var req struct {
		Content  string `json:"content" binding:"required"`
		Category string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the log belongs to a student supervised by this lecturer
	log, err := accessibleLog(user, logID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied or consultation log not found"})
		return
	}

	// Default manually added feedback to "Major" (HOC) initially.
	// The student will run their AI Oracle to classify it properly.
	category := string(models.CategoryMajor)

	feedback := models.FeedbackItem{
		ConsultationLogID: log.ID,
		Content:           req.Content,
		Category:          models.FeedbackCategory(category),
		Status:            models.StatusPending,
	}
	if err := koneksi.DB.Create(&feedback).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast to real-time subscribers — use dedicated "feedback.new" event
	// so clients can APPEND the item instead of trying to mutate a non-existent one.
	if WebSocketHub != nil {
		WebSocketHub.Broadcast("consultation."+strconv.FormatUint(log.ID, 10), "feedback.new", gin.H{
			"id":                   feedback.ID,
			"feedback_id":          feedback.ID,
			"log_id":               feedback.ConsultationLogID,
			"consultation_log_id":  feedback.ConsultationLogID,
			"content":              feedback.Content,
			"status":               feedback.Status,
			"category":             feedback.Category,
			"created_at":           feedback.CreatedAt,
			"updated_by_role":      user.Role,
		})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Feedback dispatched successfully", "data": feedback})
}

// GetDirectMessages fetches all direct messages for a given consultation log ID.
func GetDirectMessages(c *gin.Context) {
	user := middleware.CurrentUser(c)
	logIDStr := c.Param("id")
	logID, err := strconv.ParseUint(logIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	// Verify accessibility (log belongs to student or supervisor)
	log, err := accessibleLog(user, logID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var messages []models.DirectMessage
	if err := koneksi.DB.Where("log_id = ?", log.ID).Order("created_at asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}

// SendDirectMessage saves a new direct message to the database and broadcasts it via WebSocket.
func SendDirectMessage(c *gin.Context) {
	user := middleware.CurrentUser(c)
	logIDStr := c.Param("id")
	logID, err := strconv.ParseUint(logIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	// Verify accessibility
	log, err := accessibleLog(user, logID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := models.DirectMessage{
		LogID:      log.ID,
		SenderID:   user.ID,
		SenderRole: string(user.Role),
		Content:    req.Content,
	}

	if err := koneksi.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast via WebSocket
	if WebSocketHub != nil {
		WebSocketHub.Broadcast("consultation."+strconv.FormatUint(log.ID, 10), "chat.direct-message", gin.H{
			"id":          msg.ID,
			"log_id":      msg.LogID,
			"sender_id":   msg.SenderID,
			"sender_role": msg.SenderRole,
			"content":     msg.Content,
			"created_at":  msg.CreatedAt,
		})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Message sent successfully", "data": msg})
}

// extractJSONBounds returns the tightest JSON substring found in input.
// It prefers a leading '{' object, otherwise a leading '[' array.
func extractJSONBounds(input string) string {
	input = strings.TrimSpace(input)
	// Strip common markdown code fences
	for _, fence := range []string{"```json", "```JSON", "```"} {
		if strings.HasPrefix(input, fence) {
			input = strings.TrimPrefix(input, fence)
			if idx := strings.LastIndex(input, "```"); idx != -1 {
				input = input[:idx]
			}
			input = strings.TrimSpace(input)
			break
		}
	}

	firstBrace := strings.Index(input, "{")
	firstBracket := strings.Index(input, "[")

	start := -1
	var closing string
	if firstBrace != -1 && (firstBracket == -1 || firstBrace < firstBracket) {
		start = firstBrace
		closing = "}"
	} else if firstBracket != -1 {
		start = firstBracket
		closing = "]"
	}

	if start == -1 {
		return input
	}

	end := strings.LastIndex(input, closing)
	if end == -1 || end < start {
		return input
	}
	return input[start : end+1]
}

// Kept for backward compatibility — some call sites still use the old name.
func extractJSONString(input string) string { return extractJSONBounds(input) }

func parseClassificationResponse(aiResponse string, cleanedResponse string, log *models.ConsultationLog) []ClassificationItem {
	var result []ClassificationItem

	var wrapper struct {
		Items           []ClassificationItem `json:"items"`
		Classifications []ClassificationItem `json:"classifications"`
		Feedbacks       []ClassificationItem `json:"feedbacks"`
		Data            []ClassificationItem `json:"data"`
		Results         []ClassificationItem `json:"results"`
	}
	if err := json.Unmarshal([]byte(cleanedResponse), &wrapper); err == nil {
		switch {
		case len(wrapper.Items) > 0:
			return wrapper.Items
		case len(wrapper.Classifications) > 0:
			return wrapper.Classifications
		case len(wrapper.Feedbacks) > 0:
			return wrapper.Feedbacks
		case len(wrapper.Data) > 0:
			return wrapper.Data
		case len(wrapper.Results) > 0:
			return wrapper.Results
		}
	}

	if err := json.Unmarshal([]byte(cleanedResponse), &result); err == nil && len(result) > 0 {
		return result
	}

	var flatMap map[string]string
	if err := json.Unmarshal([]byte(cleanedResponse), &flatMap); err == nil && len(flatMap) > 0 {
		for k, v := range flatMap {
			if id, parseErr := strconv.ParseUint(k, 10, 64); parseErr == nil {
				result = append(result, ClassificationItem{ID: id, Category: v})
			}
		}
		if len(result) > 0 {
			return result
		}
	}

	for _, fb := range log.FeedbackItems {
		idStr := strconv.FormatUint(fb.ID, 10)
		idMarker := `"id":` + idStr
		if idx := strings.Index(aiResponse, idMarker); idx != -1 {
			chunk := aiResponse[idx:]
			if len(chunk) > 80 {
				chunk = chunk[:80]
			}
			chunk = strings.ToLower(chunk)
			cat := "Minor"
			if strings.Contains(chunk, "major") {
				cat = "Major"
			}
			result = append(result, ClassificationItem{ID: fb.ID, Category: cat})
		}
	}

	return result
}

func ClassifyFeedbackV2(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user.Role != models.RoleStudent {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only students can initiate AI classification"})
		return
	}

	logIDStr := c.Param("id")
	logID, err := strconv.ParseUint(logIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	log, err := accessibleLog(user, logID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if len(log.FeedbackItems) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No feedback items to classify", "data": log.FeedbackItems})
		return
	}

	type classificationInput struct {
		ID      uint64 `json:"id"`
		Content string `json:"content"`
	}
	var inputItems []classificationInput
	for _, item := range log.FeedbackItems {
		inputItems = append(inputItems, classificationInput{ID: item.ID, Content: item.Content})
	}
	itemsData, _ := json.Marshal(inputItems)

	systemPrompt := `You are an expert academic writing advisor.

Your task: classify each feedback item below as either "Major" or "Minor".
- "Major" (HOC – Higher Order Concerns): core substance — research structure, arguments, methodology, analysis, research model, thesis title.
- "Minor" (LOC – Lower Order Concerns): surface-level — formatting, typos, citation style, bibliography, spacing, spelling, grammar.

Return ONLY valid JSON in this exact shape — no explanation, no markdown, no extra text:
{"items":[{"id":1,"category":"Major"},{"id":2,"category":"Minor"}]}`

	userPrompt := fmt.Sprintf("Classify the following feedback items:\n%s", string(itemsData))

	aiResponse, err := callAI(user, systemPrompt, userPrompt, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI classification failed: " + err.Error()})
		return
	}

	// Sanitize JSON before parsing
	cleanedResponse := sanitizeJSON(aiResponse)
	cleanedResponse = extractJSONBounds(cleanedResponse)
	cleanedResponse = strings.TrimSpace(cleanedResponse)

	finalClassifications := parseClassificationResponse(aiResponse, cleanedResponse, log)

	if len(finalClassifications) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse AI classification results. Please try sorting again.",
		})
		return
	}

	for _, cl := range finalClassifications {
		cat := cl.Category
		if strings.ToLower(cat) == "major" {
			cat = "Major"
		} else if strings.ToLower(cat) == "minor" {
			cat = "Minor"
		}
		if cat == "Major" || cat == "Minor" {
			koneksi.DB.Model(&models.FeedbackItem{}).Where("id = ? AND log_id = ?", cl.ID, log.ID).Update("category", cat)
		}
	}

	var updatedItems []models.FeedbackItem
	koneksi.DB.Where("log_id = ?", log.ID).Find(&updatedItems)

	if WebSocketHub != nil {
		for _, item := range updatedItems {
			WebSocketHub.Broadcast("consultation."+strconv.FormatUint(log.ID, 10), "feedback.status-updated", gin.H{
				"feedback_id":         item.ID,
				"log_id":              item.ConsultationLogID,
				"consultation_log_id": item.ConsultationLogID,
				"status":              item.Status,
				"category":            item.Category,
				"updated_by_role":     user.Role,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feedback items classified successfully", "data": updatedItems})
}

// ─────────────────────────────────────────────────────────────────────────────
//  SESSION DELETION & DISK CLEANUP
// ─────────────────────────────────────────────────────────────────────────────

// DeleteConsultationV2 removes a consultation session, its files, and all related records.
// Authorization: Students can only delete their own sessions.
//
//	Lecturers can only delete sessions of their supervised students.
func DeleteConsultationV2(c *gin.Context) {
	user := middleware.CurrentUser(c)
	logIDStr := c.Param("id")
	logID, err := strconv.ParseUint(logIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	// Verify access (same logic as accessibleLog)
	log, err := accessibleLog(user, logID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied or session not found"})
		return
	}

	// ── Step 1: Delete physical files from disk ──
	filesToDelete := []string{}

	if log.AudioFilename != "" {
		filesToDelete = append(filesToDelete, filepath.Join("storage", "audio", log.AudioFilename))
	}
	if log.PaperFilename != "" {
		filesToDelete = append(filesToDelete, filepath.Join("storage", "paper", log.PaperFilename))
	}
	if log.TranscriptFilename != "" {
		filesToDelete = append(filesToDelete, filepath.Join("storage", "transcript", log.TranscriptFilename))
	}

	// Load annotations to delete their files
	var annotations []models.RevisionAnnotation
	koneksi.DB.Where("log_id = ?", log.ID).Find(&annotations)
	for _, ann := range annotations {
		filesToDelete = append(filesToDelete, filepath.Join("storage", "annotations", ann.Filename))
	}

	for _, filePath := range filesToDelete {
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("[DELETE] Warning: failed to remove file %s: %v\n", filePath, err)
		} else {
			fmt.Printf("[DELETE] Removed file: %s\n", filePath)
		}
	}

	// ── Step 2: Delete database records in transaction ──
	if err := koneksi.DB.Transaction(func(tx *gorm.DB) error {
		// Delete AI chat messages
		if err := tx.Where("log_id = ?", log.ID).Delete(&models.AIChatMessage{}).Error; err != nil {
			return err
		}
		// Delete direct messages
		if err := tx.Where("log_id = ?", log.ID).Delete(&models.DirectMessage{}).Error; err != nil {
			return err
		}
		// Delete feedback items (cascades via GORM constraint, but explicit for safety)
		if err := tx.Where("log_id = ?", log.ID).Delete(&models.FeedbackItem{}).Error; err != nil {
			return err
		}
		// Delete revision annotations
		if err := tx.Where("log_id = ?", log.ID).Delete(&models.RevisionAnnotation{}).Error; err != nil {
			return err
		}
		// Delete the consultation log itself
		if err := tx.Delete(log).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session: " + err.Error()})
		return
	}

	// ── Step 3: Broadcast deletion event ──
	if WebSocketHub != nil {
		WebSocketHub.Broadcast("consultation."+strconv.FormatUint(log.ID, 10), "session.deleted", gin.H{
			"log_id":     log.ID,
			"deleted_by": user.ID,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session deleted successfully"})
}

// ─────────────────────────────────────────────────────────────────────────────
//  FEEDBACK COMMENTS (Threaded Discussion)
// ─────────────────────────────────────────────────────────────────────────────

// GetFeedbackComments returns all comments for a given feedback item.
func GetFeedbackComments(c *gin.Context) {
	user := middleware.CurrentUser(c)
	feedbackIDStr := c.Param("id")
	feedbackID, err := strconv.ParseUint(feedbackIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feedback ID"})
		return
	}

	// Verify the feedback item exists and is accessible
	var feedback models.FeedbackItem
	if err := koneksi.DB.First(&feedback, feedbackID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feedback item not found"})
		return
	}

	if _, err := accessibleLog(user, feedback.ConsultationLogID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var comments []models.FeedbackComment
	if err := koneksi.DB.Where("feedback_item_id = ?", feedbackID).
		Order("created_at asc").
		Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": comments})
}

// AddFeedbackComment creates a new comment on a feedback item and broadcasts via WebSocket.
func AddFeedbackComment(c *gin.Context) {
	user := middleware.CurrentUser(c)
	feedbackIDStr := c.Param("id")
	feedbackID, err := strconv.ParseUint(feedbackIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feedback ID"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the feedback item exists and is accessible
	var feedback models.FeedbackItem
	if err := koneksi.DB.First(&feedback, feedbackID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feedback item not found"})
		return
	}

	log, err := accessibleLog(user, feedback.ConsultationLogID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	comment := models.FeedbackComment{
		FeedbackItemID: feedbackID,
		SenderID:       user.ID,
		SenderRole:     string(user.Role),
		Content:        req.Content,
	}

	if err := koneksi.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast comment via WebSocket
	if WebSocketHub != nil {
		WebSocketHub.Broadcast("consultation."+strconv.FormatUint(log.ID, 10), "feedback.comment.new", gin.H{
			"id":              comment.ID,
			"feedback_item_id": comment.FeedbackItemID,
			"sender_id":       comment.SenderID,
			"sender_role":     comment.SenderRole,
			"content":         comment.Content,
			"created_at":      comment.CreatedAt,
		})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Comment added", "data": comment})
}

// ─────────────────────────────────────────────────────────────────────────────
//  SESSION-BASED FILTERING
// ─────────────────────────────────────────────────────────────────────────────

// LecturerStudentSessions returns session metadata for a specific student,
// enabling the lecturer dashboard to filter revisions per session.
func LecturerStudentSessions(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user.Role != models.RoleLecturer || user.Lecturer == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only lecturers can access this"})
		return
	}

	studentIDStr := c.Param("id")
	studentID, err := strconv.ParseUint(studentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	// Verify student belongs to this lecturer
	var student models.Student
	if err := koneksi.DB.Where("id = ? AND lecturer_id = ?", studentID, user.Lecturer.ID).First(&student).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Student not in your supervision"})
		return
	}

	var logs []models.ConsultationLog
	if err := koneksi.DB.Preload("FeedbackItems").Preload("FeedbackItems.Comments").
		Where("student_id = ?", student.ID).
		Order("created_at asc").
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type SessionInfo struct {
		SessionNumber int    `json:"session_number"`
		LogID         uint64 `json:"log_id"`
		PaperFilename string `json:"paper_filename"`
		CreatedAt     string `json:"created_at"`
		FeedbackCount int    `json:"feedback_count"`
		PendingCount  int    `json:"pending_count"`
	}

	var sessions []SessionInfo
	for i, log := range logs {
		pending := 0
		for _, f := range log.FeedbackItems {
			if f.Status == models.StatusPending {
				pending++
			}
		}
		sessions = append(sessions, SessionInfo{
			SessionNumber: i + 1,
			LogID:         log.ID,
			PaperFilename: log.PaperFilename,
			CreatedAt:     log.CreatedAt.Format(time.RFC3339),
			FeedbackCount: len(log.FeedbackItems),
			PendingCount:  pending,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

// GetConsultationDraftsV2 returns all docx revision annotations (draft versions) for a consultation log
func GetConsultationDraftsV2(c *gin.Context) {
	user := middleware.CurrentUser(c)
	logIDStr := c.Param("id")
	logID, err := strconv.ParseUint(logIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consultation log ID"})
		return
	}

	log, err := accessibleLog(user, logID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	type DraftItem struct {
		ID                uint64    `json:"id"`
		ConsultationLogID uint64    `json:"consultation_log_id"`
		Version           int       `json:"version"`
		Filename          string    `json:"filename"`
		CreatedAt         time.Time `json:"created_at"`
		Notes             string    `json:"notes"`
	}

	var annotations []models.RevisionAnnotation
	if err := koneksi.DB.Where("log_id = ? AND file_type = 'docx'", log.ID).Order("created_at asc").Find(&annotations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	drafts := make([]DraftItem, 0, len(annotations))
	for i, ann := range annotations {
		drafts = append(drafts, DraftItem{
			ID:                ann.ID,
			ConsultationLogID: ann.ConsultationLogID,
			Version:           i + 1,
			Filename:          ann.Filename,
			CreatedAt:         ann.CreatedAt,
			Notes:             "Revised draft version",
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": drafts})
}

// UploadFinalDraftV2 allows a student to upload their final/revised docx file which gets stored as a docx revision annotation
func UploadFinalDraftV2(c *gin.Context) {
	user := middleware.CurrentUser(c)
	logIDStr := c.Param("id")
	logID, err := strconv.ParseUint(logIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consultation log ID"})
		return
	}

	log, err := accessibleLog(user, logID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Docx file is required"})
		return
	}

	if file.Size > maxPaperUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File exceeds maximum size limits (20 MB)"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".docx" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .docx files are accepted as final drafts"})
		return
	}

	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%d_final_%s", timestamp, sanitizeFilename(file.Filename))
	savePath := filepath.Join("storage", "annotations", filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file on server"})
		return
	}

	// Read docx text (optional, for search/indexing/OCR compatibility)
	extractedText, _ := utils.ExtractDocxTrackChanges(savePath)

	record := models.RevisionAnnotation{
		ConsultationLogID: log.ID,
		Filename:          filename,
		FileType:          models.AnnotationDocx,
		ExtractedText:     extractedText,
	}

	if err := koneksi.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save draft record: " + err.Error()})
		return
	}

	// Broadcast update event
	if WebSocketHub != nil {
		WebSocketHub.Broadcast("consultation."+strconv.FormatUint(log.ID, 10), "feedback.new", gin.H{
			"log_id": log.ID,
		})
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Final document uploaded successfully", "data": record})
}
