package koneksi

import (
	"fmt"
	"os"
	"testing_go/models"

	"gorm.io/driver/mysql"	
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {

	// Load database configurations from environment variables (loaded by godotenv in main.go)
	user := os.Getenv("DB_USERNAME")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	dbname := os.Getenv("DB_DATABASE")
	if dbname == "" {
		dbname = "struct_go"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbname)
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Terjadi kesalahan saat menghubungkan ke database: " + err.Error())
	}

	database.AutoMigrate(
		&models.User{},
		&models.Lecturer{},
		&models.Student{},
		&models.ConsultationLog{},
		&models.FeedbackItem{},
		&models.RedeemCode{},
	)


	DB = database
	fmt.Println("Koneksi database berhasil!")
}
