package handlers

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"time"

	"github.com/muga20/artsMarket/pkg/logs/models"
	"gorm.io/gorm"
)

// Handle processes and returns a response with proper error handling
func (rh *ResponseHandler) Handle(c *fiber.Ctx, data interface{}, err error) error {
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(data)
}

// GetLogsHandler handles fetching error logs with pagination
// @Summary Get error logs with pagination
// @Description Get error logs with pagination (supports query parameters for pagination)
// @Tags Logs
// @Accept  json
// @Produce  json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /logs [get]
func GetLogsHandler(db *gorm.DB, responseHandler *ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := 1
		pageSize := 10
		if pageParam := c.Query("page"); pageParam != "" {
			fmt.Sscanf(pageParam, "%d", &page)
		}
		if pageSizeParam := c.Query("page_size"); pageSizeParam != "" {
			fmt.Sscanf(pageSizeParam, "%d", &pageSize)
		}

		var logs []models.ErrorLog
		var total int64
		offset := (page - 1) * pageSize

		// Get the total number of logs
		if err := db.Model(&models.ErrorLog{}).Count(&total).Error; err != nil {
			return responseHandler.Handle(c, nil, fmt.Errorf("failed to count logs: %v", err))
		}

		// Fetch the logs with pagination
		if err := db.Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
			return responseHandler.Handle(c, nil, fmt.Errorf("failed to retrieve logs: %v", err))
		}

		return responseHandler.Handle(c, fiber.Map{
			"total":     total,
			"logs":      logs,
			"page":      page,
			"page_size": pageSize,
		}, nil)
	}
}

// DeleteLogsHandler handles deleting error logs based on time range
// @Summary Delete error logs based on date or time range
// @Description Delete logs by date, weekly, monthly, or all logs
// @Tags Logs
// @Accept  json
// @Produce  json
// @Param date query string false "Date (format: YYYY-MM-DD)"
// @Param range query string false "Time range (daily, weekly, monthly, all)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /logs [delete]
func DeleteLogsHandler(db *gorm.DB, responseHandler *ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		date := c.Query("date")
		rangeParam := c.Query("range")

		var err error

		switch rangeParam {
		case "daily":
			if date == "" {
				return responseHandler.Handle(c, nil, errors.New("date is required for daily range"))
			}
			parsedDate, parseErr := time.Parse("2006-01-02", date)
			if parseErr != nil {
				return responseHandler.Handle(c, nil, errors.New("invalid date format"))
			}
			err = db.Where("DATE(occurred_at) = ?", parsedDate.Format("2006-01-02")).Delete(&models.ErrorLog{}).Error
		case "weekly":
			err = db.Where("occurred_at >= ? AND occurred_at <= ?", time.Now().AddDate(0, 0, -7), time.Now()).Delete(&models.ErrorLog{}).Error
		case "monthly":
			err = db.Where("occurred_at >= ? AND occurred_at <= ?", time.Now().AddDate(0, -1, 0), time.Now()).Delete(&models.ErrorLog{}).Error
		case "all":
			err = db.Delete(&models.ErrorLog{}).Error
		default:
			return responseHandler.Handle(c, nil, errors.New("invalid range parameter, valid values are: daily, weekly, monthly, all"))
		}

		if err != nil {
			return responseHandler.Handle(c, nil, fmt.Errorf("failed to delete logs: %v", err))
		}

		return responseHandler.Handle(c, fiber.Map{"message": "Logs deleted successfully"}, nil)
	}
}
