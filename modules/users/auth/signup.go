package auth

import (
	"errors"
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
			return responseHandler.Handle(c, nil, errors.New("invalid request payload"))
		}

		// Validate email format using validation package
		if !validation.IsValidEmail(req.Email) {
			return responseHandler.Handle(c, nil, errors.New("invalid email format"))
		}

		// Validate password strength
		if !validation.IsValidPassword(req.Password) {
			return responseHandler.Handle(c, nil, errors.New("password must contain at least 8 characters, including an uppercase letter, a lowercase letter, a number, and a special character"))
		}

		// Check if user already exists
		var existingUser models.User
		if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			return responseHandler.Handle(c, nil, errors.New("user already exists"))
		}

		// Generate unique username based on email
		username, err := generateUniqueUsername(db, req.Email)
		if err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to generate unique username"))
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

		if err := db.Create(&newUser).Error; err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to create user"))
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to hash password"))
		}

		// Store password in UserSecurity table
		userSecurity := models.UserSecurity{
			ID:       uuid.New(),
			UserID:   newUser.ID,
			Password: string(hashedPassword),
		}

		if err := db.Create(&userSecurity).Error; err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to store user security data"))
		}

		// Assign default "user" role
		if err := assignUserRole(db, newUser.ID); err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to assign user role"))
		}

		return responseHandler.Handle(c, fiber.Map{"message": "User created successfully"}, nil)
	}
}

// generateUniqueUsername generates a unique username based on email
func generateUniqueUsername(db *gorm.DB, email string) (string, error) {
	baseUsername := strings.ToLower(strings.Split(email, "@")[0]) // Get the part before the "@" in the email

	// Check if the username is already taken
	var user models.User
	if err := db.Where("username = ?", baseUsername).First(&user).Error; err == nil {
		// If username is taken, append a unique UUID to it
		baseUsername = baseUsername + "-" + uuid.New().String()[:8]
	}

	return baseUsername, nil
}

// assignUserRole assigns the "user" role to a new user
func assignUserRole(db *gorm.DB, userID uuid.UUID) error {
	var role models.Role
	if err := db.Where("role_name = ?", "user").First(&role).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		// If role doesn't exist, create it
		role = models.Role{
			ID:         uuid.New(),
			RoleName:   "user",
			RoleNumber: 1,
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := db.Create(&role).Error; err != nil {
			return err
		}
	}

	// Assign role to user
	userRole := models.UserRole{
		ID:       uuid.New(),
		UserID:   userID,
		RoleID:   role.ID,
		IsActive: true,
	}

	return db.Create(&userRole).Error
}
