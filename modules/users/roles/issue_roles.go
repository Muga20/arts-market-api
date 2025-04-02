package roles

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
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
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false, "message": "Invalid request body",
			}, errors.New("invalid request body"))
		}

		// Check if user exists
		var user models.User
		if err := db.First(&user, "id = ?", req.UserID).Error; err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false, "message": "User not found",
			}, errors.New("user not found"))
		}

		// Check if role exists
		var role models.Role
		if err := db.First(&role, "id = ?", req.RoleID).Error; err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false, "message": "Role not found",
			}, errors.New("role not found"))
		}

		// Check if user already has the role
		var userRole models.UserRole
		if err := db.Where("user_id = ? AND role_id = ?", req.UserID, req.RoleID).First(&userRole).Error; err == nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false, "message": "User already has this role",
			}, errors.New("user already has this role"))
		}

		// Assign role to user
		userRole = models.UserRole{
			UserID:   req.UserID,
			RoleID:   req.RoleID,
			IsActive: true,
		}
		if err := db.Create(&userRole).Error; err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false, "message": "Failed to assign role",
			}, errors.New("failed to assign role"))
		}

		return responseHandler.Handle(c, map[string]interface{}{
			"success": true, "message": "Role assigned successfully",
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
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false, "message": "Invalid request body",
			}, errors.New("invalid request body"))
		}

		// Find user role entry
		var userRole models.UserRole
		if err := db.Where("user_id = ? AND role_id = ?", req.UserID, req.RoleID).First(&userRole).Error; err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false, "message": "Role not found for this user",
			}, errors.New("role not found for this user"))
		}

		// Delete role assignment
		if err := db.Delete(&userRole).Error; err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false, "message": "Failed to remove role",
			}, errors.New("failed to remove role"))
		}

		return responseHandler.Handle(c, map[string]interface{}{
			"success": true, "message": "Role removed successfully",
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
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false, "message": "user_id is required",
			}, errors.New("user_id is required"))
		}

		parsedID, err := uuid.Parse(userID)
		if err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false, "message": "Invalid user_id format",
			}, errors.New("invalid user_id format"))
		}

		// Define a slice to store only role data
		var roles []models.Role

		// Fetch roles assigned to the user
		if err := db.Joins("JOIN user_roles ON user_roles.role_id = roles.id").
			Where("user_roles.user_id = ?", parsedID).
			Find(&roles).Error; err != nil {
			return responseHandler.Handle(c, map[string]interface{}{
				"success": false, "message": "Failed to fetch user roles",
			}, errors.New("failed to fetch user roles"))
		}

		return responseHandler.Handle(c, map[string]interface{}{
			"success": true, "message": "User roles retrieved successfully", "roles": roles,
		}, nil)
	}
}
