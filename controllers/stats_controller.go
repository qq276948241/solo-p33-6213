package controllers

import (
	"equipment-borrow-system/config"
	"equipment-borrow-system/models"
	"equipment-borrow-system/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type StatsController struct{}

func NewStatsController() *StatsController {
	return &StatsController{}
}

func (sc *StatsController) GetOverview(c *gin.Context) {
	var totalDevices int64
	if err := config.DB.Model(&models.Device{}).Count(&totalDevices).Error; err != nil {
		utils.InternalError(c, "Failed to count devices")
		return
	}

	var availableDevices int64
	if err := config.DB.Model(&models.Device{}).Where("status = ?", "available").Count(&availableDevices).Error; err != nil {
		utils.InternalError(c, "Failed to count available devices")
		return
	}

	var borrowedDevices int64
	if err := config.DB.Model(&models.Device{}).Where("status = ?", "borrowed").Count(&borrowedDevices).Error; err != nil {
		utils.InternalError(c, "Failed to count borrowed devices")
		return
	}

	var totalUsers int64
	if err := config.DB.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		utils.InternalError(c, "Failed to count users")
		return
	}

	var activeBorrows int64
	if err := config.DB.Model(&models.BorrowRecord{}).Where("status = ?", "borrowed").Count(&activeBorrows).Error; err != nil {
		utils.InternalError(c, "Failed to count active borrows")
		return
	}

	now := time.Now()
	var overdueCount int64
	if err := config.DB.Model(&models.BorrowRecord{}).
		Where("status = 'borrowed' AND expected_return < ?", now).
		Count(&overdueCount).Error; err != nil {
		utils.InternalError(c, "Failed to count overdue records")
		return
	}

	var totalRecords int64
	if err := config.DB.Model(&models.BorrowRecord{}).Count(&totalRecords).Error; err != nil {
		utils.InternalError(c, "Failed to count total records")
		return
	}

	var returnedRecords int64
	if err := config.DB.Model(&models.BorrowRecord{}).
		Where("status IN ?", []string{"returned", "overdue_returned"}).
		Count(&returnedRecords).Error; err != nil {
		utils.InternalError(c, "Failed to count returned records")
		return
	}

	utils.Success(c, gin.H{
		"total_devices":      totalDevices,
		"available_devices":  availableDevices,
		"borrowed_devices":   borrowedDevices,
		"total_users":        totalUsers,
		"active_borrows":     activeBorrows,
		"overdue_count":      overdueCount,
		"total_records":      totalRecords,
		"returned_records":   returnedRecords,
		"borrow_rate":        float64(borrowedDevices) / float64(totalDevices) * 100,
		"overdue_rate":       float64(overdueCount) / float64(activeBorrows) * 100,
	})
}

func (sc *StatsController) GetByCategory(c *gin.Context) {
	type CategoryStats struct {
		Category      string `json:"category"`
		Total         int64  `json:"total"`
		Available     int64  `json:"available"`
		Borrowed      int64  `json:"borrowed"`
	}

	var categories []CategoryStats

	rows, err := config.DB.Model(&models.Device{}).
		Select("category, COUNT(*) as total").
		Group("category").
		Rows()

	if err != nil {
		utils.InternalError(c, "Failed to get category stats")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var cs CategoryStats
		rows.Scan(&cs.Category, &cs.Total)

		config.DB.Model(&models.Device{}).
			Where("category = ? AND status = ?", cs.Category, "available").
			Count(&cs.Available)

		config.DB.Model(&models.Device{}).
			Where("category = ? AND status = ?", cs.Category, "borrowed").
			Count(&cs.Borrowed)

		categories = append(categories, cs)
	}

	utils.Success(c, categories)
}

func (sc *StatsController) GetOverdueDetails(c *gin.Context) {
	now := time.Now()

	var records []models.BorrowRecord
	if err := config.DB.Preload("User").Preload("Device").
		Where("status = 'borrowed' AND expected_return < ?", now).
		Order("expected_return asc").
		Find(&records).Error; err != nil {
		utils.InternalError(c, "Failed to get overdue details")
		return
	}

	type OverdueDetail struct {
		ID              uint      `json:"id"`
		UserName        string    `json:"user_name"`
		DeviceName      string    `json:"device_name"`
		DeviceCategory  string    `json:"device_category"`
		BorrowDate      time.Time `json:"borrow_date"`
		ExpectedReturn  time.Time `json:"expected_return"`
		DaysOverdue     int       `json:"days_overdue"`
	}

	var details []OverdueDetail
	for _, r := range records {
		days := int(now.Sub(r.ExpectedReturn).Hours() / 24)
		if days < 1 {
			days = 1
		}
		details = append(details, OverdueDetail{
			ID:             r.ID,
			UserName:       r.User.Name,
			DeviceName:     r.Device.Name,
			DeviceCategory: r.Device.Category,
			BorrowDate:     r.BorrowDate,
			ExpectedReturn: r.ExpectedReturn,
			DaysOverdue:    days,
		})
	}

	utils.Success(c, details)
}
