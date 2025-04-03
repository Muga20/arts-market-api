package security

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// DeleteAccountRequest defines the structure of the request body for deleting an account
type DeleteAccountRequest struct {
	Password string `json:"password" validate:"required"`
}

// DeleteAccount handles the deletion of the user's account after password verification
// @Summary Delete the authenticated user's account
// @Description Permanently deletes the user's account and associated data after password verification
// @Tags Security
// @Accept  json
// @Produce  json
// @Param   body  body  DeleteAccountRequest  true  "Password for verification"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /security/delete-account [delete]
func DeleteAccount(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authentication check with proper type assertion
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Parse request body
		var req DeleteAccountRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))
		}

		// Validate password is provided
		if req.Password == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Password is required"))
		}

		// Fetch user security record
		var userSecurity models.UserSecurity
		if err := db.Where("user_id = ?", user.ID).First(&userSecurity).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "User account not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to fetch user security: %w", err))
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(userSecurity.Password), []byte(req.Password)); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Incorrect password"))
		}

		// Start explicit transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Soft delete by setting is_active to false
		if err := tx.Model(&models.User{}).Where("id = ?", user.ID).
			Update("is_active", false).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to deactivate user: %w", err))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Account deactivated successfully",
		}, nil)
	}
}
