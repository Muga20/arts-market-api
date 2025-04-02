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
)

// NotificationWorker is the background worker that processes notifications.
type NotificationWorker struct {
	client          *asynq.Client
	service         *services.NotificationService
	ResponseHandler *handlers.ResponseHandler
}

// NewNotificationWorker creates a new notification worker
func NewNotificationWorker(service *services.NotificationService, responseHandler *handlers.ResponseHandler) *NotificationWorker {
	client := asynq.NewClient(*config.RedisConfig) // Initialize Redis client
	return &NotificationWorker{
		client:          client,
		service:         service,
		ResponseHandler: responseHandler,
	}
}

// Start starts the worker to process tasks
func (w *NotificationWorker) Start() {
	// Create the Asynq server with Redis connection options
	server := asynq.NewServer(
		*config.RedisConfig, // Correctly use RedisConfig
		asynq.Config{
			Concurrency: 5, // Set concurrency to process tasks in parallel
		},
	)

	// Register the task handler for processing "notification:send" tasks
	mux := asynq.NewServeMux()
	mux.HandleFunc("notification:send", w.handleTask)

	// Start the server with the mux (handler)
	err := server.Start(mux)
	if err != nil {
		// Use ResponseHandler to log error to the database
		w.ResponseHandler.Handle(nil, nil, fmt.Errorf("Error starting worker: %v", err))
		return
	}
}

// handleTask processes an individual notification task with retry logic
func (w *NotificationWorker) handleTask(ctx context.Context, task *asynq.Task) error {
	var payload models.Notification

	// Decode the task payload manually
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		// Use ResponseHandler to log error to the database instead of logging to the terminal
		w.ResponseHandler.Handle(nil, nil, fmt.Errorf("failed to unmarshal task payload: %v", err))
		// Returning an error will trigger a retry, based on the retry policy
		return fmt.Errorf("failed to unmarshal task payload: %v", err)
	}

	// Log the notification action (this is where actual sending logic would go)
	log.Printf("Sending notification to user %v: %v", payload.UserID, payload.Message)

	// Simulating a transient error (e.g., network issue) to trigger retry
	if someTransientErrorOccurred() {
		// Log the error
		w.ResponseHandler.Handle(nil, nil, fmt.Errorf("transient error sending notification to user %v", payload.UserID))
		// Returning an error will trigger a retry, based on the retry policy
		return fmt.Errorf("transient error sending notification")
	}

	// If everything is successful, just return nil to indicate success
	return nil
}

// someTransientErrorOccurred simulates a transient error (for demonstration purposes)
func someTransientErrorOccurred() bool {
	// Simulate a failure with a probability of 30%
	return time.Now().Unix()%3 == 0
}
