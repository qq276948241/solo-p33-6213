package controllers

import (
	"equipment-borrow-system/config"
	"equipment-borrow-system/models"
	"equipment-borrow-system/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeviceController struct{}

func NewDeviceController() *DeviceController {
	return &DeviceController{}
}

func (dc *DeviceController) Create(c *gin.Context) {
	var req models.DeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body")
		return
	}

	var count int64
	config.DB.Model(&models.Device{}).Where("serial_number = ?", req.SerialNumber).Count(&count)
	if count > 0 {
		utils.BadRequest(c, "Serial number already exists")
		return
	}

	device := models.Device{
		Name:         req.Name,
		Category:     req.Category,
		SerialNumber: req.SerialNumber,
		Description:  req.Description,
		Status:       models.StatusAvailable,
	}

	if err := config.DB.Create(&device).Error; err != nil {
		utils.InternalError(c, "Failed to create device")
		return
	}

	c.JSON(http.StatusCreated, utils.Response{
		Code:    0,
		Message: "Device created successfully",
		Data:    device,
	})
}

func (dc *DeviceController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid device ID")
		return
	}

	var device models.Device
	if err := config.DB.First(&device, uint(id)).Error; err != nil {
		utils.NotFound(c, "Device not found")
		return
	}

	var req models.DeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body")
		return
	}

	if req.SerialNumber != device.SerialNumber {
		var count int64
		config.DB.Model(&models.Device{}).Where("serial_number = ? AND id != ?", req.SerialNumber, id).Count(&count)
		if count > 0 {
			utils.BadRequest(c, "Serial number already exists")
			return
		}
	}

	device.Name = req.Name
	device.Category = req.Category
	device.SerialNumber = req.SerialNumber
	device.Description = req.Description

	if err := config.DB.Save(&device).Error; err != nil {
		utils.InternalError(c, "Failed to update device")
		return
	}

	utils.Success(c, device)
}

func (dc *DeviceController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid device ID")
		return
	}

	var device models.Device
	if err := config.DB.First(&device, uint(id)).Error; err != nil {
		utils.NotFound(c, "Device not found")
		return
	}

	var borrowCount int64
	config.DB.Model(&models.BorrowRecord{}).Where("device_id = ? AND status = ?", id, models.StatusBorrowed).Count(&borrowCount)
	if borrowCount > 0 {
		utils.BadRequest(c, "Cannot delete device that is currently borrowed")
		return
	}

	if err := config.DB.Delete(&device).Error; err != nil {
		utils.InternalError(c, "Failed to delete device")
		return
	}

	utils.Success(c, gin.H{"message": "Device deleted successfully"})
}

func (dc *DeviceController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid device ID")
		return
	}

	var device models.Device
	if err := config.DB.First(&device, uint(id)).Error; err != nil {
		utils.NotFound(c, "Device not found")
		return
	}

	utils.Success(c, device)
}

func (dc *DeviceController) GetAll(c *gin.Context) {
	category := c.Query("category")
	status := c.Query("status")
	keyword := c.Query("keyword")

	query := config.DB.Model(&models.Device{})

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR serial_number LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var devices []models.Device
	if err := query.Find(&devices).Error; err != nil {
		utils.InternalError(c, "Failed to get devices")
		return
	}

	utils.Success(c, devices)
}
