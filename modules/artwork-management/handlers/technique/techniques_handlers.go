package technique

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	art "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	models "github.com/muga20/artsMarket/modules/artwork-management/models/technique"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// CreateTechniqueRequest represents the request body for creating a technique
type CreateTechniqueRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

// UpdateTechniqueRequest represents the request body for updating a technique
type UpdateTechniqueRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateTechniqueHandler handles creating a new technique
// @Summary Create a new technique
// @Description Create a new technique with name and description
// @Tags Techniques
// @Accept  json
// @Produce  json
// @Param request body CreateTechniqueRequest true "Technique creation payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /techniques [post]
func CreateTechniqueHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateTechniqueRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		// First check if technique name already exists
		var existingTechnique models.Technique
		if err := db.Where("name = ?", req.Name).First(&existingTechnique).Error; err == nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "Technique with this name already exists"))
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to check technique existence: %w", err))
		}

		technique := models.Technique{
			ID:          uuid.New(),
			Name:        req.Name,
			Description: req.Description,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := db.Create(&technique).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to create technique: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Technique created successfully",
		}, nil)
	}
}

// GetAllTechniquesHandler handles fetching all techniques
// @Summary Get all techniques
// @Description Retrieve a list of all techniques
// @Tags Techniques
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /techniques [get]
func GetAllTechniquesHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var techniques []models.Technique
		if err := db.Find(&techniques).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve techniques"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"techniques": techniques,
			"count":      len(techniques),
		}, nil)
	}
}

// GetTechniqueByIDHandler handles fetching a technique by ID
// @Summary Get a technique by ID
// @Description Retrieve a technique's details using its ID
// @Tags Techniques
// @Accept  json
// @Produce  json
// @Param id path string true "Technique ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /techniques/{id} [get]
func GetTechniqueByIDHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		techniqueID := c.Params("id")
		if techniqueID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Technique ID is required"))
		}

		var technique models.Technique
		if err := db.Where("id = ?", techniqueID).First(&technique).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Technique not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve technique"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"data": technique,
		}, nil)
	}
}

// UpdateTechniqueHandler handles updating an existing technique
// @Summary Update an existing technique
// @Description Update technique details (name, description)
// @Tags Techniques
// @Accept  json
// @Produce  json
// @Param id path string true "Technique ID"
// @Param request body UpdateTechniqueRequest true "Technique update payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /techniques/{id} [put]
func UpdateTechniqueHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		techniqueID := c.Params("id")
		if techniqueID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Technique ID is required"))
		}

		var req UpdateTechniqueRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		var technique models.Technique
		if err := db.Where("id = ?", techniqueID).First(&technique).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Technique not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve technique"))
		}

		// Check if new name already exists (if being changed)
		if req.Name != "" && req.Name != technique.Name {
			var existingTechnique models.Technique
			if err := db.Where("name = ? AND id != ?", req.Name, techniqueID).
				First(&existingTechnique).Error; err == nil {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusConflict, "Technique with this name already exists"))
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fmt.Errorf("failed to check technique existence: %w", err))
			}
		}

		// Update fields if provided
		if req.Name != "" {
			technique.Name = req.Name
		}
		if req.Description != "" {
			technique.Description = req.Description
		}
		technique.UpdatedAt = time.Now()

		if err := db.Save(&technique).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to update technique"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Technique updated successfully",
		}, nil)
	}
}

// DeleteTechniqueHandler handles deleting a technique
// @Summary Delete a technique
// @Description Permanently delete a technique
// @Tags Techniques
// @Accept  json
// @Produce  json
// @Param id path string true "Technique ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /techniques/{id} [delete]
func DeleteTechniqueHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		techniqueID := c.Params("id")
		if techniqueID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Technique ID is required"))
		}

		var technique models.Technique
		if err := db.Where("id = ?", techniqueID).First(&technique).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Technique not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve technique"))
		}

		// Check if technique is used in any artwork before deleting
		var count int64
		if err := db.Model(&art.Artwork{}).Where("technique_id = ?", techniqueID).Count(&count).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to check technique usage"))
		}

		if count > 0 {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "Cannot delete technique - it's being used in artworks"))
		}

		if err := db.Delete(&technique).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to delete technique"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Technique deleted successfully",
		}, nil)
	}
}
