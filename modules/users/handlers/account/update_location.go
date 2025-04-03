package account

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// UpdateLocationRequest defines the expected input for updating the location
type UpdateLocationRequest struct {
	Country string `json:"country,omitempty" validate:"omitempty,max=100"`
	State   string `json:"state,omitempty" validate:"omitempty,max=100"`
	City    string `json:"city,omitempty" validate:"omitempty,max=100"`
	Zip     string `json:"zip,omitempty" validate:"omitempty,max=20"`
}

// UpdateLocation updates the authenticated user's location
// @Summary Update the authenticated user's location
// @Description Update user location based on provided fields
// @Tags Account
// @Accept  json
// @Produce  json
// @Param   body  body  UpdateLocationRequest  true  "Location update data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /account/location [put]
func UpdateLocation(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authentication check with proper type assertion
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Parse request body
		var req UpdateLocationRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))
		}

		// Initialize validator with detailed error messages
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			errorMessages := make(map[string]string)

			for _, e := range validationErrors {
				field := strings.ToLower(e.Field())
				var message string

				switch e.Tag() {
				case "required":
					message = "This field is required"
				case "max":
					message = fmt.Sprintf("Cannot exceed %s characters", e.Param())
				default:
					message = fmt.Sprintf("Invalid %s value", field)
				}

				errorMessages[field] = message
			}

			return responseHandler.HandleResponse(c, fiber.Map{
				"message": "Validation failed",
				"errors":  errorMessages,
			}, fiber.NewError(fiber.StatusUnprocessableEntity, "Validation failed"))
		}

		// Start database transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Fetch or create UserLocation
		var userLocation models.UserLocation
		if err := tx.Where("user_id = ?", user.ID).First(&userLocation).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				userLocation = models.UserLocation{UserID: &user.ID}
			} else {
				tx.Rollback()
				return responseHandler.HandleResponse(c, nil,
					fmt.Errorf("failed to fetch location: %w", err))
			}
		}

		// Update fields
		if req.Country != "" {
			userLocation.Country = req.Country
		}
		if req.State != "" {
			userLocation.State = req.State
		}
		if req.City != "" {
			userLocation.City = req.City
		}
		if req.Zip != "" {
			userLocation.Zip = req.Zip
		}

		// Save changes
		if err := tx.Save(&userLocation).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update location: %w", err))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Location updated successfully",
		}, nil)
	}
}
