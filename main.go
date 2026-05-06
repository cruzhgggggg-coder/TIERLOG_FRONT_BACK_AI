package main

import (
	"fmt"
	"os"

	"testing_go/controller"
	"testing_go/koneksi"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// Ensure external storage directories exist as per instructions
	folders := []string{
		"storage/audio",
		"storage/transcript",
		"storage/paper",
	}

	for _, folder := range folders {
		err := os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			fmt.Printf("Failed to create directory %s: %v\n", folder, err)
		}
	}
}

func main() {
	// Load environment variables from Laravel's .env file
	godotenv.Load("tierlog_v2/.env")

	// Initialize Database (GORM with MySQL)
	koneksi.ConnectDatabase()

	// Initialize AI Provider Mode
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.SetTrustedProxies(nil)

	// --- Middleware: CORS ---
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Serving the external storage files
	r.Static("/storage", "./storage")

	// API Endpoints
	api := r.Group("/api")
	{
		api.POST("/consultation", controller.CreateConsultation)
		api.GET("/consultation", controller.GetConsultations)
		api.GET("/stats", controller.GetStats)
		api.POST("/ai/assist", controller.AIAssistHandler)
		api.PUT("/feedback/:id/status", controller.UpdateFeedbackStatus)
		api.GET("/lecturer/:id/consultations", controller.GetLecturerConsultations)
		api.GET("/lecturer/:id/students", controller.GetLecturerStudents)
	}

	// Identity Management
	r.GET("/users", controller.GetUsers)
	r.POST("/users", controller.CreateUser)
	r.GET("/lecturers", controller.GetLecturers)
	r.POST("/lecturers", controller.CreateLecturer)
	r.GET("/students", controller.GetStudents)
	r.POST("/students", controller.CreateStudent)

	fmt.Println("TierLog API Backend is running at http://localhost:8080")
	r.Run(":8080")
}
