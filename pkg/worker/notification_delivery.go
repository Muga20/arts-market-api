package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/modules/notifications/models"
	"github.com/muga20/artsMarket/modules/notifications/services"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// NotificationWorker is the background worker that processes notifications.
type NotificationWorker struct {
	client          *asynq.Client
	service         *services.NotificationService
	ResponseHandler *handlers.ResponseHandler
	db              *gorm.DB // Database connection
}

// NewNotificationWorker creates a new notification worker with database support
func NewNotificationWorker(
	service *services.NotificationService,
	responseHandler *handlers.ResponseHandler,
	db *gorm.DB, // Database connection parameter
) *NotificationWorker {
	client := asynq.NewClient(*config.RedisConfig)
	return &NotificationWorker{
		client:          client,
		service:         service,
		ResponseHandler: responseHandler,
		db:              db,
	}
}

// Start starts the worker to process tasks
func (w *NotificationWorker) Start() {
	// Create the Asynq server with Redis connection options
	server := asynq.NewServer(
		*config.RedisConfig,
		asynq.Config{
			Concurrency: 5,
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {

				return time.Second * time.Duration(n*2)
			},
			// Removed MaxRetry as it is not a valid field in asynq.Config
		},
	)

	// Register the task handler
	mux := asynq.NewServeMux()
	mux.HandleFunc("notification:send", w.handleNotificationTask)

	// Set MaxRetry for individual tasks when enqueuing them
	// Example: w.client.Enqueue(asynq.NewTask("notification:send", payload), asynq.MaxRetry(3))

	// Start the server
	if err := server.Start(mux); err != nil {
		w.ResponseHandler.Handle(nil, nil, fmt.Errorf("error starting worker: %v", err))
	}
}

// handleNotificationTask processes and persists a notification
func (w *NotificationWorker) handleNotificationTask(ctx context.Context, task *asynq.Task) error {
	var notification models.Notification

	// Decode the task payload
	if err := json.Unmarshal(task.Payload(), &notification); err != nil {
		w.ResponseHandler.Handle(nil, nil, fmt.Errorf("failed to unmarshal notification: %v", err))
		return fmt.Errorf("failed to unmarshal notification: %v", err)
	}

	log.Printf("Processing notification for user %s: %s", notification.UserID, notification.Message)

	// Save to database with transaction
	err := w.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&notification).Error; err != nil {
			return fmt.Errorf("failed to save notification: %v", err)
		}
		return nil
	})

	if err != nil {
		w.ResponseHandler.Handle(nil, nil, fmt.Errorf("database transaction failed: %v", err))
		return err
	}

	log.Printf("Successfully saved notification ID %s for user %s",
		notification.ID, notification.UserID)

	return nil
}
