package account

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// PrivacySettingsUpdateRequest defines allowed update fields
type PrivacySettingsUpdateRequest struct {
	ShowEmail           *bool  `json:"show_email,omitempty"`
	ShowPhone           *bool  `json:"show_phone,omitempty"`
	ShowLocation        *bool  `json:"show_location,omitempty"`
	AllowFollowRequests *bool  `json:"allow_follow_requests,omitempty"`
	AllowMessagesFrom   string `json:"allow_messages_from,omitempty"`
}

// GetPrivacySettings godoc
// @Summary Get user privacy settings
// @Description Retrieves privacy settings for the authenticated user
// @Tags Account
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /account/privacy-settings [get]
func GetPrivacySettings(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authentication check
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		var privacySettings models.UserPrivacySetting
		err := db.Where("user_id = ?", user.ID).First(&privacySettings).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Return empty object when no settings exist
				return responseHandler.HandleResponse(c, fiber.Map{
					"privacy_settings": fiber.Map{},
				}, nil)
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to retrieve privacy settings: %w", err))
		}

		// Return existing settings
		return responseHandler.HandleResponse(c, fiber.Map{
			"privacy_settings": fiber.Map{
				"id":                    privacySettings.ID,
				"user_id":               privacySettings.UserID,
				"show_email":            privacySettings.ShowEmail,
				"show_phone":            privacySettings.ShowPhone,
				"show_location":         privacySettings.ShowLocation,
				"allow_follow_requests": privacySettings.AllowFollowRequests,
				"allow_messages_from":   privacySettings.AllowMessagesFrom,
			},
		}, nil)
	}
}

// UpdatePrivacySettings updates specific privacy settings
// @Summary Update the authenticated user's privacy settings
// @Description Allows users to update visibility and message preferences
// @Tags Account
// @Accept  json
// @Produce  json
// @Param   body  body  PrivacySettingsUpdateRequest  true  "Privacy settings update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /account/privacy-settings [put]
func UpdatePrivacySettings(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authentication check
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Ensure privacy settings exist
		if err := models.CreateDefaultPrivacySetting(db, user.ID); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to initialize privacy settings: %w", err))
		}

		// Parse request body
		var updateData struct {
			ShowEmail           *bool  `json:"show_email"`
			ShowPhone           *bool  `json:"show_phone"`
			ShowLocation        *bool  `json:"show_location"`
			AllowFollowRequests *bool  `json:"allow_follow_requests"`
			AllowMessagesFrom   string `json:"allow_messages_from"`
		}

		if err := c.BodyParser(&updateData); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))
		}

		// Validate AllowMessagesFrom if provided
		if updateData.AllowMessagesFrom != "" &&
			updateData.AllowMessagesFrom != "everyone" &&
			updateData.AllowMessagesFrom != "following" &&
			updateData.AllowMessagesFrom != "none" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid value for allow_messages_from"))
		}

		// Build update fields
		updates := make(map[string]interface{})
		if updateData.ShowEmail != nil {
			updates["show_email"] = *updateData.ShowEmail
		}
		if updateData.ShowPhone != nil {
			updates["show_phone"] = *updateData.ShowPhone
		}
		if updateData.ShowLocation != nil {
			updates["show_location"] = *updateData.ShowLocation
		}
		if updateData.AllowFollowRequests != nil {
			updates["allow_follow_requests"] = *updateData.AllowFollowRequests
		}
		if updateData.AllowMessagesFrom != "" {
			updates["allow_messages_from"] = updateData.AllowMessagesFrom
		}

		// Check if any updates were provided
		if len(updates) == 0 {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "No valid fields provided for update"))
		}

		// Perform the update
		result := db.Model(&models.UserPrivacySetting{}).
			Where("user_id = ?", user.ID).
			Updates(updates)

		if result.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update privacy settings: %w", result.Error))
		}

		if result.RowsAffected == 0 {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusNotFound, "Privacy settings not found"))
		}

		// Return simple success message
		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Privacy settings updated successfully",
		}, nil)
	}
}
