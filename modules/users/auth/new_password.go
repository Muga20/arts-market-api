package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/validation"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ResetPassword represents the expected reset password input
type ResetPassword struct {
	Token    string `json:"token" validate:"required,uuid4"`
	Password string `json:"password" validate:"required,min=8,complex"`
}

// ResetPasswordHandler validates the reset token and updates the password
// @Summary Reset password using reset token
// @Description Validate the reset token and update the user's password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPassword true "Reset Password"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/reset-password [post]
func ResetPasswordHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req ResetPassword
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		// Validate token format (UUID)
		if !validation.IsValidUUID(req.Token) {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid reset token format"))
		}

		// Start database transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Retrieve user security by reset token with row locking
		var userSecurity models.UserSecurity
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("password_reset_token = ?", req.Token).
			First(&userSecurity).Error; err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Invalid reset token"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to retrieve user security: %w", err))
		}

		// Check if token has expired
		if userSecurity.PasswordResetExpiresAt == nil || userSecurity.PasswordResetExpiresAt.Before(time.Now()) {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusGone, "Reset token has expired"))
		}

		// Validate password strength
		if !validation.IsValidPassword(req.Password) {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Password must contain at least 8 characters, including uppercase, lowercase, number, and special character"))
		}

		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to hash password: %w", err))
		}

		// Update user security
		userSecurity.Password = string(hashedPassword)
		userSecurity.PasswordResetToken = nil
		userSecurity.PasswordResetExpiresAt = nil
		userSecurity.UpdatedAt = time.Now()

		if err := tx.Save(&userSecurity).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update password: %w", err))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		// Invalidate all active sessions for this user
		if err := db.Where("user_id = ?", userSecurity.UserID).
			Delete(&models.UserSession{}).Error; err != nil {
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Password successfully reset. All active sessions have been terminated.",
		}, nil)
	}
}
