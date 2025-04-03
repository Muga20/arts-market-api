package security

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// UpdateUsernameRequest defines the request body for updating a username
type UpdateUsernameRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
}

// UpdateUsername updates the authenticated user's username
// @Summary Update the authenticated user's username
// @Description Changes the username if it's not taken by another user
// @Tags Security
// @Accept json
// @Produce json
// @Param body body UpdateUsernameRequest true "Username update data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /security/username [put]
func UpdateUsername(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authentication check with proper type assertion
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Parse request body
		var req UpdateUsernameRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))
		}

		// Validate username
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			errorMessages := make(map[string]string)
			for _, e := range validationErrors {
				if e.Field() == "Username" {
					errorMessages["username"] = "Username must be between 3-20 alphanumeric characters"
				}
			}
			return responseHandler.HandleResponse(c, fiber.Map{
				"errors": errorMessages,
			}, fiber.NewError(fiber.StatusUnprocessableEntity, "Validation failed"))
		}

		// Check if username is same as current
		if req.Username == user.Username {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "New username cannot be the same as current username"))
		}

		// Check if username is available
		var existingUser models.User
		if err := db.Where("username = ? AND id != ?", req.Username, user.ID).
			First(&existingUser).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return responseHandler.HandleResponse(c, nil,
					fmt.Errorf("failed to check username availability: %w", err))
			}
		} else {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusConflict, "Username is already taken"))
		}

		// Update username
		user.Username = req.Username
		if err := db.Save(&user).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update username: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Username updated successfully",
		}, nil)
	}
}

// SearchUsernames searches for usernames based on a keyword
// @Summary Search for usernames
// @Description Finds usernames that match a given keyword
// @Tags Security
// @Accept json
// @Produce json
// @Param keyword query string true "Search keyword"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /security/search-usernames [get]
func SearchUsernames(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get and validate search keyword
		keyword := c.Query("keyword")
		if keyword == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Search keyword is required"))
		}

		// Sanitize keyword (basic protection against SQL injection)
		keyword = strings.TrimSpace(keyword)
		if len(keyword) < 2 {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Keyword must be at least 2 characters"))
		}

		// Perform case-insensitive search
		var usernames []string
		searchPattern := "%" + strings.ToLower(keyword) + "%"
		if err := db.Model(&models.User{}).
			Select("username").
			Where("LOWER(username) LIKE ?", searchPattern).
			Limit(20).
			Pluck("username", &usernames).Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to search usernames: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"results": usernames,
		}, nil)
	}
}
