package auth

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/tasks"
	"github.com/muga20/artsMarket/pkg/utils"
	"github.com/muga20/artsMarket/pkg/validation"
	"gorm.io/gorm"
)

// ResetPasswordRequest represents the expected reset password request payload
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequestHandler generates a reset password token and sends it to the user's email
// @Summary Request password reset
// @Description Generate a password reset token and send it to the user's email address
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param request body ResetPasswordRequest true "Password reset request payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/reset-password-request [post]
func ResetPasswordRequestHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req ResetPasswordRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.Handle(c, nil, errors.New("invalid request payload"))
		}

		// Validate email format
		if !validation.IsValidEmail(req.Email) {
			return responseHandler.Handle(c, nil, errors.New("invalid email format"))
		}

		// Rate-limiting: Prevent multiple password reset requests in a short period
		if utils.IsRateLimited(req.Email) {
			return responseHandler.Handle(c, nil, errors.New("too many password reset attempts. Please try again later"))
		}

		// Retrieve user by email
		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Do not disclose whether the email exists for security reasons
				return responseHandler.Handle(c, nil, errors.New("password reset request received"))
			}
			return responseHandler.Handle(c, nil, errors.New("failed to retrieve user"))
		}

		// Generate password reset token
		resetToken := uuid.New().String()
		resetTokenExpiry := time.Now().Add(time.Hour * 1)

		// Update user security record with the reset token and expiry
		var userSecurity models.UserSecurity
		if err := db.Where("user_id = ?", user.ID).First(&userSecurity).Error; err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to retrieve user security information"))
		}

		// Update the reset token in the database
		userSecurity.PasswordResetToken = &resetToken
		userSecurity.PasswordResetExpiresAt = &resetTokenExpiry
		if err := db.Save(&userSecurity).Error; err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to update reset token"))
		}

		// Queue the email task instead of sending it directly
		task, err := tasks.NewSendEmailTask(req.Email, resetToken)
		if err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to create email task"))
		}

		// Connect to Redis using the centralized config (using the already established Redis client)
		client := asynq.NewClient(*config.RedisConfig)
		defer client.Close()

		// Enqueue the task for background processing
		_, err = client.Enqueue(task)
		if err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to enqueue email task"))
		}

		return responseHandler.Handle(c, fiber.Map{
			"message": "Password reset token sent to your email",
		}, nil)
	}
}
