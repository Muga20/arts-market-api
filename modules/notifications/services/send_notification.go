package services

import (
	"fmt"
	"github.com/google/uuid" // Import the uuid package
	"github.com/hibiken/asynq"
	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/modules/notifications/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
)

// NotificationService holds the necessary components for handling notifications
type NotificationService struct {
	RedisClient     *asynq.Client
	ResponseHandler *handlers.ResponseHandler
}

// NewNotificationService creates a new instance of NotificationService with a Redis client
func NewNotificationService(responseHandler *handlers.ResponseHandler) *NotificationService {
	redisClient := asynq.NewClient(*config.RedisConfig)
	return &NotificationService{
		RedisClient:     redisClient,
		ResponseHandler: responseHandler,
	}
}

// EnqueueNotification enqueues a notification task to be processed later
func (s *NotificationService) EnqueueNotification(userID, senderID, notificationType, message, entityType, entityID string) error {
	// Convert userID from string to uuid.UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return s.ResponseHandler.Handle(nil, nil, fmt.Errorf("invalid userID format: %v", err))
	}

	senderUUID, err := uuid.Parse(senderID)
	if err != nil {
		return s.ResponseHandler.Handle(nil, nil, fmt.Errorf("invalid senderID format: %v", err))
	}

	entityUUID, err := uuid.Parse(entityID)
	if err != nil {
		return s.ResponseHandler.Handle(nil, nil, fmt.Errorf("invalid entityID format: %v", err))
	}

	// Create the notification model
	notification := models.Notification{
		UserID:            userUUID,
		SenderID:          &senderUUID,
		NotificationType:  notificationType,
		Message:           message,
		EntityType:        entityType,
		EntityID:          &entityUUID,
		IsSystemGenerated: true,
	}

	// Serialize the notification into the task payload
	payload, err := notification.ToJSON()
	if err != nil {
		return s.ResponseHandler.Handle(nil, nil, fmt.Errorf("failed to serialize notification: %v", err))
	}

	// Create a new task for the notification
	task := asynq.NewTask("notification:send", payload)

	// Enqueue the task for background processing
	_, err = s.RedisClient.Enqueue(task)
	if err != nil {
		return s.ResponseHandler.Handle(nil, nil, fmt.Errorf("failed to enqueue task: %v", err))
	}

	return nil
}
