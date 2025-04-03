package public

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// GetUserProfileHandler returns a user's profile information
// @Summary Get user profile
// @Description Retrieves public profile information for a user by username, email, or phone number. Checks for blocked status and profile visibility.
// @Tags Users
// @Accept json
// @Produce json
// @Param identifier path string true "User identifier (username, email, or phone number)"
// @Success 200 {object} map[string]interface{} "Successful response with user profile"
// @Success 200 {object} map[string]interface{} "Private profile response (shows basic info only)"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Profile is private"
// @Failure 404 {object} map[string]string "User not found (or blocked)"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /profile/{identifier} [get]
func GetUserProfileHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		identifier := c.Params("identifier")
		if identifier == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Identifier is required"))
		}

		// Get requesting user from context
		requester, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// First get the target user with active/deleted checks
		var targetUser struct {
			ID       string `gorm:"type:char(36)"`
			IsActive bool
			Deleted  bool
		}

		// Fix email search by using case-insensitive comparison
		if err := db.Table("users").
			Select("id, is_active, deleted_at IS NOT NULL as deleted").
			Where("(LOWER(username) = LOWER(?) OR LOWER(email) = LOWER(?) OR phone_number = ?) AND is_active = true AND deleted_at IS NULL",
				identifier, identifier, identifier).
			Scan(&targetUser).Error; err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "User not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to find target user: %w", err))
		}

		// If user is inactive or deleted
		if !targetUser.IsActive || targetUser.Deleted {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusNotFound, "User not found"))
		}

		// Parse the target user ID
		targetUserID, err := uuid.Parse(targetUser.ID)
		if err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("invalid user ID format: %w", err))
		}

		// Check if either user has blocked the other
		var blockCount int64
		err = db.Model(&models.BlockedUser{}).
			Where("(user_id = ? AND blocked_user_id = ?) OR (user_id = ? AND blocked_user_id = ?)",
				targetUserID, requester.ID,
				requester.ID, targetUserID).
			Count(&blockCount).Error

		if err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to check block status: %w", err))
		}

		if blockCount > 0 {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusNotFound, "User not found"))
		}

		// Get profile visibility and basic info with active/deleted check
		var profile struct {
			User struct {
				ID              uuid.UUID `json:"id"`
				Username        string    `json:"username"`
				MemberSince     time.Time `json:"member_since"`
				IsProfilePublic bool      `json:"-"`
			}
		}

		if err := db.Table("users").
			Select("users.id, users.username, users.created_at as member_since, COALESCE(user_details.is_profile_public, false) as is_profile_public").
			Joins("LEFT JOIN user_details ON users.id = user_details.user_id").
			Where("(LOWER(users.username) = LOWER(?) OR LOWER(users.email) = LOWER(?) OR users.phone_number = ?) AND users.is_active = true AND users.deleted_at IS NULL",
				identifier, identifier, identifier).
			Scan(&profile.User).Error; err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "User not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to get user profile: %w", err))
		}

		// If profile isn't public and requester isn't the owner
		if !profile.User.IsProfilePublic && profile.User.ID != requester.ID {
			return responseHandler.HandleResponse(c, fiber.Map{
				"user": fiber.Map{
					"id":           profile.User.ID,
					"username":     profile.User.Username,
					"member_since": profile.User.MemberSince,
				},
				"message": "Profile is private",
			}, nil)
		}

		// Get additional profile data only if visible
		var extended struct {
			Details struct {
				FirstName         string  `json:"first_name"`
				LastName          string  `json:"last_name"`
				ProfileImage      string  `json:"profile_image"`
				CoverImage        string  `json:"cover_image"`
				About             *string `json:"about,omitempty"`
				Gender            *string `json:"gender,omitempty"`
				PreferredPronouns *string `json:"preferred_pronouns,omitempty"`
			} `json:"details"`
			Stats struct {
				Followers int64 `json:"followers"`
				Following int64 `json:"following"`
			} `json:"stats"`
		}

		// Get user details in parallel with stats
		errChan := make(chan error, 2)
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			if err := db.Table("user_details").
				Select("first_name, last_name, profile_image, cover_image, about_the_user as about, gender, preferred_pronouns").
				Where("user_id = ?", profile.User.ID).
				Scan(&extended.Details).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				errChan <- fmt.Errorf("failed to get user details: %w", err)
			}
		}()

		go func() {
			defer wg.Done()
			var counts struct {
				Followers int64
				Following int64
			}
			if err := db.Raw(`
                SELECT 
                    (SELECT COUNT(*) FROM followers WHERE following_id = ?) as followers,
                    (SELECT COUNT(*) FROM followers WHERE follower_id = ?) as following
            `, profile.User.ID, profile.User.ID).Scan(&counts).Error; err != nil {
				errChan <- fmt.Errorf("failed to get follower stats: %w", err)
				return
			}
			extended.Stats.Followers = counts.Followers
			extended.Stats.Following = counts.Following
		}()

		wg.Wait()
		close(errChan)

		for err := range errChan {
			if err != nil {
				return responseHandler.HandleResponse(c, nil, err)
			}
		}

		// Combine all profile data
		fullProfile := fiber.Map{
			"user": fiber.Map{
				"id":           profile.User.ID,
				"username":     profile.User.Username,
				"member_since": profile.User.MemberSince,
			},
			"details": extended.Details,
			"stats":   extended.Stats,
		}

		return responseHandler.HandleResponse(c, fullProfile, nil)
	}
}
