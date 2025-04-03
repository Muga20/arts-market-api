package roles

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/validation"
	"gorm.io/gorm"
)

// Define the Request struct globally
type Request struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

// AssignRole assigns a role to a user
// @Summary Assign a role to a user
// @Description Assigns a specified role to a user in the system
// @Tags Roles
// @Accept json
// @Produce json
// @Param request body Request true "User and Role IDs"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /roles/assign [post]
func AssignRole(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req Request
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))
		}

		// Validate UUIDs
		if !validation.IsValidUUID(req.UserID.String()) || !validation.IsValidUUID(req.RoleID.String()) {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid ID format"))
		}

		// Start transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Check if user exists with row lock
		var user models.User
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ?", req.UserID).
			First(&user).Error; err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "User not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to find user: %w", err))
		}

		// Check if role exists
		var role models.Role
		if err := tx.Where("id = ?", req.RoleID).First(&role).Error; err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Role not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to find role: %w", err))
		}

		// Check for existing role assignment
		var existingAssignment models.UserRole
		if err := tx.Where("user_id = ? AND role_id = ?", req.UserID, req.RoleID).
			First(&existingAssignment).Error; err == nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "User already has this role"))
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to check existing role assignment: %w", err))
		}

		// Create new role assignment
		newAssignment := models.UserRole{
			ID:       uuid.New(),
			UserID:   req.UserID,
			RoleID:   req.RoleID,
			IsActive: true,
		}

		if err := tx.Create(&newAssignment).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to assign role: %w", err))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Role assigned successfully",
		}, nil)
	}
}

// RemoveRole removes a role from a user
// @Summary Remove a role from a user
// @Description Removes a specific role from a user
// @Tags Roles
// @Accept json
// @Produce json
// @Param request body Request true "User and Role IDs"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /roles/remove [post]
func RemoveRole(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req Request
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))
		}

		// Validate UUIDs
		if !validation.IsValidUUID(req.UserID.String()) || !validation.IsValidUUID(req.RoleID.String()) {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid ID format"))
		}

		// Start transaction with serializable isolation level for data consistency
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Find and lock the user role entry
		var userRole models.UserRole
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ? AND role_id = ?", req.UserID, req.RoleID).
			First(&userRole).Error; err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Role assignment not found for this user"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to find role assignment: %w", err))
		}

		// Delete role assignment
		if err := tx.Delete(&userRole).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to remove role assignment: %w", err))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Role removed successfully",
		}, nil)
	}
}

// GetUserRoles fetches all roles assigned to a user
// @Summary Get all roles assigned to a user
// @Description Retrieves a list of roles assigned to a user
// @Tags Roles
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /roles/for/{user_id} [get]
func GetUserRoles(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("user_id")
		if userID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "user_id is required"))
		}

		parsedID, err := uuid.Parse(userID)
		if err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid user_id format"))
		}

		// Use parallel execution for better performance
		errChan := make(chan error, 1)
		var wg sync.WaitGroup
		wg.Add(1)

		var roles []struct {
			ID         uuid.UUID `json:"id"`
			RoleName   string    `json:"role_name"`
			RoleNumber int       `json:"role_number"`
			CreatedAt  time.Time `json:"created_at"`
			UpdatedAt  time.Time `json:"updated_at"`
		}

		// Fetch roles assigned to the user
		go func() {
			defer wg.Done()
			if err := db.Table("roles").
				Select("roles.id, roles.role_name, roles.role_number, roles.created_at, roles.updated_at").
				Joins("JOIN user_roles ON user_roles.role_id = roles.id").
				Where("user_roles.user_id = ?", parsedID).
				Scan(&roles).Error; err != nil {
				errChan <- fmt.Errorf("failed to fetch user roles: %w", err)
			}
		}()

		wg.Wait()
		close(errChan)

		// Check for errors
		for e := range errChan {
			if e != nil {
				return responseHandler.HandleResponse(c, nil, e)
			}
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "User roles retrieved successfully",
		}, nil)
	}
}
