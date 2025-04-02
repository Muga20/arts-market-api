package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/validation"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LoginRequest represents the expected login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// JWT secret key (keep it secure)
var jwtSecretKey = []byte(config.Envs.DBUser)

// Rate limiter map to track failed login attempts (based on IP)
var failedLoginAttempts = make(map[string]int)

// Max failed login attempts before triggering lock
const maxFailedLoginAttempts = 5

// LoginHandler handles user login and JWT generation
// @Summary Login with email and password
// @Description Authenticate user and generate JWT token
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param request body LoginRequest true "User login payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func LoginHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.Handle(c, nil, errors.New("invalid request payload"))
		}

		// Validate email format using validation package
		if !validation.IsValidEmail(req.Email) {
			return responseHandler.Handle(c, nil, errors.New("invalid email format"))
		}
		// Prevent SQL Injection
		if containsSQLInjection(req.Email) || containsSQLInjection(req.Password) {
			return responseHandler.Handle(c, nil, errors.New("invalid request payload"))
		}

		// Check if IP exceeded login attempt limit
		ipAddress := c.IP()
		if failedLoginAttempts[ipAddress] >= maxFailedLoginAttempts {
			return responseHandler.Handle(c, nil, errors.New("too many failed login attempts, try again later"))
		}

		// Retrieve user by email
		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				failedLoginAttempts[ipAddress]++
				return responseHandler.Handle(c, nil, errors.New("invalid credentials"))
			}
			return responseHandler.Handle(c, nil, errors.New("failed to retrieve user"))
		}

		// Retrieve user's security information
		var userSecurity models.UserSecurity
		if err := db.Where("user_id = ?", user.ID).First(&userSecurity).Error; err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to retrieve user security information"))
		}

		// Compare the provided password with the stored hashed password
		if err := bcrypt.CompareHashAndPassword([]byte(userSecurity.Password), []byte(req.Password)); err != nil {
			failedLoginAttempts[ipAddress]++
			return responseHandler.Handle(c, nil, errors.New("invalid credentials"))
		}

		// Check if there are any active sessions for the user
		var activeSessions []models.UserSession
		if err := db.Where("user_id = ? AND session_token != ?", user.ID, "").Find(&activeSessions).Error; err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to retrieve active sessions"))
		}

		// If there's an active session from a different device, block login
		if len(activeSessions) > 0 {
			return responseHandler.Handle(c, nil, errors.New("user already logged in from another device"))
		}

		// Reset failed login attempts on successful login
		failedLoginAttempts[ipAddress] = 0

		// Generate JWT token
		token, err := generateJWT(user.ID)
		if err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to generate JWT token"))
		}

		// Set the JWT token as an HTTP-only cookie
		c.Cookie(&fiber.Cookie{
			Name:     "auth_token",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: true,
			Secure:   true,
			SameSite: fiber.CookieSameSiteStrictMode,
		})

		// Create or update session in the UserSession table
		session := models.UserSession{
			ID:           uuid.New(),
			UserID:       user.ID,
			SessionToken: token,
			IPAddress:    c.IP(),
			DeviceInfo:   c.Get("User-Agent"),
			CreatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(time.Hour * 24),
		}

		if err := db.Create(&session).Error; err != nil {
			return responseHandler.Handle(c, nil, errors.New("failed to create session"))
		}

		return responseHandler.Handle(c, fiber.Map{
			"message": "Login successful",
		}, nil)
	}
}

// generateJWT generates a JWT token for the user
func generateJWT(userID uuid.UUID) (string, error) {
	// Set claims (payload)
	claims := &jwt.RegisteredClaims{
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	return token.SignedString(jwtSecretKey)
}

// Helper function to detect SQL injection patterns
func containsSQLInjection(input string) bool {
	// Basic patterns that indicate SQL injection
	badPatterns := []string{"'--", "' OR 1=1", "' OR 'a'='a'", "' DROP", ";--"}
	for _, pattern := range badPatterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}
	return false
}
