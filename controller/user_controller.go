package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"testing_go/auth"
	"testing_go/koneksi"
	"testing_go/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetUsers(c *gin.Context) {
	var users []models.User
	if err := koneksi.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sanitized := make([]gin.H, 0, len(users))
	for _, u := range users {
		sanitized = append(sanitized, sanitizeUser(&u))
	}
	c.JSON(http.StatusOK, gin.H{"data": sanitized})
}

func CreateUser(c *gin.Context) {
	var req struct {
		Name     string          `json:"name" binding:"required"`
		Email    string          `json:"email" binding:"required,email"`
		Password string          `json:"password" binding:"required,min=8"`
		Role     models.UserRole `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role != models.RoleStudent && req.Role != models.RoleLecturer {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role must be student or lecturer"})
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

	if err := koneksi.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "New user created successfully", "data": sanitizeUser(&user)})
}

func GetLecturers(c *gin.Context) {
	var lecturers []models.Lecturer
	if err := koneksi.DB.Preload("User").Find(&lecturers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": lecturers})
}

func CreateLecturer(c *gin.Context) {
	var req struct {
		UserID   uint64 `json:"user_id" binding:"required"`
		NIP      string `json:"nip" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Faculty  string `json:"faculty"`
		Keahlian string `json:"keahlian"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lecturer := models.Lecturer{
		UserID:   req.UserID,
		NIP:      req.NIP,
		Name:     req.Name,
		Faculty:  req.Faculty,
		Keahlian: req.Keahlian,
	}
	if err := koneksi.DB.Create(&lecturer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Lecturer data added successfully", "data": lecturer})
}

func GetStudents(c *gin.Context) {
	var students []models.Student
	if err := koneksi.DB.Preload("User").Preload("Lecturer").Find(&students).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": students})
}

func CreateStudent(c *gin.Context) {
	var req struct {
		UserID      uint64 `json:"user_id" binding:"required"`
		LecturerID  uint64 `json:"lecturer_id" binding:"required"`
		NIM         string `json:"nim" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Prodi       string `json:"prodi"`
		ThesisTitle string `json:"thesis_title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	student := models.Student{
		UserID:      req.UserID,
		LecturerID:  req.LecturerID,
		NIM:         req.NIM,
		Name:        req.Name,
		Prodi:       req.Prodi,
		ThesisTitle: req.ThesisTitle,
	}
	if err := koneksi.DB.Create(&student).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Student data added successfully", "data": student})
}

func UpdateAIGatewaySettings(c *gin.Context) {
	var req struct {
		UserID         uint64 `json:"user_id" binding:"required"`
		OpenAIKey      string `json:"openai_key"`
		GeminiKey      string `json:"gemini_key"`
		AnthropicKey   string `json:"anthropic_key"`
		NvidiaKey      string `json:"nvidia_key"`
		GroqKey        string `json:"groq_key"`
		PreferredModel string `json:"preferred_model"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	var user models.User
	if err := koneksi.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if req.OpenAIKey != "" {
		user.OpenAIKey = auth.EncryptAPIKey(req.OpenAIKey)
	}
	if req.GeminiKey != "" {
		user.GeminiKey = auth.EncryptAPIKey(req.GeminiKey)
	}
	if req.AnthropicKey != "" {
		user.AnthropicKey = auth.EncryptAPIKey(req.AnthropicKey)
	}
	if req.NvidiaKey != "" {
		user.NvidiaKey = auth.EncryptAPIKey(req.NvidiaKey)
	}
	if req.GroqKey != "" {
		user.GroqKey = auth.EncryptAPIKey(req.GroqKey)
	}
	if req.PreferredModel != "" {
		user.PreferredModel = req.PreferredModel
	}

	if err := koneksi.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save settings: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "AI Gateway settings updated successfully"})
}

func RedeemGatewayCode(c *gin.Context) {
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	var redeemCode models.RedeemCode
	if err := koneksi.DB.Where("code = ? AND is_used = ?", req.Code, false).First(&redeemCode).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Redeem code is invalid or has already been used"})
		return
	}

	var user models.User
	if err := koneksi.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := koneksi.DB.Transaction(func(tx *gorm.DB) error {
		redeemCode.IsUsed = true
		redeemCode.UsedBy = &user.ID
		if err := tx.Save(&redeemCode).Error; err != nil {
			return err
		}
		user.IsGatewayActive = true
		return tx.Save(&user).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "AI Gateway Activated Successfully!"})
}

func GenerateRedeemCode(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code is required"})
		return
	}

	if len(req.Code) < 4 || len(req.Code) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code must be between 4 and 50 characters"})
		return
	}

	newCode := models.RedeemCode{
		Code:      strings.ToUpper(strings.TrimSpace(req.Code)),
		CreatedAt: time.Now(),
	}
	if err := koneksi.DB.Create(&newCode).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create code: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Redeem code created successfully", "code": newCode.Code})
}
