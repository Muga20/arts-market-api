package account

import (
	"fmt"
	"strings"

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
		// Authentication check with proper type assertion
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Parse request body
		var req SocialLinkRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))
		}

		// Basic validation (add more specific validation as needed)
		if req.Platform == "" || req.Link == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Platform and link are required"))
		}

		// Create social link
		socialLink := models.SocialLink{
			UserID:   user.ID,
			Platform: req.Platform,
			Link:     req.Link,
		}

		if err := db.Create(&socialLink).Error; err != nil {
			// Handle duplicate entry or other DB errors
			if strings.Contains(err.Error(), "duplicate") {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusConflict, "Social link for this platform already exists"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to create social link: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Social link created successfully",
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
		// Authentication check with proper type assertion
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Retrieve social links
		var socialLinks []models.SocialLink
		if err := db.Where("user_id = ?", user.ID).
			Select("id, user_id, platform, link").
			Find(&socialLinks).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to retrieve social links: %w", err))
		}

		// Prepare response
		responseLinks := make([]SocialLinkResponse, 0, len(socialLinks))
		for _, link := range socialLinks {
			responseLinks = append(responseLinks, SocialLinkResponse{
				ID:       link.ID,
				UserID:   link.UserID,
				Platform: link.Platform,
				Link:     link.Link,
			})
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"social_links": responseLinks,
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
		// Authentication check
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Get social link ID
		socialLinkID := c.Params("id")
		if _, err := uuid.Parse(socialLinkID); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid social link ID"))
		}

		// Find existing social link
		var socialLink models.SocialLink
		if err := db.Where("id = ? AND user_id = ?", socialLinkID, user.ID).
			First(&socialLink).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Social link not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to find social link: %w", err))
		}

		// Parse request body
		var req SocialLinkRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))
		}

		// Validate at least one field is provided
		if req.Platform == "" && req.Link == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "No fields provided for update"))
		}

		// Prepare updates
		updates := make(map[string]interface{})
		if req.Platform != "" {
			updates["platform"] = req.Platform
		}
		if req.Link != "" {
			updates["link"] = req.Link
		}

		// Perform update
		if err := db.Model(&socialLink).Updates(updates).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update social link: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Social link updated successfully",
		}, nil)
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
		// Authentication check with type assertion
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Validate social link ID
		socialLinkID := c.Params("id")
		if _, err := uuid.Parse(socialLinkID); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid social link ID"))
		}

		// Verify social link exists and belongs to user
		var socialLink models.SocialLink
		if err := db.Where("id = ? AND user_id = ?", socialLinkID, user.ID).
			First(&socialLink).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Social link not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to find social link: %w", err))
		}

		// Delete the social link
		if err := db.Delete(&socialLink).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to delete social link: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Social link deleted successfully",
		}, nil)
	}
}
