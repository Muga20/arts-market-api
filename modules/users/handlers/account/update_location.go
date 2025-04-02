package account

import (
	"errors"

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
		// Retrieve the user from the context set by AuthMiddleware
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.Handle(c, nil, errors.New("user not found in context"))
		}

		// Parse the request body
		var req UpdateLocationRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.Handle(c, nil, errors.New("invalid request body"))
		}

		// Initialize validator
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			errorMessages := make(map[string]string)
			for _, e := range validationErrors {
				switch e.Field() {
				case "Country":
					errorMessages["country"] = "Country must be less than 100 characters"
				case "State":
					errorMessages["state"] = "State must be less than 100 characters"
				case "City":
					errorMessages["city"] = "City must be less than 100 characters"
				case "Zip":
					errorMessages["zip"] = "Zip must be less than 20 characters"
				}
			}
			return responseHandler.Handle(c, map[string]interface{}{
				"error":  "validation failed",
				"fields": errorMessages,
			}, errors.New("validation failed"))
		}

		// Fetch or create UserLocation
		var userLocation models.UserLocation
		if err := db.Where("user_id = ?", user.ID).First(&userLocation).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				userLocation = models.UserLocation{UserID: &user.ID} // Initialize with UserID
			} else {
				return responseHandler.Handle(c, nil, err)
			}
		}

		// Update UserLocation fields if provided
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

		// Save UserLocation (create or update)
		if err := db.Save(&userLocation).Error; err != nil {
			return responseHandler.Handle(c, nil, err)
		}

		// Return success message
		return responseHandler.Handle(c, map[string]string{
			"message": "updated successfully",
		}, nil)
	}
}
