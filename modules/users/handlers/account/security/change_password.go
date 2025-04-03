package security

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ChangePasswordRequest defines the structure for password change requests
type ChangePasswordRequest struct {
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	CurrentPassword string `json:"current_password"`
}

// ChangePassword allows the user to change their password
// @Summary Change the authenticated user's password
// @Description Allows the user to change their password without reusing the previous one
// @Tags Security
// @Accept  json
// @Produce  json
// @Param   body  body  ChangePasswordRequest  true  "New password data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /security/change-password [put]
func ChangePassword(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authentication check with proper type assertion
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Parse request body
		var req ChangePasswordRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))
		}

		// Validate password requirements
		if len(req.NewPassword) < 8 {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Password must be at least 8 characters"))
		}

		// Fetch user security record
		var userSecurity models.UserSecurity
		if err := db.Where("user_id = ?", user.ID).First(&userSecurity).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to fetch user security: %w", err))
		}

		// Verify current password if provided (for password change flow)
		if req.CurrentPassword != "" {
			if err := bcrypt.CompareHashAndPassword([]byte(userSecurity.Password), []byte(req.CurrentPassword)); err != nil {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusUnauthorized, "Current password is incorrect"))
			}
		}

		// Check if new password is same as current
		if err := bcrypt.CompareHashAndPassword([]byte(userSecurity.Password), []byte(req.NewPassword)); err == nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "New password must be different from current password"))
		}

		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to hash password: %w", err))
		}

		// Start transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Update password and clear any password reset tokens
		userSecurity.Password = string(hashedPassword)
		userSecurity.PasswordResetToken = nil
		userSecurity.PasswordResetExpiresAt = nil

		if err := tx.Save(&userSecurity).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update password: %w", err))
		}

		// Invalidate all active sessions (optional)
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.UserSession{}).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to clear sessions: %w", err))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Password changed successfully. Please log in again.",
		}, nil)
	}
}
