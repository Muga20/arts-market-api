package service

import (
	"github.com/muga20/artsMarket/pkg/logs/models"
	"gorm.io/gorm"
	"time"
)

type LogService struct {
	db *gorm.DB
}

func NewLogService(db *gorm.DB) *LogService {
	return &LogService{db: db}
}

func (s *LogService) CreateErrorLog(log *models.ErrorLog) error {
	// Check for existing similar error from same IP
	var existing models.ErrorLog
	err := s.db.Where(
		"message = ? AND file_name = ? AND method_name = ? AND line_number = ? AND ip_address = ?",
		log.Message,
		log.FileName,
		log.MethodName,
		log.LineNumber,
		log.IPAddress,
	).First(&existing).Error

	if err == nil {
		// Update existing record
		return s.db.Model(&existing).Updates(map[string]interface{}{
			"attempts":         gorm.Expr("attempts + 1"),
			"last_occurred_at": time.Now(),
			"user_agent":       log.UserAgent, // Update user agent in case it changed
		}).Error
	}

	// Create new record if not found
	return s.db.Create(log).Error
}

func (s *LogService) GetErrorLogs() ([]models.ErrorLog, error) {
	var logs []models.ErrorLog
	err := s.db.Find(&logs).Error
	return logs, err
}
