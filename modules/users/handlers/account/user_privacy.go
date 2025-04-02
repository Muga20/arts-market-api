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

// GetPrivacySettings retrieves the current privacy settings of the authenticated user
// @Summary Get the authenticated user's privacy settings
// @Description Retrieves the current privacy settings for the authenticated user
// @Tags Account
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{} "Privacy Settings Data"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /account/privacy-settings [get]
func GetPrivacySettings(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)

		// Retrieve privacy settings for the authenticated user
		var privacySettings models.UserPrivacySetting
		if err := db.Where("user_id = ?", user.ID).First(&privacySettings).Error; err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{
				"success": false, "message": "Failed to retrieve privacy settings",
			}, err)
		}

		// Return privacy settings without unnecessary user data
		return responseHandler.HandleResponse(c, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"privacySettings": map[string]interface{}{
					"id":                    privacySettings.ID,
					"user_id":               privacySettings.UserID,
					"show_email":            privacySettings.ShowEmail,
					"show_phone":            privacySettings.ShowPhone,
					"show_location":         privacySettings.ShowLocation,
					"allow_follow_requests": privacySettings.AllowFollowRequests,
					"allow_messages_from":   privacySettings.AllowMessagesFrom,
				},
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
		user := c.Locals("user").(models.User)

		// Ensure privacy settings exist
		if err := models.CreateDefaultPrivacySetting(db, user.ID); err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{
				"success": false, "message": "Failed to initialize privacy settings",
			}, err)
		}

		// Retrieve existing privacy settings
		var privacySettings models.UserPrivacySetting
		if err := db.Where("user_id = ?", user.ID).First(&privacySettings).Error; err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{
				"success": false, "message": "Failed to retrieve privacy settings",
			}, err)
		}

		// Parse JSON body for updates
		var updateData PrivacySettingsUpdateRequest
		if err := c.BodyParser(&updateData); err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{
				"success": false, "message": "Invalid request body",
			}, err)
		}

		// Apply only provided fields
		updateFields := make(map[string]interface{})
		if updateData.ShowEmail != nil {
			updateFields["show_email"] = *updateData.ShowEmail
		}
		if updateData.ShowPhone != nil {
			updateFields["show_phone"] = *updateData.ShowPhone
		}
		if updateData.ShowLocation != nil {
			updateFields["show_location"] = *updateData.ShowLocation
		}
		if updateData.AllowFollowRequests != nil {
			updateFields["allow_follow_requests"] = *updateData.AllowFollowRequests
		}
		if updateData.AllowMessagesFrom != "" {
			updateFields["allow_messages_from"] = updateData.AllowMessagesFrom
		}

		// Perform update if there's something to change
		if len(updateFields) > 0 {
			if err := db.Model(&privacySettings).Updates(updateFields).Error; err != nil {
				return responseHandler.HandleResponse(c, map[string]interface{}{
					"success": false, "message": "Failed to update privacy settings",
				}, fmt.Errorf("failed to update privacy settings"))
			}
		}

		return responseHandler.HandleResponse(c, map[string]interface{}{
			"success": true, "message": "Privacy settings updated successfully",
		}, nil)
	}
}
