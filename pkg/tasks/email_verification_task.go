package tasks

import (
    "encoding/json"
    "errors"

    "github.com/hibiken/asynq"
)

// EmailVerificationPayload defines the payload structure for email verification tasks
type EmailVerificationPayload struct {
    Email string `json:"email"`
    Token string `json:"token"`
}

// NewSendEmailVerificationTask creates a new task for sending email verification
func NewSendEmailVerificationTask(email, token string) (*asynq.Task, error) {
    payload, err := json.Marshal(EmailVerificationPayload{
        Email: email,
        Token: token,
    })
    if err != nil {
        return nil, errors.New("failed to marshal email verification payload")
    }

    return asynq.NewTask("email:send_verification", payload), nil
}
