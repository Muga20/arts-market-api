package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/muga20/artsMarket/modules/notifications/emails"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
)

const TypeSendEmail = "email:send"

// EmailTaskPayload represents the data we need to send an email.
type EmailTaskPayload struct {
	ToEmail    string `json:"to_email"`
	ResetToken string `json:"reset_token"`
}

// NewSendEmailTask creates a new task to send a password reset email.
func NewSendEmailTask(toEmail, resetToken string) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailTaskPayload{ToEmail: toEmail, ResetToken: resetToken})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendEmail, payload), nil
}

// HandleSendEmailTask is the function that handles the email sending task.
func HandleSendEmailTask(ctx context.Context, t *asynq.Task) error {
	var payload EmailTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to parse task payload: %v", err)
	}

	// Call your existing email sending function
	return emails.SendPasswordResetEmail(payload.ToEmail, payload.ResetToken, &handlers.ResponseHandler{})
}
