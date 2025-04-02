package account

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// SocialLinkRequest struct for create and update operations
type SocialLinkRequest struct {
	Platform string `json:"platform" validate:"required"`
	Link     string `json:"link" validate:"required,url"`
}

type SocialLinkResponse struct {
	ID       uuid.UUID `json:"id"`
	UserID   uuid.UUID `json:"user_id"`
	Platform string    `json:"platform"`
	Link     string    `json:"link"`
}

// CreateSocialLink creates a new social link for the authenticated user
// @Summary Create a new social link
// @Description Creates a new social link for the authenticated user
// @Tags Account
// @Accept  json
// @Produce  json
// @Param   body  body  SocialLinkRequest  true  "Social Link Data"
// @Success 200 {object} map[string]interface{}  "Success Response"
// @Failure 400 {object} map[string]interface{}  "Bad Request"
// @Failure 401 {object} map[string]string  "Unauthorized"
// @Failure 500 {object} map[string]interface{}  "Internal Server Error"
// @Router /account/social-links [post]
func CreateSocialLink(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)

		var req SocialLinkRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{
				"success": false, "message": "Invalid request body",
			}, err)
		}

		socialLink := models.SocialLink{
			UserID:   user.ID,
			Platform: req.Platform,
			Link:     req.Link,
		}

		if err := db.Create(&socialLink).Error; err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{
				"success": false, "message": "Failed to create social link",
			}, err)
		}

		return responseHandler.HandleResponse(c, map[string]interface{}{
			"success": true, "message": "Social link created successfully",
		}, nil)
	}
}

// GetSocialLinks retrieves all social links for the authenticated user
// @Summary Retrieve all social links
// @Tags Account
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{} "List of Social Links"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /account/social-links [get]
func GetSocialLinks(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)
		var socialLinks []models.SocialLink
		if err := db.Where("user_id = ?", user.ID).
			Select("id, user_id, platform, link").
			Find(&socialLinks).Error; err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{
				"success": false, "message": "Failed to retrieve social links",
			}, err)
		}

		// Prepare response by excluding the 'User' field
		var responseLinks []SocialLinkResponse
		for _, link := range socialLinks {
			responseLinks = append(responseLinks, SocialLinkResponse{
				ID:       link.ID,
				UserID:   link.UserID,
				Platform: link.Platform,
				Link:     link.Link,
			})
		}

		return responseHandler.HandleResponse(c, map[string]interface{}{
			"success":     true,
			"socialLinks": responseLinks, // Return without 'User'
		}, nil)
	}
}

// UpdateSocialLink updates an existing social link
// @Summary Update an existing social link
// @Tags Account
// @Accept  json
// @Produce  json
// @Param   id   path    string  true  "Social Link ID"
// @Param   body body    SocialLinkRequest true "Updated Social Link Data"
// @Success 200 {object} map[string]interface{} "Success Response"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /account/social-links/{id} [put]
func UpdateSocialLink(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)
		socialLinkID := c.Params("id")
		var socialLink models.SocialLink
		if err := db.Where("id = ? AND user_id = ?", socialLinkID, user.ID).First(&socialLink).Error; err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{"success": false, "message": "Social link not found"}, err)
		}
		var req SocialLinkRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{"success": false, "message": "Invalid request body"}, err)
		}
		updates := map[string]interface{}{}
		if req.Platform != "" {
			updates["platform"] = req.Platform
		}
		if req.Link != "" {
			updates["link"] = req.Link
		}
		if err := db.Model(&socialLink).Updates(updates).Error; err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{"success": false, "message": "Failed to update social link"}, err)
		}
		return responseHandler.HandleResponse(c, map[string]interface{}{"success": true, "message": "Social link updated successfully"}, nil)
	}
}

// DeleteSocialLink deletes a social link by ID
// @Summary Delete a social link
// @Tags Account
// @Accept  json
// @Produce  json
// @Param   id   path    string  true  "Social Link ID"
// @Success 200 {object} map[string]interface{} "Success Response"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /account/social-links/{id} [delete]
func DeleteSocialLink(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)
		socialLinkID := c.Params("id")
		var socialLink models.SocialLink
		if err := db.Where("id = ? AND user_id = ?", socialLinkID, user.ID).First(&socialLink).Error; err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{"success": false, "message": "Social link not found"}, err)
		}
		if err := db.Delete(&socialLink).Error; err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{"success": false, "message": "Failed to delete social link"}, err)
		}
		return responseHandler.HandleResponse(c, map[string]interface{}{"success": true, "message": "Social link deleted successfully"}, nil)
	}
}
