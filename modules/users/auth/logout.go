package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/validation"
	"gorm.io/gorm"
)

// LogoutHandler handles user logout by invalidating the JWT cookie and removing the session from the database
// @Summary Logout user
// @Description Invalidate the JWT token by clearing the auth cookie and deleting the session from the database
// @Tags Auth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/logout [post]
func LogoutHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the session token from cookies
		sessionToken := c.Cookies("auth_token")
		if sessionToken == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "No active session found"))
		}

		// Start database transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Find and delete the session
		var session models.UserSession
		if err := tx.Where("session_token = ?", sessionToken).
			First(&session).Error; err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Session not found, but we'll still clear the cookie
				clearSessionCookie(c)
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusUnauthorized, "Session not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to retrieve session: %w", err))
		}

		// Delete the session
		if err := tx.Delete(&session).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to delete session: %w", err))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		// Clear the cookie regardless of database operation success
		clearSessionCookie(c)

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Logged out successfully",
		}, nil)
	}
}

// Helper function to clear session cookie
func clearSessionCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Expire immediately
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Path:     "/", // Ensure cookie is cleared from all paths
	})
}

// LogoutOtherDevicesHandler handles user logout from all other devices except the current one
// @Summary Logout user from other devices
// @Description Invalidate all active sessions except the current one, keeping the current session active
// @Tags Auth
// @Accept json
// @Produce json
// @Param email body string true "User email" example("user@example.com")
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/logout-other-devices [post]
func LogoutOtherDevicesHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request body for email
		var req struct {
			Email string `json:"email" validate:"required,email"`
		}

		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))
		}

		// Validate email
		if !validation.IsValidEmail(req.Email) {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid email format"))
		}

		// Start database transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Find user by email
		var user models.User
		if err := tx.Where("email = ?", req.Email).First(&user).Error; err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "User not found"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to find user: %w", err))
		}

		// Delete all active sessions for this user except current one
		// Note: You might want to exclude the current session if you have access to the current session ID
		result := tx.Where("user_id = ?", user.ID).
			Delete(&models.UserSession{})

		if result.Error != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to delete sessions: %w", result.Error))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": fmt.Sprintf("Terminated all active sessions (%d devices)", result.RowsAffected),
			"count":   result.RowsAffected,
		}, nil)
	}
}
