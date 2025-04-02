package auth

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/validation"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ResetPassword represents the expected reset password input
type ResetPassword struct {
	Token    string `json:"token" validate:"required,uuid4"`
	Password string `json:"password" validate:"required,min=8,complex"`
}

// ResetPasswordHandler validates the reset token and updates the password
// @Summary Reset password using reset token
// @Description Validate the reset token and update the user's password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPassword true "Reset Password"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/reset-password [post]
func ResetPasswordHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req ResetPassword
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.Handle(c, nil, errors.New("invalid request payload"))
		}

		// Validate token format (UUID)
		if !validation.IsValidUUID(req.Token) {
			return responseHandler.Handle(c, nil, errors.New("invalid reset token format"))
		}

		// Retrieve user security by reset token
		var userSecurity models.UserSecurity
		if err := db.Where("password_reset_token = ?", req.Token).First(&userSecurity).Error; err != nil {
			return responseHandler.Handle(c, nil, errors.New("invalid reset token"))
		}

		// Check if token has expired
		if userSecurity.PasswordResetExpiresAt.Before(time.Now()) {
			return responseHandler.Handle(c, nil, errors.New("reset token has expired"))
		}

		// Validate password strength
		if !validation.IsValidPassword(req.Password) {
			return responseHandler.Handle(c, nil, errors.New("password does not meet complexity requirements"))
		}

		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to hash password"))
		}

		// Update the user's password and clear the reset token
		userSecurity.Password = string(hashedPassword)
		userSecurity.PasswordResetToken = nil
		userSecurity.PasswordResetExpiresAt = nil

		if err := db.Save(&userSecurity).Error; err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to update password"))
		}

		return responseHandler.Handle(c, fiber.Map{
			"message": "Password successfully reset",
		}, nil)
	}
}
