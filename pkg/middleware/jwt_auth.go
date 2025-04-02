package middleware

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// AuthMiddleware verifies the user's JWT token and retrieves user information
func AuthMiddleware(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {

		tokenString := c.Cookies("auth_token")

		// If no token is found, return an error
		if tokenString == "" {
			return responseHandler.Handle(c, nil, errors.New("please login"))
		}

		// Parse and validate the JWT token
		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is correct
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(config.Envs.DBUser), nil // Use your actual secret key
		})

		if err != nil || !token.Valid {
			return responseHandler.Handle(c, nil, errors.New("invalid or expired token"))
		}

		// Extract the user ID from the JWT claims
		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok {
			return responseHandler.Handle(c, nil, errors.New("invalid token claims"))
		}

		// Retrieve the user by ID from the database
		var user models.User
		if err := db.Where("id = ?", claims.Subject).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.Handle(c, nil, errors.New("user not found"))
			}
			return responseHandler.Handle(c, nil, fmt.Errorf("failed to retrieve user: %v", err))
		}

		c.Locals("user", user)

		// Proceed with the next handler
		return c.Next()
	}
}
