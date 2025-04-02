package engagement

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/muga20/artsMarket/modules/notifications/services"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// MaxFollowsPerDay limits how many users someone can follow per day
const MaxFollowsPerDay = 10

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
		// Get the authenticated user
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "User not authenticated"}, fmt.Errorf("unauthenticated request"))
		}

		// Extract following ID from the URL params
		followingID := c.Params("id")
		if user.ID.String() == followingID {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "You cannot follow yourself"}, nil)
		}

		// Parse the following user UUID
		followingUUID, err := uuid.Parse(followingID)
		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Invalid user ID"}, err)
		}

		// Check if the user is already following the target user
		var count int64
		err = db.Model(&models.Follower{}).Where("follower_id = ? AND following_id = ?", user.ID, followingUUID).Count(&count).Error
		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Database error"}, err)
		}
		if count > 0 {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "You already follow this user"}, nil)
		}

		// Check if the user has exceeded the daily follow limit
		var followCount int64
		startOfDay := time.Now().Truncate(24 * time.Hour)
		err = db.Model(&models.Follower{}).Where("follower_id = ? AND created_at >= ?", user.ID, startOfDay).Count(&followCount).Error
		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Database error"}, err)
		}
		if followCount >= MaxFollowsPerDay {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "You have reached your daily follow limit"}, nil)
		}

		// Create the follow relationship
		newFollow := models.Follower{
			FollowerID:  user.ID,
			FollowingID: followingUUID,
		}
		if err := db.Create(&newFollow).Error; err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Could not follow user"}, err)
		}

		// Send a notification to the user who is being followed
		notificationService := services.NewNotificationService(responseHandler)
		err = notificationService.EnqueueNotification(
			followingUUID.String(), // Target user (the one being followed)
			user.ID.String(),       // Follower (the one who followed)
			"follow",               // Notification type
			"user followed you",    // Message
			"user",                 // Entity type
			followingUUID.String(), // Entity ID (user being followed)
		)
		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Notification sending failed"}, err)
		}

		return responseHandler.HandleResponse(c, fiber.Map{"success": true, "message": "You are now following this user"}, nil)
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
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "User not authenticated"}, fmt.Errorf("unauthenticated request"))
		}

		followingID := c.Params("id")
		followingUUID, err := uuid.Parse(followingID)
		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Invalid user ID"}, err)
		}

		if err := db.Where("follower_id = ? AND following_id = ?", user.ID, followingUUID).Delete(&models.Follower{}).Error; err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Could not unfollow user"}, err)
		}

		return responseHandler.HandleResponse(c, fiber.Map{"success": true, "message": "You have unfollowed this user"}, nil)
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
		// Get logged-in user
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "User not authenticated"}, fmt.Errorf("unauthenticated request"))
		}

		var followerCount, followingCount int64
		var followers []models.Follower
		var followings []models.Follower

		// Count and retrieve followers (people following the user)
		err := db.Model(&models.Follower{}).
			Where("following_id = ?", user.ID).
			Count(&followerCount).Find(&followers).Error
		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Error retrieving followers"}, err)
		}

		// Count and retrieve followings (people the user follows)
		err = db.Model(&models.Follower{}).
			Where("follower_id = ?", user.ID).
			Count(&followingCount).Find(&followings).Error
		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Error retrieving followings"}, err)
		}

		// Struct to store follower/following user details
		type UserDetails struct {
			ID        uuid.UUID `json:"id"`
			Username  string    `json:"username"`
			FirstName string    `json:"first_name"`
			LastName  string    `json:"last_name"`
		}

		var followersDetails []UserDetails
		var followingsDetails []UserDetails

		// Retrieve detailed information for followers
		for _, follower := range followers {
			var details UserDetails
			err := db.Table("users").
				Select("users.id, users.username, user_details.first_name, user_details.last_name").
				Joins("LEFT JOIN user_details ON users.id = user_details.user_id").
				Where("users.id = ?", follower.FollowerID).
				First(&details).Error
			if err != nil {
				return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Error retrieving follower details"}, err)
			}
			followersDetails = append(followersDetails, details)
		}

		// Retrieve detailed information for followings
		for _, following := range followings {
			var details UserDetails
			err := db.Table("users").
				Select("users.id, users.username, user_details.first_name, user_details.last_name").
				Joins("LEFT JOIN user_details ON users.id = user_details.user_id").
				Where("users.id = ?", following.FollowingID).
				First(&details).Error
			if err != nil {
				return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Error retrieving following details"}, err)
			}
			followingsDetails = append(followingsDetails, details)
		}

		// Prepare the response with follower and following details
		return responseHandler.HandleResponse(c, fiber.Map{
			"followers_count": followerCount,
			"following_count": followingCount,
			"followers":       followersDetails,
			"followings":      followingsDetails,
		}, nil)
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
		userID := c.Params("id")
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Invalid user ID"}, err)
		}

		var followerCount, followingCount int64
		var followers []models.Follower
		var followings []models.Follower

		// Count and retrieve followers (people following the user)
		err = db.Model(&models.Follower{}).
			Where("following_id = ?", userUUID).
			Count(&followerCount).Find(&followers).Error
		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Error retrieving followers"}, err)
		}

		// Count and retrieve followings (people the user follows)
		err = db.Model(&models.Follower{}).
			Where("follower_id = ?", userUUID).
			Count(&followingCount).Find(&followings).Error
		if err != nil {
			return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Error retrieving followings"}, err)
		}

		// Struct to store follower/following user details
		type UserDetails struct {
			ID        uuid.UUID `json:"id"`
			Username  string    `json:"username"`
			FirstName string    `json:"first_name"`
			LastName  string    `json:"last_name"`
		}

		var followersDetails []UserDetails
		var followingsDetails []UserDetails

		// Retrieve detailed information for followers
		for _, follower := range followers {
			var details UserDetails
			err := db.Table("users").
				Select("users.id, users.username, user_details.first_name, user_details.last_name, user_details.profile_image").
				Joins("LEFT JOIN user_details ON users.id = user_details.user_id").
				Where("users.id = ?", follower.FollowerID).
				First(&details).Error
			if err != nil {
				return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Error retrieving follower details"}, err)
			}
			followersDetails = append(followersDetails, details)
		}

		// Retrieve detailed information for followings
		for _, following := range followings {
			var details UserDetails
			err := db.Table("users").
				Select("users.id, users.username, user_details.first_name, user_details.last_name, user_details.profile_image").
				Joins("LEFT JOIN user_details ON users.id = user_details.user_id").
				Where("users.id = ?", following.FollowingID).
				First(&details).Error
			if err != nil {
				return responseHandler.HandleResponse(c, fiber.Map{"success": false, "message": "Error retrieving following details"}, err)
			}
			followingsDetails = append(followingsDetails, details)
		}

		// Prepare the response with follower and following details
		return responseHandler.HandleResponse(c, fiber.Map{
			"followers_count": followerCount,
			"following_count": followingCount,
			"followers":       followersDetails,
			"followings":      followingsDetails,
		}, nil)
	}
}
