package security

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ChangePasswordRequest defines the structure for password change requests
type ChangePasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8"`
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
		// Get the user from context (assumed to be set by authentication middleware)
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false,
				"message": "user not found in context",
			}, errors.New("user not found in context"))
		}

		// Parse the request body for the new password
		var req ChangePasswordRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false,
				"message": "invalid request body",
			}, err)
		}

		// Validate the new password length
		if len(req.NewPassword) < 8 {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false,
				"message": "password must be at least 8 characters",
			}, errors.New("password too short"))
		}

		// Retrieve the user's current hashed password from the database
		var userSecurity models.UserSecurity
		if err := db.Where("user_id = ?", user.ID).First(&userSecurity).Error; err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false,
				"message": "failed to retrieve user security data",
			}, err)
		}

		// Check if the new password is the same as the old one
		err := bcrypt.CompareHashAndPassword([]byte(userSecurity.Password), []byte(req.NewPassword))
		if err == nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false,
				"message": "new password must not be the same as the previous password",
			}, errors.New("password reuse attempt"))
		}

		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false,
				"message": "failed to hash password",
			}, err)
		}

		// Update the password in the user security table
		userSecurity.Password = string(hashedPassword)
		if err := db.Save(&userSecurity).Error; err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false,
				"message": "failed to update password",
			}, err)
		}

		// Return a success response
		return responseHandler.Handle(c, map[string]interface{}{
			"success": true,
			"message": "password changed successfully",
		}, nil)
	}
}
