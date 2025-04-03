package roles

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// CreateRoleRequest represents the request body for creating a role
type CreateRoleRequest struct {
	RoleName   string `json:"role_name" validate:"required"`
	RoleNumber int    `json:"role_number" validate:"required"`
}

// UpdateRoleRequest represents the request body for updating a role
type UpdateRoleRequest struct {
	RoleName   string `json:"role_name"`
	RoleNumber int    `json:"role_number"`
}

// CreateRoleHandler handles creating a new role
// @Summary Create a new role
// @Description Create a new role with name, role number, and active status
// @Tags Roles
// @Accept  json
// @Produce  json
// @Param request body CreateRoleRequest true "Role creation payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /roles [post]
func CreateRoleHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateRoleRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		role := models.Role{
			ID:         uuid.New(),
			RoleName:   req.RoleName,
			RoleNumber: req.RoleNumber,
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := db.Create(&role).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to create role: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Role created successfully",
		}, nil)
	}
}

// GetAllRolesHandler handles fetching all roles
// @Summary Get all roles
// @Description Retrieve a list of all roles
// @Tags Roles
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /roles [get]
func GetAllRolesHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var roles []models.Role
		if err := db.Find(&roles).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve roles"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"data": fiber.Map{
				"roles": roles,
				"count": len(roles),
			},
		}, nil)
	}
}

// GetRoleByIDHandler handles fetching a role by ID
// @Summary Get a role by ID
// @Description Retrieve a role's details using its ID
// @Tags Roles
// @Accept  json
// @Produce  json
// @Param id path string true "Role ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /roles/{id} [get]
func GetRoleByIDHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID := c.Params("id")
		if roleID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Role ID is required"))
		}

		var role models.Role
		if err := db.Where("id = ?", roleID).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Role not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve role"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"data": fiber.Map{
				"role": role,
			},
		}, nil)
	}
}

// UpdateRoleHandler handles updating an existing role
// @Summary Update an existing role
// @Description Update role details (name, number)
// @Tags Roles
// @Accept  json
// @Produce  json
// @Param id path string true "Role ID"
// @Param request body UpdateRoleRequest true "Role update payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /roles/{id} [put]
func UpdateRoleHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID := c.Params("id")
		if roleID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Role ID is required"))
		}

		var req UpdateRoleRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		var role models.Role
		if err := db.Where("id = ?", roleID).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Role not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve role"))
		}

		// Update role details if provided
		if req.RoleName != "" {
			role.RoleName = req.RoleName
		}
		if req.RoleNumber != 0 {
			role.RoleNumber = req.RoleNumber
		}
		role.UpdatedAt = time.Now()

		if err := db.Save(&role).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to update role"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Role updated successfully",
		}, nil)
	}
}

// ActivateRoleHandler handles activating a role
// @Summary Activate a role
// @Description Set the status of a role to active
// @Tags Roles
// @Accept  json
// @Produce  json
// @Param id path string true "Role ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /roles/{id}/activate [put]
func ActivateRoleHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID := c.Params("id")
		if roleID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Role ID is required"))
		}

		var role models.Role
		if err := db.Where("id = ?", roleID).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Role not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve role"))
		}

		// Set role as active
		role.IsActive = true
		role.UpdatedAt = time.Now()

		if err := db.Save(&role).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to activate role"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Role activated successfully",
		}, nil)
	}
}

// DeactivateRoleHandler handles deactivating a role
// @Summary Deactivate a role
// @Description Set the status of a role to inactive
// @Tags Roles
// @Accept  json
// @Produce  json
// @Param id path string true "Role ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /roles/{id}/deactivate [put]
func DeactivateRoleHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID := c.Params("id")
		if roleID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Role ID is required"))
		}

		var role models.Role
		if err := db.Where("id = ?", roleID).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Role not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve role"))
		}

		// Set role as inactive
		role.IsActive = false
		role.UpdatedAt = time.Now()

		if err := db.Save(&role).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to deactivate role"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Role deactivated successfully",
		}, nil)
	}
}
