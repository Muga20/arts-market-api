package auth

import (
	"errors"
	"fmt"
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
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		// Validate email format
		if !validation.IsValidEmail(req.Email) {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid email format"))
		}

		// Rate-limiting
		if utils.IsRateLimited(req.Email) {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusTooManyRequests, "Too many password reset attempts. Please try again later"))
		}

		// Retrieve user by email
		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Security: Always return same message whether user exists or not
				return responseHandler.HandleResponse(c, fiber.Map{
					"message": "If an account exists with this email, a password reset link has been sent",
				}, nil)
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to retrieve user: %w", err))
		}

		// Generate password reset token
		resetToken := uuid.New().String()
		resetTokenExpiry := time.Now().Add(time.Hour * 1)

		// Start transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Update user security record
		var userSecurity models.UserSecurity
		if err := tx.Where("user_id = ?", user.ID).First(&userSecurity).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to retrieve user security information: %w", err))
		}

		userSecurity.PasswordResetToken = &resetToken
		userSecurity.PasswordResetExpiresAt = &resetTokenExpiry
		if err := tx.Save(&userSecurity).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update reset token: %w", err))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		// Queue email task
		task, err := tasks.NewSendEmailTask(req.Email, resetToken)
		if err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to create email task: %w", err))
		}

		client := asynq.NewClient(*config.RedisConfig)
		defer client.Close()

		if _, err := client.Enqueue(task); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to enqueue email task: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "If an account exists with this email, a password reset link has been sent",
		}, nil)
	}
}
