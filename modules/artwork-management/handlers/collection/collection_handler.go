package collection

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	art "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	collectionModels "github.com/muga20/artsMarket/modules/artwork-management/models/collection"
	models "github.com/muga20/artsMarket/modules/users/models"

	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// GetAllCollectionsHandler handles fetching all collections
// @Summary Get all user collections
// @Description Retrieves all collections belonging to the authenticated user
// @Tags Collections
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Returns list of collections and count"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections [get]
func GetAllCollectionsHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)

		// Use Select to specify exactly which fields we want
		var collections []struct {
			ID              string    `json:"id"`
			UserID          string    `json:"user_id"`
			Name            string    `json:"name"`
			Description     string    `json:"description"`
			Status          string    `json:"status"`
			CreatedAt       time.Time `json:"created_at"`
			UpdatedAt       time.Time `json:"updated_at"`
			CoverImageURL   string    `json:"cover_image_url"`
			PrimaryImageURL string    `json:"primary_image_url"`
		}

		if err := db.Table("collections").
			Select("id, user_id, name, description, status, created_at, updated_at, cover_image_url, primary_image_url").
			Where("user_id = ?", user.ID).
			Find(&collections).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve collections"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"collections": collections,
			"count":       len(collections),
		}, nil)
	}
}

// GetCollectionByIDHandler handles fetching a collection by ID
// @Summary Get collection by ID
// @Description Retrieves a collection by its ID. Collections can be viewed if public or if owned by the authenticated user.
// @Tags Collections
// @Accept json
// @Produce json
// @Param id path string true "Collection ID" example("550e8400-e29b-41d4-a716-446655440000")
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections/{id} [get]
func GetCollectionByIDHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(models.User)
		collectionID := c.Params("id")
		var collection collectionModels.Collection

		// Load collection with basic user info
		if err := db.
			Preload("User").
			Where("id = ?", collectionID).
			First(&collection).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Collection not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve collection"))
		}

		// Check if requester is the owner
		isOwner := ok && collection.UserID == user.ID

		// Access control
		if !isOwner && collection.Status != collectionModels.PublishedStatus {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusForbidden, "This collection is not published"))
		}

		// Fetch user details separately
		var userDetail models.UserDetail
		if err := db.Where("user_id = ?", collection.UserID).First(&userDetail).Error; err != nil {
			// If no user details found, we'll proceed with empty values
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve user details"))
			}
		}

		// Prepare response
		response := fiber.Map{
			"data": fiber.Map{
				"id":                collection.ID,
				"user_id":           collection.UserID,
				"name":              collection.Name,
				"description":       collection.Description,
				"status":            collection.Status,
				"cover_image_url":   collection.CoverImageURL,
				"primary_image_url": collection.PrimaryImageURL,
				"created_at":        collection.CreatedAt,

				"user": fiber.Map{
					"id":            collection.User.ID,
					"username":      collection.User.Username,
					"profile_image": userDetail.ProfileImage,
					"first_name":    userDetail.FirstName,
					"last_name":     userDetail.LastName,
				},
			},
		}

		return responseHandler.HandleResponse(c, response, nil)
	}
}

// DeleteCollectionHandler handles deleting a collection
// @Summary Delete a collection
// @Description Deletes a user's collection if it's empty. Requires authentication and ownership of the collection.
// @Tags Collections
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Collection ID" example("550e8400-e29b-41d4-a716-446655440000")
// @Success 200 {object} map[string]string "Returns success message"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections/{id} [delete]
func DeleteCollectionHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)
		collectionID := c.Params("id")

		var collection collectionModels.Collection
		if err := db.Where("id = ? AND user_id = ?", collectionID, user.ID).First(&collection).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Collection not found or you don't have permission"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve collection"))
		}

		// Check if collection is empty
		var count int64
		if err := db.Model(&art.Artwork{}).Where("collection_id = ?", collectionID).Count(&count).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to check collection contents"))
		}

		if count > 0 {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "Cannot delete collection - it contains artworks"))
		}

		if err := db.Delete(&collection).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to delete collection"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Collection deleted successfully",
		}, nil)
	}
}
