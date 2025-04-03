package tags

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	models "github.com/muga20/artsMarket/modules/artwork-management/models/tags"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// CreateTagRequest represents the request body for creating a tag
type CreateTagRequest struct {
	TagName string `json:"tag_name" validate:"required"`
}

// UpdateTagRequest represents the request body for updating a tag
type UpdateTagRequest struct {
	TagName string `json:"tag_name"`
}

// CreateTagHandler handles creating a new tag
// @Summary Create a new tag
// @Description Create a new tag with name and active status
// @Tags Tags
// @Accept  json
// @Produce  json
// @Param request body CreateTagRequest true "Tag creation payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tags [post]
func CreateTagHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateTagRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		tag := models.Tag{
			ID:        uuid.New(),
			TagName:   req.TagName,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := db.Create(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusConflict, "Tag with this name already exists"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to create tag: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Tag created successfully",
		}, nil)
	}
}

// GetAllTagsHandler handles fetching all tags
// @Summary Get all tags
// @Description Retrieve a list of all tags
// @Tags Tags
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /tags [get]
func GetAllTagsHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tags []models.Tag
		if err := db.Find(&tags).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve tags"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"tags":  tags,
			"count": len(tags),
		}, nil)
	}
}

// GetTagByIDHandler handles fetching a tag by ID
// @Summary Get a tag by ID
// @Description Retrieve a tag's details using its ID
// @Tags Tags
// @Accept  json
// @Produce  json
// @Param id path string true "Tag ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tags/{id} [get]
func GetTagByIDHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tagID := c.Params("id")
		if tagID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Tag ID is required"))
		}

		var tag models.Tag
		if err := db.Where("id = ?", tagID).First(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Tag not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve tag"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"data": tag,
		}, nil)
	}
}

// UpdateTagHandler handles updating an existing tag
// @Summary Update an existing tag
// @Description Update tag details (name)
// @Tags Tags
// @Accept  json
// @Produce  json
// @Param id path string true "Tag ID"
// @Param request body UpdateTagRequest true "Tag update payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tags/{id} [put]
func UpdateTagHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tagID := c.Params("id")
		if tagID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Tag ID is required"))
		}

		var req UpdateTagRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		var tag models.Tag
		if err := db.Where("id = ?", tagID).First(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Tag not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve tag"))
		}

		// Update tag name if provided
		if req.TagName != "" {
			tag.TagName = req.TagName
		}
		tag.UpdatedAt = time.Now()

		if err := db.Save(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusConflict, "Tag with this name already exists"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to update tag"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Tag updated successfully",
		}, nil)
	}
}

// ToggleTagActiveHandler handles toggling a tag's active status
// @Summary Toggle tag active status
// @Description Toggle a tag's active status (activate/deactivate)
// @Tags Tags
// @Accept  json
// @Produce  json
// @Param id path string true "Tag ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tags/{id}/toggle-active [put]
func ToggleTagActiveHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tagID := c.Params("id")
		if tagID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Tag ID is required"))
		}

		var tag models.Tag
		if err := db.Where("id = ?", tagID).First(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Tag not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve tag"))
		}

		// Toggle active status
		tag.IsActive = !tag.IsActive
		tag.UpdatedAt = time.Now()

		if err := db.Save(&tag).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to update tag status"))
		}

		action := "activated"
		if !tag.IsActive {
			action = "deactivated"
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": fmt.Sprintf("Tag %s successfully", action),
		}, nil)
	}
}

// DeleteTagHandler handles permanently deleting a tag
// @Summary Delete a tag
// @Description Permanently delete a tag (use with caution)
// @Tags Tags
// @Accept  json
// @Produce  json
// @Param id path string true "Tag ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tags/{id} [delete]
func DeleteTagHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tagID := c.Params("id")
		if tagID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Tag ID is required"))
		}

		// First check if tag exists
		var tag models.Tag
		if err := db.Where("id = ?", tagID).First(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Tag not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve tag"))
		}

		// Check if tag is used in any artwork before deleting
		var count int64
		if err := db.Model(&models.ArtworkTag{}).Where("tag_id = ?", tagID).Count(&count).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to check tag usage"))
		}

		if count > 0 {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "Cannot delete tag - it's being used in artworks"))
		}

		// Perform the delete
		if err := db.Delete(&tag).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to delete tag"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Tag deleted successfully",
		}, nil)
	}
}
