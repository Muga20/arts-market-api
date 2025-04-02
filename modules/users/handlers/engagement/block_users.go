package engagement

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// BlockUser godoc
// @Summary Block a user
// @Description Allows an authenticated user to block another user
// @Tags Engagement
// @Accept json
// @Produce json
// @Param id path string true "User ID to block"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /engagement/block/{id} [post]
func BlockUser(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "User not authenticated"}, fmt.Errorf("unauthenticated request"))
		}

		blockedUserID := c.Params("id")
		if user.ID.String() == blockedUserID {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "You cannot block yourself"}, nil)
		}

		blockedUUID, err := uuid.Parse(blockedUserID)
		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Invalid user ID"}, err)
		}

		// Check if user is already blocked
		var count int64
		err = db.Model(&models.BlockedUser{}).Where("user_id = ? AND blocked_user_id = ?", user.ID, blockedUUID).Count(&count).Error
		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Database error"}, err)
		}
		if count > 0 {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "User is already blocked"}, nil)
		}

		// Block user
		blockEntry := models.BlockedUser{
			UserID:        user.ID,
			BlockedUserID: blockedUUID,
		}
		if err := db.Create(&blockEntry).Error; err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Could not block user"}, err)
		}

		return responseHandler.HandleResponse(c, fiber.Map{"success": true, "message": "User blocked successfully"}, nil)
	}
}

// UnblockUser godoc
// @Summary Unblock a user
// @Description Allows an authenticated user to unblock someone they have blocked
// @Tags Engagement
// @Accept json
// @Produce json
// @Param id path string true "User ID to unblock"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /engagement/unblock/{id} [delete]
func UnblockUser(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "User not authenticated"}, fmt.Errorf("unauthenticated request"))
		}

		blockedUserID := c.Params("id")
		blockedUUID, err := uuid.Parse(blockedUserID)
		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Invalid user ID"}, err)
		}

		// Delete block record
		if err := db.Where("user_id = ? AND blocked_user_id = ?", user.ID, blockedUUID).Delete(&models.BlockedUser{}).Error; err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Could not unblock user"}, err)
		}

		return responseHandler.HandleResponse(c, fiber.Map{"success": true, "message": "User unblocked successfully"}, nil)
	}
}

// GetBlockedUsers godoc
// @Summary Get list of users blocked by the logged-in user
// @Description Retrieves all users blocked by the authenticated user, including usernames and names
// @Tags Engagement
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /engagement/blocked [get]
func GetBlockedUsers(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "User not authenticated"}, fmt.Errorf("unauthenticated request"))
		}

		// Struct to hold the result of blocked users with necessary fields
		var blockedUsers []struct {
			BlockedUserID uuid.UUID `json:"blocked_user_id"`
			Username      string    `json:"username"`
			FirstName     string    `json:"first_name"`
			LastName      string    `json:"last_name"`
		}

		// Perform a join with users table and user_details table to get the blocked user's username and names
		err := db.Table("blocked_users").
			Select("blocked_users.blocked_user_id, users.username, user_details.first_name, user_details.last_name").
			Joins("JOIN users ON blocked_users.blocked_user_id = users.id").
			Joins("LEFT JOIN user_details ON blocked_users.blocked_user_id = user_details.user_id").
			Where("blocked_users.user_id = ?", user.ID).
			Find(&blockedUsers).Error

		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Error retrieving blocked users"}, err)
		}

		// If no blocked users found, return an empty list
		if len(blockedUsers) == 0 {
			return responseHandler.HandleResponse(c, fiber.Map{
				"success":       true,
				"blocked_users": []interface{}{},
			}, nil)
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"success":       true,
			"blocked_users": blockedUsers,
		}, nil)
	}
}
