package collection

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	art "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	collectionModels "github.com/muga20/artsMarket/modules/artwork-management/models/collection"
	userModels "github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// CollectionImageRequest represents the request for updating collection images
type CollectionImageRequest struct {
	ImageType string `form:"type" validate:"required,oneof=cover primary"`
}

// UpdateCollectionRequest represents the request body for updating a collection
// @Description UpdateCollectionRequest defines the input for updating an existing collection
type UpdateCollectionRequest struct {
	Name        string `json:"name" example:"Updated Collection Name"`
	Description string `json:"description" example:"Updated collection description"`
	IsPublic    *bool  `json:"is_public" example:"true"`
}

// UpdateCollectionHandler handles updating collection details
// @Summary Update a collection
// @Description Updates basic information about a collection (name, description, visibility)
// @Tags Collections
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Collection ID" example("550e8400-e29b-41d4-a716-446655440000")
// @Param request body UpdateCollectionRequest true "Collection update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections/{id} [put]
func UpdateCollectionHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authenticated user
		user, ok := c.Locals("user").(userModels.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		collectionID := c.Params("id")
		if collectionID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Collection ID is required"))
		}

		// Parse request body
		var req UpdateCollectionRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))
		}

		// Verify collection exists and belongs to user
		var collection collectionModels.Collection
		if err := db.Where("id = ? AND user_id = ?", collectionID, user.ID).First(&collection).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Collection not found or you don't have permission"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to retrieve collection: %w", err))
		}

		// Update collection fields if they are provided in the request
		if req.Name != "" {
			collection.Name = req.Name
		}
		if req.Description != "" {
			collection.Description = req.Description
		}
		collection.UpdatedAt = time.Now()

		// Validate collection name is unique for this user
		var existingCollection collectionModels.Collection
		if err := db.Where("user_id = ? AND name = ? AND id != ?", user.ID, collection.Name, collectionID).First(&existingCollection).Error; err == nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "You already have a collection with this name"))
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to check for existing collections: %w", err))
		}

		// Save updated collection
		if err := db.Save(&collection).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update collection: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Collection updated successfully",
		}, nil)
	}
}

// UpdateCollectionStatusHandler handles updating collection status
// @Summary Update collection status
// @Description Change the status of a collection (draft/published/archived)
// @Tags Collections
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Collection ID"
// @Param status path string true "New status (publish|archive|draft)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections/{id}/status/{status} [put]
func UpdateCollectionStatusHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authenticated user
		user, ok := c.Locals("user").(userModels.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		collectionID := c.Params("id")
		if collectionID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Collection ID is required"))
		}

		status := c.Params("status")
		if status == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Status is required"))
		}

		// Validate and convert status parameter
		var newStatus collectionModels.CollectionStatus
		switch strings.ToLower(status) {
		case "publish":
			newStatus = collectionModels.PublishedStatus
		case "archive":
			newStatus = collectionModels.ArchivedStatus
		case "draft":
			newStatus = collectionModels.DraftStatus
		default:
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid status. Must be publish, archive, or draft"))
		}

		// Verify collection exists and belongs to user
		var collection collectionModels.Collection
		if err := db.Where("id = ? AND user_id = ?", collectionID, user.ID).First(&collection).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Collection not found or you don't have permission"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to retrieve collection: %w", err))
		}

		// Additional business rule: Can't publish empty collections
		if newStatus == collectionModels.PublishedStatus {
			var artworkCount int64
			if err := db.Model(&art.Artwork{}).Where("collection_id = ?", collectionID).Count(&artworkCount).Error; err != nil {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusInternalServerError, "Failed to verify collection contents"))
			}

			if artworkCount == 0 {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusBadRequest, "Cannot publish an empty collection"))
			}
		}

		// Update collection
		collection.Status = newStatus
		collection.UpdatedAt = time.Now()

		if err := db.Save(&collection).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update collection status: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Collection status updated successfully",
		}, nil)
	}
}
