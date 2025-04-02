package auth

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
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
		// Get the session token from the auth_token cookie
		sessionToken := c.Cookies("auth_token")
		if sessionToken == "" {
			return responseHandler.Handle(c, nil, errors.New("no session found"))
		}

		// Find the session in the database by session token
		var session models.UserSession
		if err := db.Where("session_token = ?", sessionToken).First(&session).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.Handle(c, nil, errors.New("session not found"))
			}
			return responseHandler.Handle(c, nil, errors.New("failed to retrieve session"))
		}

		// Delete the session from the database
		if err := db.Delete(&session).Error; err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to delete session"))
		}

		// Clear the auth_token cookie by setting it to an expired time
		c.Cookie(&fiber.Cookie{
			Name:     "auth_token",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			HTTPOnly: true,
			Secure:   true,
			SameSite: fiber.CookieSameSiteStrictMode,
		})

		return responseHandler.Handle(c, fiber.Map{
			"message": "Logout successful",
		}, nil)
	}
}

// LogoutOtherDevicesHandler handles user logout from all other devices except the current one
// @Summary Logout user from other devices
// @Description Invalidate the JWT token from all other devices, keeping the current session active
// @Tags Auth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/logout-other-devices [post]
func LogoutOtherDevicesHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the session token from the auth_token cookie
		sessionToken := c.Cookies("auth_token")
		if sessionToken == "" {
			return responseHandler.Handle(c, nil, errors.New("no session found"))
		}

		// Find the session in the database by session token
		var session models.UserSession
		if err := db.Where("session_token = ?", sessionToken).First(&session).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.Handle(c, nil, errors.New("session not found"))
			}
			return responseHandler.Handle(c, nil, errors.New("failed to retrieve session"))
		}

		// Find all sessions for this user except the current session
		var sessions []models.UserSession
		if err := db.Where("user_id = ? AND session_token != ?", session.UserID, sessionToken).Find(&sessions).Error; err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to retrieve user sessions"))
		}

		// Delete all sessions from other devices
		for _, otherSession := range sessions {
			if err := db.Delete(&otherSession).Error; err != nil {
				return responseHandler.Handle(c, nil, errors.New("failed to delete session"))
			}
		}

		return responseHandler.Handle(c, fiber.Map{
			"message": "Logged out from other devices successfully",
		}, nil)
	}
}
