package security

import (
	"fmt"
	"log"
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
		// Authentication check with proper type assertion
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Parse request body
		var req UpdateEmailRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))
		}

		// Validate email
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			errorMessages := make(map[string]string)
			for _, e := range validationErrors {
				if e.Field() == "Email" {
					errorMessages["email"] = "Valid email address is required"
				}
			}
			return responseHandler.HandleResponse(c, fiber.Map{
				"errors": errorMessages,
			}, fiber.NewError(fiber.StatusUnprocessableEntity, "Validation failed"))
		}

		// Check if email is already in use
		var existingUser models.User
		if err := db.Where("email = ? AND id != ?", req.Email, user.ID).
			First(&existingUser).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return responseHandler.HandleResponse(c, nil,
					fmt.Errorf("failed to check email availability: %w", err))
			}
		} else {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "Email is already in use"))
		}

		// Generate verification token
		verificationToken := uuid.New().String()
		expiration := time.Now().Add(24 * time.Hour)

		// Start explicit transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Update user email
		user.Email = req.Email
		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update email: %w", err))
		}

		// Update or create user security record
		var userSecurity models.UserSecurity
		if err := tx.Where("user_id = ?", user.ID).First(&userSecurity).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				userSecurity = models.UserSecurity{
					UserID:                 user.ID,
					IsEmailVerifiedAt:      nil,
					PasswordResetToken:     &verificationToken,
					PasswordResetExpiresAt: &expiration,
				}
			} else {
				tx.Rollback()
				return responseHandler.HandleResponse(c, nil,
					fmt.Errorf("failed to fetch user security: %w", err))
			}
		} else {
			userSecurity.IsEmailVerifiedAt = nil
			userSecurity.PasswordResetToken = &verificationToken
			userSecurity.PasswordResetExpiresAt = &expiration
		}

		if err := tx.Save(&userSecurity).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update security record: %w", err))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		// Send verification email (non-blocking)
		go func() {
			task, err := tasks.NewSendEmailVerificationTask(req.Email, verificationToken)
			if err != nil {
				log.Printf("Failed to create email verification task: %v", err)
				return
			}

			client := asynq.NewClient(*config.RedisConfig)
			defer client.Close()
			if _, err := client.Enqueue(task); err != nil {
				log.Printf("Failed to enqueue email verification task: %v", err)
			}
		}()

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Email updated successfully. Please verify your new email.",
		}, nil)
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
        // Validate token
        token := c.Query("token")
        if token == "" {
            return responseHandler.HandleResponse(c, nil,
                fiber.NewError(fiber.StatusBadRequest, "Verification token is required"))
        }

        // Start database transaction
        tx := db.Begin()
        if tx.Error != nil {
            return responseHandler.HandleResponse(c, nil,
                fmt.Errorf("failed to start transaction: %w", tx.Error))
        }

        // Find valid token
        var userSecurity models.UserSecurity
        if err := tx.Where("password_reset_token = ? AND password_reset_expires_at > ?", token, time.Now()).
            First(&userSecurity).Error; err != nil {
            tx.Rollback()
            if err == gorm.ErrRecordNotFound {
                return responseHandler.HandleResponse(c, nil,
                    fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired verification token"))
            }
            return responseHandler.HandleResponse(c, nil,
                fmt.Errorf("failed to verify token: %w", err))
        }

        // Update verification status
        now := time.Now()
        userSecurity.IsEmailVerifiedAt = &now
        userSecurity.PasswordResetToken = nil
        userSecurity.PasswordResetExpiresAt = nil

        if err := tx.Save(&userSecurity).Error; err != nil {
            tx.Rollback()
            return responseHandler.HandleResponse(c, nil,
                fmt.Errorf("failed to update verification status: %w", err))
        }

        // Commit transaction
        if err := tx.Commit().Error; err != nil {
            return responseHandler.HandleResponse(c, nil,
                fmt.Errorf("failed to commit transaction: %w", err))
        }

        return responseHandler.HandleResponse(c, fiber.Map{
            "message": "Email verified successfully",
        }, nil)
    }
}
