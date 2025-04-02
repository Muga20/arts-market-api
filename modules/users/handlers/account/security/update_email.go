package security

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/tasks"
	"gorm.io/gorm"
)

// UpdateEmailRequest defines the expected input for updating the email
type UpdateEmailRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

// UpdateEmail updates the authenticated user's email and requires verification
// @Summary Update the authenticated user's email
// @Description Updates the user's email and marks it as unverified until confirmed
// @Tags Security
// @Accept  json
// @Produce  json
// @Param   body  body  UpdateEmailRequest  true  "Email update data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /security/email [put]
func UpdateEmail(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.Handle(c, nil, errors.New("user not found in context"))
		}

		var req UpdateEmailRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.Handle(c, nil, errors.New("invalid request body"))
		}

		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			return responseHandler.Handle(c, map[string]string{"error": "validation failed"}, errors.New("validation failed"))
		}

		var existingUser models.User
		if err := db.Where("email = ? AND id != ?", req.Email, user.ID).First(&existingUser).Error; err != gorm.ErrRecordNotFound {
			if err == nil {
				return responseHandler.Handle(c, nil, errors.New("email is already in use"))
			}
			return responseHandler.Handle(c, nil, err)
		}

		verificationToken := uuid.New().String()
		err := db.Transaction(func(tx *gorm.DB) error {
			user.Email = req.Email
			if err := tx.Save(&user).Error; err != nil {
				return err
			}

			var userSecurity models.UserSecurity
			if err := tx.Where("user_id = ?", user.ID).First(&userSecurity).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					userSecurity = models.UserSecurity{UserID: user.ID, Password: ""}
				} else {
					return err
				}
			}

			userSecurity.IsEmailVerifiedAt = nil
			expiration := time.Now().Add(24 * time.Hour)
			userSecurity.PasswordResetToken = &verificationToken
			userSecurity.PasswordResetExpiresAt = &expiration

			if err := tx.Save(&userSecurity).Error; err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return responseHandler.Handle(c, nil, err)
		}

		task, err := tasks.NewSendEmailVerificationTask(req.Email, verificationToken)
		if err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to create email verification task"))
		}

		client := asynq.NewClient(*config.RedisConfig)
		defer client.Close()
		_, err = client.Enqueue(task)
		if err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to enqueue email verification task"))
		}

		return responseHandler.Handle(c, map[string]string{
			"message": "email updated successfully, please verify your new email"}, nil)
	}
}

// VerifyEmail verifies the user's new email using a token
// @Summary Verify the user's email
// @Description Verifies the user's email address using a token sent after updating the email
// @Tags Security
// @Accept  json
// @Produce  json
// @Param   token  query  string  true  "Verification token"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /security/verify-email [get]
func VerifyEmail(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Query("token")
		if token == "" {
			return responseHandler.Handle(c, nil, errors.New("token is required"))
		}

		var userSecurity models.UserSecurity
		if err := db.Where("password_reset_token = ? AND password_reset_expires_at > ?", token, time.Now()).First(&userSecurity).Error; err != nil {
			return responseHandler.Handle(c, nil, errors.New("invalid or expired token"))
		}

		now := time.Now()
		userSecurity.IsEmailVerifiedAt = &now
		userSecurity.PasswordResetToken = nil
		userSecurity.PasswordResetExpiresAt = nil
		if err := db.Save(&userSecurity).Error; err != nil {
			return responseHandler.Handle(c, nil, err)
		}

		return responseHandler.Handle(c, map[string]string{
			"message": "email verified successfully",
		}, nil)
	}
}
