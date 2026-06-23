package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"`
	Name      string    `gorm:"not null" json:"name"`
	Role      string    `gorm:"not null;default:'employee'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Device struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Category     string    `gorm:"not null" json:"category"`
	SerialNumber string    `gorm:"unique;not null" json:"serial_number"`
	Status       string    `gorm:"not null;default:'available'" json:"status"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type BorrowRecord struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"not null" json:"user_id"`
	DeviceID       uint       `gorm:"not null" json:"device_id"`
	BorrowDate     time.Time  `gorm:"not null" json:"borrow_date"`
	ExpectedReturn time.Time  `gorm:"not null" json:"expected_return"`
	ActualReturn   *time.Time `json:"actual_return"`
	Status         string     `gorm:"not null;default:'borrowed'" json:"status"`
	User           User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Device         Device     `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role"`
}

type DeviceRequest struct {
	Name         string `json:"name" binding:"required"`
	Category     string `json:"category" binding:"required"`
	SerialNumber string `json:"serial_number" binding:"required"`
	Description  string `json:"description"`
}

type BorrowRequest struct {
	DeviceID       uint      `json:"device_id" binding:"required"`
	ExpectedReturn time.Time `json:"expected_return" binding:"required"`
}

type ReturnRequest struct {
	RecordID uint `json:"record_id" binding:"required"`
}

type ExtendRequest struct {
	RecordID uint `json:"record_id" binding:"required"`
	Days     int  `json:"days" binding:"required,min=1,max=30"`
}
