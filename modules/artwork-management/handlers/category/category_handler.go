package category

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	models "github.com/muga20/artsMarket/modules/artwork-management/models/category"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// CreateCategoryRequest represents the request body for creating a category
type CreateCategoryRequest struct {
	CategoryName string `json:"category_name" validate:"required"`
	Description  string `json:"description"`
}

// UpdateCategoryRequest represents the request body for updating a category
type UpdateCategoryRequest struct {
	CategoryName string `json:"category_name"`
	Description  string `json:"description"`
}

// CreateCategoryHandler handles creating a new category
// @Summary Create a new category
// @Description Create a new category with name, description, and active status
// @Tags Categories
// @Accept  json
// @Produce  json
// @Param request body CreateCategoryRequest true "Category creation payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /categories [post]
func CreateCategoryHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateCategoryRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		// First check if category name already exists
		var existingCategory models.Category
		if err := db.Where("category_name = ?", req.CategoryName).First(&existingCategory).Error; err == nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "Category with this name already exists"))
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			// If error is something other than "not found", return it
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to check category existence: %w", err))
		}

		// Create new category since name doesn't exist
		category := models.Category{
			ID:           uuid.New(),
			CategoryName: req.CategoryName,
			Description:  req.Description,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := db.Create(&category).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to create category: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Category created successfully",
		}, nil)
	}
}

// GetAllCategoriesHandler handles fetching all categories
// @Summary Get all categories
// @Description Retrieve a list of all categories
// @Tags Categories
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /categories [get]
func GetAllCategoriesHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var categories []models.Category
		if err := db.Find(&categories).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve categories"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"categories": categories,
			"count":      len(categories),
		}, nil)
	}
}

// GetCategoryHandler handles fetching a category by either ID or slug
// @Summary Get category by ID or slug
// @Description Retrieves a category by its ID or slug
// @Tags Categories
// @Accept json
// @Produce json
// @Param identifier path string true "Category ID (UUID) or slug"
// @Success 200 {object} fiber.Map "Returns the category data"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Category not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /categories/{identifier} [get]
func GetCategoryHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		identifier := c.Params("identifier")
		if identifier == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Category identifier is required"))
		}

		var category models.Category
		query := db.Model(&models.Category{})

		// Determine if identifier is UUID or slug
		if _, err := uuid.Parse(identifier); err == nil {
			// It's a valid UUID, search by ID
			query = query.Where("id = ?", identifier)
		} else {
			// Treat as slug
			query = query.Where("slug = ?", identifier)
		}

		if err := query.First(&category).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Category not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve category"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"data": category,
		}, nil)
	}
}

// UpdateCategoryHandler handles updating an existing category
// @Summary Update an existing category
// @Description Update category details (name, description)
// @Tags Categories
// @Accept  json
// @Produce  json
// @Param id path string true "Category ID"
// @Param request body UpdateCategoryRequest true "Category update payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /categories/{id} [put]
func UpdateCategoryHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		categoryID := c.Params("id")
		if categoryID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Category ID is required"))
		}

		var req UpdateCategoryRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		var category models.Category
		if err := db.Where("id = ?", categoryID).First(&category).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Category not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve category"))
		}

		// Update fields if provided
		if req.CategoryName != "" {
			category.CategoryName = req.CategoryName
		}
		if req.Description != "" {
			category.Description = req.Description
		}
		category.UpdatedAt = time.Now()

		if err := db.Save(&category).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusConflict, "Category with this name already exists"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to update category"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Category updated successfully",
		}, nil)
	}
}

// ToggleCategoryActiveHandler handles toggling a category's active status
// @Summary Toggle category active status
// @Description Toggle a category's active status (activate/deactivate)
// @Tags Categories
// @Accept  json
// @Produce  json
// @Param id path string true "Category ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /categories/{id}/toggle-active [put]
func ToggleCategoryActiveHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		categoryID := c.Params("id")
		if categoryID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Category ID is required"))
		}

		var category models.Category
		if err := db.Where("id = ?", categoryID).First(&category).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Category not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve category"))
		}

		// Toggle active status
		category.IsActive = !category.IsActive
		category.UpdatedAt = time.Now()

		if err := db.Save(&category).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to update category status"))
		}

		action := "activated"
		if !category.IsActive {
			action = "deactivated"
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": fmt.Sprintf("Category %s successfully", action),
		}, nil)
	}
}

// DeleteCategoryHandler handles permanently deleting a category
// @Summary Delete a category
// @Description Permanently delete a category (use with caution)
// @Tags Categories
// @Accept  json
// @Produce  json
// @Param id path string true "Category ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /categories/{id} [delete]
func DeleteCategoryHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		categoryID := c.Params("id")
		if categoryID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Category ID is required"))
		}

		// First check if category exists
		var category models.Category
		if err := db.Where("id = ?", categoryID).First(&category).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Category not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve category"))
		}

		// Check if category is used in any artwork before deleting
		var count int64
		if err := db.Model(&models.ArtworkCategory{}).Where("category_id = ?", categoryID).Count(&count).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to check category usage"))
		}

		if count > 0 {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "Cannot delete category - it's being used in artworks"))
		}

		// Perform the delete
		if err := db.Delete(&category).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to delete category"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Category deleted successfully",
		}, nil)
	}
}
