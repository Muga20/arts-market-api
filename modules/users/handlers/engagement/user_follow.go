package engagement

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/muga20/artsMarket/modules/notifications/services"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

const MaxFollowsPerDay = 100

// FollowUser godoc
// @Summary Follow a user
// @Description Allows an authenticated user to follow another user
// @Tags Engagement
// @Accept json
// @Produce json
// @Param id path string true "User ID to follow"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 429 {object} map[string]interface{}
// @Router /engagement/follow/{id} [post]
func FollowUser(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authentication check
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication failed: user not found in context"))
		}

		// Input validation
		followingID := c.Params("id")
		if user.ID.String() == followingID {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "You cannot follow yourself"))
		}

		// Parse UUID
		followingUUID, err := uuid.Parse(followingID)
		if err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid user ID format"))
		}

		// Start database transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Defer rollback in case of error
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Check target user's privacy settings
		var privacySettings models.UserPrivacySetting
		if err := tx.Where("user_id = ?", followingUUID).First(&privacySettings).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// If no settings exist, use default (which allows follow requests)
				privacySettings.AllowFollowRequests = true
			} else {
				tx.Rollback()
				return responseHandler.HandleResponse(c, nil,
					fmt.Errorf("failed to check privacy settings: %w", err))
			}
		}

		// Reject if follow requests not allowed
		if !privacySettings.AllowFollowRequests {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusForbidden, "This user does not allow follow requests"))
		}

		// Check existing follow
		var count int64
		if err := tx.Model(&models.Follower{}).
			Where("follower_id = ? AND following_id = ?", user.ID, followingUUID).
			Count(&count).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("database error checking follow status: %w", err))
		}

		if count > 0 {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "You already follow this user"))
		}

		// Check daily limit
		var followCount int64
		startOfDay := time.Now().Truncate(24 * time.Hour)
		if err := tx.Model(&models.Follower{}).
			Where("follower_id = ? AND created_at >= ?", user.ID, startOfDay).
			Count(&followCount).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("database error checking daily limit: %w", err))
		}

		if followCount >= MaxFollowsPerDay {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusTooManyRequests, "Daily follow limit reached"))
		}

		// Create follow relationship
		newFollow := models.Follower{
			FollowerID:  user.ID,
			FollowingID: followingUUID,
		}
		if err := tx.Create(&newFollow).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("database error creating follow: %w", err))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		// Send notification with profile photo in message (non-blocking)
		go func() {
			var profilePhoto string
			if err := db.Model(&models.UserDetail{}).
				Where("user_id = ?", user.ID).
				Select("profile_image").
				First(&profilePhoto).Error; err == nil {

				// Include profile photo URL in message if available
				notificationService := services.NewNotificationService(responseHandler)
				_ = notificationService.EnqueueNotification(
					followingUUID.String(),
					user.ID.String(),
					"follow",
					fmt.Sprintf("%s|%s followed you", user.Username, profilePhoto),
					"user",
					followingUUID.String(),
				)
			} else {
				// Fallback without photo if there's an error
				notificationService := services.NewNotificationService(responseHandler)
				_ = notificationService.EnqueueNotification(
					followingUUID.String(),
					user.ID.String(),
					"follow",
					fmt.Sprintf("%s followed you", user.Username), // Original format
					"user",
					followingUUID.String(),
				)
			}
		}()

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "You followed this user",
		}, nil)
	}
}

// UnfollowUser godoc
// @Summary Unfollow a user
// @Description Allows an authenticated user to unfollow another user
// @Tags Engagement
// @Accept json
// @Produce json
// @Param id path string true "User ID to unfollow"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /engagement/unfollow/{id} [delete]
func UnfollowUser(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authentication check
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Validate user ID
		followingID := c.Params("id")
		followingUUID, err := uuid.Parse(followingID)
		if err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid user ID"))
		}

		// Start transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Perform unfollow operation
		result := tx.Where("follower_id = ? AND following_id = ?", user.ID, followingUUID).
			Delete(&models.Follower{})

		if result.Error != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("database error unfollowing user: %w", result.Error))
		}

		// Check if any rows were actually deleted
		if result.RowsAffected == 0 {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusNotFound, "You were not following this user"))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "You unfollowed this user",
		}, nil)
	}
}

// GetUserFollowStats godoc
// @Summary Get follow stats for the logged-in user
// @Description Retrieves the number of followers and followings for the currently authenticated user
// @Tags Engagement
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /engagement/stats/session-user [get]
func GetUserFollowStatsForAuthUser(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authentication check - get user from context
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Get the user's UUID
		userUUID := user.ID

		type UserDetails struct {
			ID           uuid.UUID `json:"id"`
			Username     string    `json:"username"`
			FirstName    string    `json:"first_name,omitempty"`
			LastName     string    `json:"last_name,omitempty"`
			ProfileImage string    `json:"profile_image,omitempty"`
		}

		var stats struct {
			FollowerCount  int64         `json:"followers_count"`
			FollowingCount int64         `json:"following_count"`
			Followers      []UserDetails `json:"followers,omitempty"`
			Followings     []UserDetails `json:"followings,omitempty"`
		}

		// Create separate database connections for parallel queries
		dbFollowerCount := db.Session(&gorm.Session{})
		dbFollowingCount := db.Session(&gorm.Session{})
		dbFollowers := db.Session(&gorm.Session{})
		dbFollowings := db.Session(&gorm.Session{})

		// Use parallel execution with separate connections
		errChan := make(chan error, 4)
		var wg sync.WaitGroup
		wg.Add(4)

		// Get follower count
		go func() {
			defer wg.Done()
			if err := dbFollowerCount.Model(&models.Follower{}).
				Where("following_id = ?", userUUID).
				Count(&stats.FollowerCount).Error; err != nil {
				errChan <- fmt.Errorf("follower count error: %w", err)
			}
		}()

		// Get following count
		go func() {
			defer wg.Done()
			if err := dbFollowingCount.Model(&models.Follower{}).
				Where("follower_id = ?", userUUID).
				Count(&stats.FollowingCount).Error; err != nil {
				errChan <- fmt.Errorf("following count error: %w", err)
			}
		}()

		// Get followers with details
		go func() {
			defer wg.Done()
			if err := dbFollowers.Table("followers").
				Select(`
                    users.id,
                    users.username,
                    user_details.first_name,
                    user_details.last_name,
                    user_details.profile_image
                `).
				Joins("JOIN users ON followers.follower_id = users.id").
				Joins("LEFT JOIN user_details ON users.id = user_details.user_id").
				Where("followers.following_id = ?", userUUID).
				Scan(&stats.Followers).Error; err != nil {
				errChan <- fmt.Errorf("follower details error: %w", err)
			}
		}()

		// Get followings with details
		go func() {
			defer wg.Done()
			if err := dbFollowings.Table("followers").
				Select(`
                    users.id,
                    users.username,
                    user_details.first_name,
                    user_details.last_name,
                    user_details.profile_image
                `).
				Joins("JOIN users ON followers.following_id = users.id").
				Joins("LEFT JOIN user_details ON users.id = user_details.user_id").
				Where("followers.follower_id = ?", userUUID).
				Scan(&stats.Followings).Error; err != nil {
				errChan <- fmt.Errorf("following details error: %w", err)
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

		return responseHandler.HandleResponse(c, stats, nil)
	}
}

// GetUserFollowStats godoc
// @Summary Get follow stats of a user
// @Description Retrieves the number of followers and followings of a user, including follower and following lists
// @Tags Engagement
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /engagement/stats/{id} [get]
func GetUserFollowStats(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Validate and parse user ID
		userID := c.Params("id")
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid user ID format"))
		}

		type UserDetails struct {
			ID           uuid.UUID `json:"id"`
			Username     string    `json:"username"`
			FirstName    string    `json:"first_name,omitempty"`
			LastName     string    `json:"last_name,omitempty"`
			ProfileImage string    `json:"profile_image,omitempty"`
		}

		var stats struct {
			FollowerCount  int64         `json:"followers_count"`
			FollowingCount int64         `json:"following_count"`
			Followers      []UserDetails `json:"followers,omitempty"`
			Followings     []UserDetails `json:"followings,omitempty"`
		}

		// Use parallel execution for better performance
		errChan := make(chan error, 4)
		var wg sync.WaitGroup
		wg.Add(4)

		// Get follower count
		go func() {
			defer wg.Done()
			if err := db.Model(&models.Follower{}).
				Where("following_id = ?", userUUID).
				Count(&stats.FollowerCount).Error; err != nil {
				errChan <- fmt.Errorf("follower count error: %w", err)
			}
		}()

		// Get following count
		go func() {
			defer wg.Done()
			if err := db.Model(&models.Follower{}).
				Where("follower_id = ?", userUUID).
				Count(&stats.FollowingCount).Error; err != nil {
				errChan <- fmt.Errorf("following count error: %w", err)
			}
		}()

		// Get followers with details
		go func() {
			defer wg.Done()
			if err := db.Table("followers").
				Select(`
                    users.id,
                    users.username,
                    user_details.first_name,
                    user_details.last_name,
                    user_details.profile_image
                `).
				Joins("JOIN users ON followers.follower_id = users.id").
				Joins("LEFT JOIN user_details ON users.id = user_details.user_id").
				Where("followers.following_id = ?", userUUID).
				Scan(&stats.Followers).Error; err != nil {
				errChan <- fmt.Errorf("follower details error: %w", err)
			}
		}()

		// Get followings with details
		go func() {
			defer wg.Done()
			if err := db.Table("followers").
				Select(`
                    users.id,
                    users.username,
                    user_details.first_name,
                    user_details.last_name,
                    user_details.profile_image
                `).
				Joins("JOIN users ON followers.following_id = users.id").
				Joins("LEFT JOIN user_details ON users.id = user_details.user_id").
				Where("followers.follower_id = ?", userUUID).
				Scan(&stats.Followings).Error; err != nil {
				errChan <- fmt.Errorf("following details error: %w", err)
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

		return responseHandler.HandleResponse(c, stats, nil)
	}
}
