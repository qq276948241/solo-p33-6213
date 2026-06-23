package config

import (
	"equipment-borrow-system/models"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var JWTSecret = []byte("equipment-borrow-secret-key-change-in-production")

func InitDB() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	var err error
	DB, err = gorm.Open(sqlite.Open("equipment.db"), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(&models.User{}, &models.Device{}, &models.BorrowRecord{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	initDefaultAdmin()
}

func initDefaultAdmin() {
	var count int64
	DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash admin password: %v", err)
			return
		}
		admin := models.User{
			Username: "admin",
			Password: string(hashedPassword),
			Name:     "系统管理员",
			Role:     "admin",
		}
		if err := DB.Create(&admin).Error; err != nil {
			log.Printf("Failed to create default admin: %v", err)
		} else {
			log.Println("Default admin created: admin/admin123")
		}
	}

	var empCount int64
	DB.Model(&models.User{}).Where("username = ?", "employee").Count(&empCount)
	if empCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("employee123"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash employee password: %v", err)
			return
		}
		employee := models.User{
			Username: "employee",
			Password: string(hashedPassword),
			Name:     "测试员工",
			Role:     "employee",
		}
		if err := DB.Create(&employee).Error; err != nil {
			log.Printf("Failed to create test employee: %v", err)
		} else {
			log.Println("Test employee created: employee/employee123")
		}
	}
}
