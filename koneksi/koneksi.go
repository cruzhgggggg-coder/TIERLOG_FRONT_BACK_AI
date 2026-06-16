package koneksi

import (
	"fmt"
	"os"
	"time"

	"testing_go/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func ConnectDatabase() {
	user := envOrDefault("DB_USERNAME", "root")
	password := envOrDefault("DB_PASSWORD", "")
	host := envOrDefault("DB_HOST", "127.0.0.1")
	port := envOrDefault("DB_PORT", "3306")
	dbname := envOrDefault("DB_DATABASE", "struct_go")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbname)
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("An error occurred while connecting to the database: " + err.Error())
	}

	if err := database.AutoMigrate(
		&models.User{},
		&models.Lecturer{},
		&models.Student{},
		&models.ConsultationLog{},
		&models.FeedbackItem{},
		&models.FeedbackComment{},
		&models.RevisionAnnotation{},
		&models.RedeemCode{},
		&models.RefreshToken{},
		&models.DirectMessage{},
		&models.AIChatMessage{},
	); err != nil {
		panic("An error occurred during database migration: " + err.Error())
	}

	migrator := database.Migrator()
	if migrator.HasConstraint(&models.FeedbackItem{}, "feedback_items_consultation_log_id_foreign") {
		_ = migrator.DropConstraint(&models.FeedbackItem{}, "feedback_items_consultation_log_id_foreign")
	}
	if migrator.HasConstraint(&models.FeedbackItem{}, "fk_consultation_logs_feedback_items") {
		_ = migrator.DropConstraint(&models.FeedbackItem{}, "fk_consultation_logs_feedback_items")
	}
	if migrator.HasColumn(&models.FeedbackItem{}, "consultation_log_id") {
		_ = migrator.DropColumn(&models.FeedbackItem{}, "consultation_log_id")
	}

	_ = database.Exec("ALTER TABLE feedback_items MODIFY COLUMN status ENUM('Fixed', 'Pending', 'Validated', 'Rejected') NOT NULL DEFAULT 'Pending'").Error

	indexes := []struct {
		table string
		cols  string
		name  string
	}{
		{"feedback_items", "log_id", "idx_feedback_log_id"},
		{"feedback_items", "log_id,status", "idx_feedback_log_status"},
		{"feedback_items", "log_id,category", "idx_feedback_log_category"},
		{"consultation_logs", "student_id", "idx_consultation_student_id"},
		{"consultation_logs", "student_id,created_at", "idx_consultation_student_created"},
		{"direct_messages", "log_id", "idx_dm_log_id"},
		{"direct_messages", "log_id,created_at", "idx_dm_log_created"},
		{"ai_chat_messages", "log_id", "idx_aichat_log_id"},
		{"ai_chat_messages", "log_id,created_at", "idx_aichat_log_created"},
		{"refresh_tokens", "user_id", "idx_refresh_user_id"},
		{"refresh_tokens", "expires_at", "idx_refresh_expires"},
		{"feedback_comments", "feedback_item_id", "idx_fcomment_item_id"},
		{"feedback_comments", "feedback_item_id,created_at", "idx_fcomment_item_created"},
	}

	for _, idx := range indexes {
		sql := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", idx.name, idx.table, idx.cols)
		_ = database.Exec(sql).Error
	}

	sqlDB, err := database.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(15)
		sqlDB.SetMaxOpenConns(150)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		sqlDB.SetConnMaxIdleTime(10 * time.Minute)
	}

	DB = database
	fmt.Println("Database connection successful!")
}
