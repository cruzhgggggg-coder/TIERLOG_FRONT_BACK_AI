package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"testing_go/controller"
	"testing_go/koneksi"
	testing_middleware "testing_go/middleware"
	"testing_go/realtime"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	folders := []string{
		"storage/audio",
		"storage/transcript",
		"storage/paper",
		"storage/annotations",
		"storage/final",
	}

	for _, folder := range folders {
		if err := os.MkdirAll(folder, os.ModePerm); err != nil {
			fmt.Printf("Failed to create directory %s: %v\n", folder, err)
		}
	}
}

func loadEnv() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("tierlog_v2/.env")
}

func getCORSAllowedOrigins() []string {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		return []string{"http://localhost:5173", "http://localhost:8000"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			origins = append(origins, p)
		}
	}
	return origins
}

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, o := range allowedOrigins {
			if o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func requestSizeLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("Request body too large. Maximum size is %d bytes.", maxBytes),
			})
			return
		}
		c.Next()
	}
}

func main() {
	loadEnv()
	koneksi.ConnectDatabase()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.SetTrustedProxies(nil)

	allowedOrigins := getCORSAllowedOrigins()
	r.Use(corsMiddleware(allowedOrigins))
	r.Use(testing_middleware.RequestID())

	generalLimiter := testing_middleware.NewRateLimiter(100, time.Minute)
	r.Use(func(c *gin.Context) {
		key := c.ClientIP()
		if !generalLimiter.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}
		c.Next()
	})

	r.Use(requestSizeLimit(100 << 20))

	hub := realtime.NewHub(hubAllowedOrigins(allowedOrigins))
	controller.SetRealtimeHub(hub)

	r.Static("/storage", "./storage")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().UTC()})
	})

	r.GET("/ws", hub.HandleWebSocket)

	r.POST("/auth/register", testing_middleware.RateLimit(5, time.Minute), controller.Register)
	r.POST("/auth/login", testing_middleware.RateLimit(10, time.Minute), controller.Login)
	r.POST("/auth/refresh", testing_middleware.RateLimit(20, time.Minute), controller.Refresh)
	r.POST("/auth/logout", controller.Logout)
	r.GET("/lecturers", controller.GetLecturers)

	protected := r.Group("/")
	protected.Use(testing_middleware.AuthRequired())
	{
		protected.GET("/auth/me", controller.Me)
		protected.PATCH("/settings/profile", controller.UpdateProfile)
		protected.PUT("/settings/password", controller.UpdatePassword)
		protected.PATCH("/settings/ai-gateway", controller.UpdateAIGatewaySettingsV2)
		protected.POST("/settings/ai-gateway/redeem", controller.RedeemGatewayCodeV2)

		protected.GET("/dashboard/stats", controller.DashboardStatsV2)
		protected.GET("/consultations", controller.ConsultationListV2)
		protected.POST("/consultations", testing_middleware.RoleRequired("student"), controller.CreateConsultationV2)
		protected.POST("/consultations/chat", controller.ConsultationChatV2)
		protected.PUT("/consultations/feedback/:id/status", controller.UpdateFeedbackStatusV2)
		protected.GET("/consultations/feedback/:id/comments", controller.GetFeedbackComments)
		protected.POST("/consultations/feedback/:id/comments", controller.AddFeedbackComment)
		protected.POST("/consultations/:id/add-feedback", testing_middleware.RoleRequired("lecturer"), controller.LecturerAddFeedbackV2)
		protected.GET("/consultations/:id/direct-messages", controller.GetDirectMessages)
		protected.POST("/consultations/:id/direct-messages", controller.SendDirectMessage)
		protected.GET("/consultations/:id/ai-chats", controller.GetAIChats)
		protected.POST("/consultations/:id/classify-feedback", testing_middleware.RoleRequired("student"), controller.ClassifyFeedbackV2)
		protected.DELETE("/consultations/:id", controller.DeleteConsultationV2)
		protected.GET("/consultations/:id/drafts", controller.GetConsultationDraftsV2)
		protected.POST("/consultations/:id/final", testing_middleware.RoleRequired("student"), controller.UploadFinalDraftV2)
		protected.GET("/lecturer/consultations", testing_middleware.RoleRequired("lecturer"), controller.LecturerConsultationsV2)
		protected.GET("/lecturer/students", testing_middleware.RoleRequired("lecturer"), controller.LecturerStudentsV2)
		protected.GET("/lecturer/student/:id/sessions", testing_middleware.RoleRequired("lecturer"), controller.LecturerStudentSessions)
		protected.GET("/ai/models", controller.ListFilteredModels)
		protected.GET("/logs", controller.ArchiveListV2)
	}

	legacyAPI := r.Group("/api")
	legacyAPI.Use(testing_middleware.AuthRequired())
	{
		legacyAPI.POST("/consultation", testing_middleware.RoleRequired("student"), controller.CreateConsultation)
		legacyAPI.GET("/consultation", controller.GetConsultations)
		legacyAPI.GET("/stats", controller.GetStats)
		legacyAPI.POST("/ai/assist", controller.AIAssistHandler)
		legacyAPI.GET("/ai/models", controller.GetAIModels)
		legacyAPI.PUT("/feedback/:id/status", controller.UpdateFeedbackStatus)
		legacyAPI.GET("/lecturer/:id/consultations", controller.GetLecturerConsultations)
		legacyAPI.GET("/lecturer/:id/students", controller.GetLecturerStudents)
		legacyAPI.POST("/settings/ai-keys", controller.UpdateAIGatewaySettings)
		legacyAPI.POST("/settings/redeem", controller.RedeemGatewayCode)
		legacyAPI.POST("/admin/generate-code", testing_middleware.RoleRequired("lecturer"), controller.GenerateRedeemCode)
	}

	adminGroup := r.Group("/admin")
	adminGroup.Use(testing_middleware.AuthRequired(), testing_middleware.RoleRequired("lecturer"))
	{
		adminGroup.POST("/generate-code", controller.GenerateRedeemCode)
	}

	fmt.Println("TierLog unified backend is running at http://localhost:8080")

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	fmt.Println("Server exited gracefully")
}

func hubAllowedOrigins(origins []string) []string {
	return origins
}
