package engagement

import (
	"fmt"
	"sync"

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
		// Authentication check
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Validate blocked user ID
		blockedUserID := c.Params("id")
		if user.ID.String() == blockedUserID {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "You cannot block yourself"))
		}

		// Parse UUID
		blockedUUID, err := uuid.Parse(blockedUserID)
		if err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid user ID"))
		}

		// Check if user is already blocked
		var count int64
		if err := db.Model(&models.BlockedUser{}).
			Where("user_id = ? AND blocked_user_id = ?", user.ID, blockedUUID).
			Count(&count).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("database error checking block status: %w", err))
		}

		if count > 0 {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "User is already blocked"))
		}

		// Create block entry
		blockEntry := models.BlockedUser{
			UserID:        user.ID,
			BlockedUserID: blockedUUID,
		}

		if err := db.Create(&blockEntry).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("database error creating block: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "User Blocked",
		}, nil)
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
		// Authentication check
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Validate and parse user ID
		blockedUserID := c.Params("id")
		blockedUUID, err := uuid.Parse(blockedUserID)
		if err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid user ID"))
		}

		// Perform unblock operation
		result := db.Where("user_id = ? AND blocked_user_id = ?", user.ID, blockedUUID).
			Delete(&models.BlockedUser{})

		if result.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("database error unblocking user: %w", result.Error))
		}

		// Check if any rows were actually deleted
		if result.RowsAffected == 0 {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusNotFound, "User was not blocked"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "User unblocked",
		}, nil)
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
		// Authentication check
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		type BlockedUser struct {
			BlockedUserID uuid.UUID `json:"blocked_user_id"`
			Username      string    `json:"username"`
			FirstName     string    `json:"first_name,omitempty"`
			LastName      string    `json:"last_name,omitempty"`
			ProfileImage  string    `json:"profile_image,omitempty"`
		}

		var result struct {
			BlockedCount int64         `json:"blocked_count"`
			BlockedUsers []BlockedUser `json:"blocked_users"`
		}

		// Use parallel execution for better performance
		errChan := make(chan error, 2)
		var wg sync.WaitGroup
		wg.Add(2)

		// Get blocked users count
		go func() {
			defer wg.Done()
			if err := db.Model(&models.BlockedUser{}).
				Where("user_id = ?", user.ID).
				Count(&result.BlockedCount).Error; err != nil {
				errChan <- fmt.Errorf("blocked count error: %w", err)
			}
		}()

		// Get blocked users with details
		go func() {
			defer wg.Done()
			if err := db.Table("blocked_users").
				Select(`
                    blocked_users.blocked_user_id,
                    users.username,
                    user_details.first_name,
                    user_details.last_name,
                    user_details.profile_image
                `).
				Joins("JOIN users ON blocked_users.blocked_user_id = users.id").
				Joins("LEFT JOIN user_details ON users.id = user_details.user_id").
				Where("blocked_users.user_id = ?", user.ID).
				Scan(&result.BlockedUsers).Error; err != nil {
				errChan <- fmt.Errorf("blocked users details error: %w", err)
			}
		}()

		wg.Wait()
		close(errChan)

		// Check for errors
		for e := range errChan {
			if e != nil {
				return responseHandler.HandleResponse(c, nil, e)
			}
		}

		return responseHandler.HandleResponse(c, result, nil)
	}
}
