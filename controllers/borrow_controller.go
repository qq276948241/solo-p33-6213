package controllers

import (
	"equipment-borrow-system/config"
	"equipment-borrow-system/models"
	"equipment-borrow-system/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type BorrowController struct{}

func NewBorrowController() *BorrowController {
	return &BorrowController{}
}

func (bc *BorrowController) Borrow(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req models.BorrowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body")
		return
	}

	if req.ExpectedReturn.Before(time.Now()) {
		utils.BadRequest(c, "Expected return date cannot be in the past")
		return
	}

	var device models.Device
	if err := config.DB.First(&device, req.DeviceID).Error; err != nil {
		utils.NotFound(c, "Device not found")
		return
	}

	if device.Status != "available" {
		utils.BadRequest(c, "Device is not available for borrowing")
		return
	}

	tx := config.DB.Begin()

	device.Status = "borrowed"
	if err := tx.Save(&device).Error; err != nil {
		tx.Rollback()
		utils.InternalError(c, "Failed to update device status")
		return
	}

	record := models.BorrowRecord{
		UserID:         userID.(uint),
		DeviceID:       req.DeviceID,
		BorrowDate:     time.Now(),
		ExpectedReturn: req.ExpectedReturn,
		Status:         "borrowed",
	}

	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		utils.InternalError(c, "Failed to create borrow record")
		return
	}

	tx.Commit()

	record.Device = device
	var user models.User
	config.DB.First(&user, userID)
	record.User = user

	c.JSON(http.StatusCreated, utils.Response{
		Code:    0,
		Message: "Borrow request successful",
		Data:    record,
	})
}

func (bc *BorrowController) Return(c *gin.Context) {
	var req models.ReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body")
		return
	}

	var record models.BorrowRecord
	if err := config.DB.First(&record, req.RecordID).Error; err != nil {
		utils.NotFound(c, "Borrow record not found")
		return
	}

	if record.Status != "borrowed" {
		utils.BadRequest(c, "This device has already been returned")
		return
	}

	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	if role != "admin" && record.UserID != userID.(uint) {
		utils.Forbidden(c, "You can only return your own borrowed items")
		return
	}

	tx := config.DB.Begin()

	now := time.Now()
	record.ActualReturn = &now
	if now.After(record.ExpectedReturn) {
		record.Status = "overdue_returned"
	} else {
		record.Status = "returned"
	}

	if err := tx.Save(&record).Error; err != nil {
		tx.Rollback()
		utils.InternalError(c, "Failed to update borrow record")
		return
	}

	if err := tx.Model(&models.Device{}).Where("id = ?", record.DeviceID).Update("status", "available").Error; err != nil {
		tx.Rollback()
		utils.InternalError(c, "Failed to update device status")
		return
	}

	tx.Commit()

	utils.Success(c, record)
}

func (bc *BorrowController) GetMyRecords(c *gin.Context) {
	userID, _ := c.Get("userID")
	status := c.Query("status")

	query := config.DB.Preload("Device").Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var records []models.BorrowRecord
	if err := query.Order("created_at desc").Find(&records).Error; err != nil {
		utils.InternalError(c, "Failed to get borrow records")
		return
	}

	utils.Success(c, records)
}

func (bc *BorrowController) GetAllRecords(c *gin.Context) {
	status := c.Query("status")
	userID := c.Query("user_id")
	deviceID := c.Query("device_id")

	query := config.DB.Preload("User").Preload("Device")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}

	var records []models.BorrowRecord
	if err := query.Order("created_at desc").Find(&records).Error; err != nil {
		utils.InternalError(c, "Failed to get borrow records")
		return
	}

	utils.Success(c, records)
}

func (bc *BorrowController) GetOverdueRecords(c *gin.Context) {
	now := time.Now()

	var records []models.BorrowRecord
	if err := config.DB.Preload("User").Preload("Device").
		Where("status = 'borrowed' AND expected_return < ?", now).
		Order("expected_return asc").
		Find(&records).Error; err != nil {
		utils.InternalError(c, "Failed to get overdue records")
		return
	}

	utils.Success(c, records)
}

func (bc *BorrowController) Extend(c *gin.Context) {
	var req models.ExtendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body, days must be between 1 and 30")
		return
	}

	var record models.BorrowRecord
	if err := config.DB.Preload("Device").Preload("User").First(&record, req.RecordID).Error; err != nil {
		utils.NotFound(c, "Borrow record not found")
		return
	}

	if record.Status != "borrowed" {
		utils.BadRequest(c, "Only active borrow records can be extended")
		return
	}

	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	if role != "admin" && record.UserID != userID.(uint) {
		utils.Forbidden(c, "You can only extend your own borrowed items")
		return
	}

	newExpectedReturn := record.ExpectedReturn.AddDate(0, 0, req.Days)
	record.ExpectedReturn = newExpectedReturn

	if err := config.DB.Save(&record).Error; err != nil {
		utils.InternalError(c, "Failed to extend borrow record")
		return
	}

	utils.Success(c, record)
}

func (bc *BorrowController) GetRecordByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid record ID")
		return
	}

	var record models.BorrowRecord
	if err := config.DB.Preload("User").Preload("Device").First(&record, uint(id)).Error; err != nil {
		utils.NotFound(c, "Borrow record not found")
		return
	}

	utils.Success(c, record)
}
