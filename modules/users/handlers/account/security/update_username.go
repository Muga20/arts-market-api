package security

import (
	"errors"
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
		// Get the user from context (assumed to be set by authentication middleware)
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.Handle(c, nil, errors.New("user not found in context"))
		}

		// Parse the request body for the new username
		var req UpdateUsernameRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.Handle(c, nil, errors.New("invalid request body"))
		}

		// Validate the new username
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			return responseHandler.Handle(c, map[string]string{"error": "validation failed"}, err)
		}

		// Check if the new username is the same as the current one
		if req.Username == user.Username {
			return responseHandler.Handle(c, nil, errors.New("new username cannot be the same as the current username"))
		}

		// Check if the new username is already taken by another user
		var existingUser models.User
		if err := db.Where("username = ? AND id != ?", req.Username, user.ID).First(&existingUser).Error; err != gorm.ErrRecordNotFound {
			if err == nil {
				return responseHandler.Handle(c, nil, errors.New("username is already in use"))
			}
			return responseHandler.Handle(c, nil, err)
		}

		// Update username
		user.Username = req.Username
		if err := db.Save(&user).Error; err != nil {
			return responseHandler.Handle(c, nil, err)
		}

		// Return success message
		return responseHandler.Handle(c, map[string]string{"message": "username updated successfully"}, nil)
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
		keyword := c.Query("keyword")
		if keyword == "" {
			return responseHandler.Handle(c, nil, errors.New("keyword is required"))
		}

		var usernames []string
		searchPattern := "%" + strings.ToLower(keyword) + "%"
		if err := db.Model(&models.User{}).Where("LOWER(username) LIKE ?", searchPattern).Pluck("username", &usernames).Error; err != nil {
			return responseHandler.Handle(c, nil, err)
		}

		return responseHandler.Handle(c, usernames, nil)
	}
}
