package medium

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	art "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	models "github.com/muga20/artsMarket/modules/artwork-management/models/medium"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// CreateMediumRequest represents the request body for creating a medium
type CreateMediumRequest struct {
	MediumName  string `json:"medium_name" validate:"required"`
	Description string `json:"description"`
}

// UpdateMediumRequest represents the request body for updating a medium
type UpdateMediumRequest struct {
	MediumName  string `json:"medium_name"`
	Description string `json:"description"`
}

// CreateMediumHandler handles creating a new medium
// @Summary Create a new medium
// @Description Create a new medium with name and description
// @Tags Mediums
// @Accept  json
// @Produce  json
// @Param request body CreateMediumRequest true "Medium creation payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /mediums [post]
func CreateMediumHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateMediumRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		// First check if medium name already exists
		var existingMedium models.Medium
		if err := db.Where("medium_name = ?", req.MediumName).First(&existingMedium).Error; err == nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "Medium with this name already exists"))
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to check medium existence: %w", err))
		}

		medium := models.Medium{
			ID:          uuid.New(),
			MediumName:  req.MediumName,
			Description: req.Description,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := db.Create(&medium).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to create medium: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Medium created successfully",
		}, nil)
	}
}

// GetAllMediumsHandler handles fetching all mediums
// @Summary Get all mediums
// @Description Retrieve a list of all mediums
// @Tags Mediums
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /mediums [get]
func GetAllMediumsHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var mediums []models.Medium
		if err := db.Find(&mediums).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve mediums"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"mediums": mediums,
			"count":   len(mediums),
		}, nil)
	}
}

// GetMediumByIDHandler handles fetching a medium by ID
// @Summary Get a medium by ID
// @Description Retrieve a medium's details using its ID
// @Tags Mediums
// @Accept  json
// @Produce  json
// @Param id path string true "Medium ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /mediums/{id} [get]
func GetMediumByIDHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mediumID := c.Params("id")
		if mediumID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Medium ID is required"))
		}

		var medium models.Medium
		if err := db.Where("id = ?", mediumID).First(&medium).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Medium not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve medium"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"data": medium,
		}, nil)
	}
}

// UpdateMediumHandler handles updating an existing medium
// @Summary Update an existing medium
// @Description Update medium details (name, description)
// @Tags Mediums
// @Accept  json
// @Produce  json
// @Param id path string true "Medium ID"
// @Param request body UpdateMediumRequest true "Medium update payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /mediums/{id} [put]
func UpdateMediumHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mediumID := c.Params("id")
		if mediumID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Medium ID is required"))
		}

		var req UpdateMediumRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		var medium models.Medium
		if err := db.Where("id = ?", mediumID).First(&medium).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Medium not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve medium"))
		}

		// Check if new name already exists (if being changed)
		if req.MediumName != "" && req.MediumName != medium.MediumName {
			var existingMedium models.Medium
			if err := db.Where("medium_name = ? AND id != ?", req.MediumName, mediumID).
				First(&existingMedium).Error; err == nil {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusConflict, "Medium with this name already exists"))
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fmt.Errorf("failed to check medium existence: %w", err))
			}
		}

		// Update fields if provided
		if req.MediumName != "" {
			medium.MediumName = req.MediumName
		}
		if req.Description != "" {
			medium.Description = req.Description
		}
		medium.UpdatedAt = time.Now()

		if err := db.Save(&medium).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to update medium"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Medium updated successfully",
		}, nil)
	}
}

// DeleteMediumHandler handles deleting a medium
// @Summary Delete a medium
// @Description Permanently delete a medium
// @Tags Mediums
// @Accept  json
// @Produce  json
// @Param id path string true "Medium ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /mediums/{id} [delete]
func DeleteMediumHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mediumID := c.Params("id")
		if mediumID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Medium ID is required"))
		}

		var medium models.Medium
		if err := db.Where("id = ?", mediumID).First(&medium).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Medium not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve medium"))
		}

		// Check if medium is used in any artwork before deleting
		var count int64
		if err := db.Model(&art.Artwork{}).Where("medium_id = ?", mediumID).Count(&count).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to check medium usage"))
		}

		if count > 0 {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "Cannot delete medium - it's being used in artworks"))
		}

		if err := db.Delete(&medium).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to delete medium"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Medium deleted successfully",
		}, nil)
	}
}
