package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/validation"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SignupRequest represents the expected request payload
type SignupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// SignupHandler handles user registration
// @Summary Register a new user
// @Description Create a new account with email, username, and password
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param request body SignupRequest true "User registration payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /auth/signup [post]
func SignupHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req SignupRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		// Validate email format
		if !validation.IsValidEmail(req.Email) {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid email format"))
		}

		// Validate password strength
		if !validation.IsValidPassword(req.Password) {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Password must contain at least 8 characters, including an uppercase letter, a lowercase letter, a number, and a special character"))
		}

		// Check if user already exists
		var existingUser models.User
		if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "User already exists"))
		}

		// Generate unique username
		username, err := generateUniqueUsername(db, req.Email)
		if err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to generate unique username: %w", err))
		}

		// Start transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Create new user
		newUser := models.User{
			ID:        uuid.New(),
			Email:     req.Email,
			Username:  username,
			Status:    "active",
			AuthType:  "email",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := tx.Create(&newUser).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to create user: %w", err))
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to hash password: %w", err))
		}

		// Store password
		userSecurity := models.UserSecurity{
			ID:       uuid.New(),
			UserID:   newUser.ID,
			Password: string(hashedPassword),
		}

		if err := tx.Create(&userSecurity).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to store user security data: %w", err))
		}

		// Assign default role
		if err := assignUserRole(tx, newUser.ID); err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to assign user role: %w", err))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "User created successfully",
		}, nil)
	}
}

func generateUniqueUsername(db *gorm.DB, email string) (string, error) {
	baseUsername := strings.ToLower(strings.Split(email, "@")[0])

	var user models.User
	if err := db.Where("username = ?", baseUsername).First(&user).Error; err == nil {
		baseUsername = baseUsername + "-" + uuid.New().String()[:8]
	}

	return baseUsername, nil
}

func assignUserRole(db *gorm.DB, userID uuid.UUID) error {
	var role models.Role
	if err := db.Where("role_name = ?", "user").First(&role).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		role = models.Role{
			ID:         uuid.New(),
			RoleName:   "user",
			RoleNumber: 1,
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := db.Create(&role).Error; err != nil {
			return fmt.Errorf("failed to create role: %w", err)
		}
	}

	userRole := models.UserRole{
		ID:       uuid.New(),
		UserID:   userID,
		RoleID:   role.ID,
		IsActive: true,
	}

	if err := db.Create(&userRole).Error; err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}
