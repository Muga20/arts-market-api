package attributes

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	models "github.com/muga20/artsMarket/modules/artwork-management/models/arttributes"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// CreateAttributeRequest represents the request body for creating an attribute
type CreateAttributeRequest struct {
	AttributeName string `json:"attribute_name" validate:"required"`
	Type          string `json:"type" validate:"required,oneof=string integer float boolean"`
}

// UpdateAttributeRequest represents the request body for updating an attribute
type UpdateAttributeRequest struct {
	AttributeName string `json:"attribute_name"`
	Type          string `json:"type" validate:"omitempty,oneof=string integer float boolean"`
}

// CreateAttributeHandler handles creating a new attribute
// @Summary Create a new attribute
// @Description Create a new attribute with name and type
// @Tags Attributes
// @Accept  json
// @Produce  json
// @Param request body CreateAttributeRequest true "Attribute creation payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /attributes [post]
func CreateAttributeHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateAttributeRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		// First check if attribute name already exists
		var existingAttr models.Attribute
		if err := db.Where("attribute_name = ?", req.AttributeName).First(&existingAttr).Error; err == nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "Attribute with this name already exists"))
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to check attribute existence: %w", err))
		}

		attribute := models.Attribute{
			ID:            uuid.New(),
			AttributeName: req.AttributeName,
			Type:          req.Type,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := db.Create(&attribute).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to create attribute: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Attribute created successfully",
			"data":    attribute,
		}, nil)
	}
}

// GetAllAttributesHandler handles fetching all attributes
// @Summary Get all attributes
// @Description Retrieve a list of all attributes
// @Tags Attributes
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /attributes [get]
func GetAllAttributesHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var attributes []models.Attribute
		if err := db.Find(&attributes).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve attributes"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"attributes": attributes,
			"count":      len(attributes),
		}, nil)
	}
}

// GetAttributeByIDHandler handles fetching an attribute by ID
// @Summary Get an attribute by ID
// @Description Retrieve an attribute's details using its ID
// @Tags Attributes
// @Accept  json
// @Produce  json
// @Param id path string true "Attribute ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /attributes/{id} [get]
func GetAttributeByIDHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		attributeID := c.Params("id")
		if attributeID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Attribute ID is required"))
		}

		var attribute models.Attribute
		if err := db.Where("id = ?", attributeID).First(&attribute).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Attribute not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve attribute"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"data": attribute,
		}, nil)
	}
}

// UpdateAttributeHandler handles updating an existing attribute
// @Summary Update an existing attribute
// @Description Update attribute details (name, type)
// @Tags Attributes
// @Accept  json
// @Produce  json
// @Param id path string true "Attribute ID"
// @Param request body UpdateAttributeRequest true "Attribute update payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /attributes/{id} [put]
func UpdateAttributeHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		attributeID := c.Params("id")
		if attributeID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Attribute ID is required"))
		}

		var req UpdateAttributeRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		var attribute models.Attribute
		if err := db.Where("id = ?", attributeID).First(&attribute).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Attribute not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve attribute"))
		}

		// Check if new name already exists (if being changed)
		if req.AttributeName != "" && req.AttributeName != attribute.AttributeName {
			var existingAttr models.Attribute
			if err := db.Where("attribute_name = ? AND id != ?", req.AttributeName, attributeID).
				First(&existingAttr).Error; err == nil {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusConflict, "Attribute with this name already exists"))
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fmt.Errorf("failed to check attribute existence: %w", err))
			}
		}

		// Update fields if provided
		if req.AttributeName != "" {
			attribute.AttributeName = req.AttributeName
		}
		if req.Type != "" {
			attribute.Type = req.Type
		}
		attribute.UpdatedAt = time.Now()

		if err := db.Save(&attribute).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to update attribute"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Attribute updated successfully",
			"data":    attribute,
		}, nil)
	}
}

// DeleteAttributeHandler handles deleting an attribute
// @Summary Delete an attribute
// @Description Permanently delete an attribute
// @Tags Attributes
// @Accept  json
// @Produce  json
// @Param id path string true "Attribute ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /attributes/{id} [delete]
func DeleteAttributeHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		attributeID := c.Params("id")
		if attributeID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Attribute ID is required"))
		}

		var attribute models.Attribute
		if err := db.Where("id = ?", attributeID).First(&attribute).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Attribute not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve attribute"))
		}

		// Check if attribute is used in any artwork before deleting
		var count int64
		if err := db.Model(&models.ArtworkAttribute{}).Where("attribute_id = ?", attributeID).Count(&count).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to check attribute usage"))
		}

		if count > 0 {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "Cannot delete attribute - it's being used in artworks"))
		}

		if err := db.Delete(&attribute).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to delete attribute"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Attribute deleted successfully",
		}, nil)
	}
}
