package collection

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	collection "github.com/muga20/artsMarket/modules/artwork-management/models/collection"
	models "github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

type CreateCollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// CollectionStatus represents the possible statuses of a collection
type CollectionStatus string

const (
	DraftStatus     CollectionStatus = "draft"
	PublishedStatus CollectionStatus = "published"
	ArchivedStatus  CollectionStatus = "archived"
)

// CreateCollectionHandler handles creating a new collection
// @Summary Create a new collection
// @Description Creates a new art collection for the authenticated user
// @Tags Collections
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body CreateCollectionRequest true "Collection creation data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections [post]
func CreateCollectionHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authenticated user
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		var req CreateCollectionRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		// Basic validation - just check required fields
		if req.Name == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Collection name is required"))
		}

		// Check if collection name already exists for this user
		var existingCollection collection.Collection
		if err := db.Where("user_id = ? AND name = ?", user.ID, req.Name).First(&existingCollection).Error; err == nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "You already have a collection with this name"))
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to check collection existence: %w", err))
		}

		newCollection := collection.Collection{
			ID:          uuid.New(),
			UserID:      user.ID,
			Name:        req.Name,
			Description: req.Description,
			Status:      collection.CollectionStatus(DraftStatus),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := db.Create(&newCollection).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to create collection: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Collection created successfully",
		}, nil)
	}
}
