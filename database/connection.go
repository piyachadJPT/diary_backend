package database

import (
	"fmt"
	"log"
	"os"

	"gofiber-auth/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	log.Println("Database connected")

	autoMigrate()
}

func autoMigrate() {
	// DB.Migrator().DropTable(&models.Diary{})

	err := DB.AutoMigrate(
		&models.User{},
		&models.StudentAdvisor{},
		&models.Diary{},
		&models.Comment{},
		&models.Attachment{},
		&models.Notification{},
	)

	if err != nil {
		log.Fatalf(" Auto migration failed: %v", err)
	}
}
