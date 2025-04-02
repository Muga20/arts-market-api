package security

import (
	"errors"

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
		// Get user from context (assumed to be set by the authentication middleware)
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false,
				"message": "user not found in context",
			}, errors.New("user not found in context"))
		}

		// Parse the body to get the password for verification
		var req DeleteAccountRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false,
				"message": "invalid request body",
			}, err)
		}

		// Fetch the user’s hashed password from the database
		var userSecurity models.UserSecurity
		if err := db.Where("user_id = ?", user.ID).First(&userSecurity).Error; err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false,
				"message": "user not found",
			}, err)
		}

		// Compare the provided password with the hashed password stored in the database
		err := bcrypt.CompareHashAndPassword([]byte(userSecurity.Password), []byte(req.Password))
		if err != nil {
			// If password does not match
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false,
				"message": "incorrect password",
			}, errors.New("incorrect password"))
		}

		// Begin transaction to delete user and user security data
		err = db.Transaction(func(tx *gorm.DB) error {
			// Delete the associated UserSecurity record
			if err := tx.Where("user_id = ?", user.ID).Delete(&models.UserSecurity{}).Error; err != nil {
				return err
			}

			// Delete the User record
			if err := tx.Delete(&models.User{}, user.ID).Error; err != nil {
				return err
			}

			return nil
		})

		// Handle error during transaction
		if err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false,
				"message": "failed to delete account",
			}, err)
		}

		// Return success message
		return responseHandler.Handle(c, map[string]interface{}{
			"success": true,
			"message": "account deleted successfully",
		}, nil)
	}
}

