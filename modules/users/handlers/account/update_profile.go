package account

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// UpdateProfileRequest defines the expected input for updating the profile
type UpdateProfileRequest struct {
	FirstName         string  `json:"first_name,omitempty" validate:"omitempty,max=100"`
	MiddleName        *string `json:"middle_name,omitempty" validate:"omitempty,max=100"`
	LastName          string  `json:"last_name,omitempty" validate:"omitempty,max=100"`
	Gender            *string `json:"gender,omitempty" validate:"omitempty,max=50"`
	DateOfBirth       *string `json:"date_of_birth,omitempty" validate:"omitempty,datetime=2006-01-02"`
	AboutTheUser      *string `json:"about_the_user,omitempty" validate:"omitempty,max=1000"`
	IsProfilePublic   *bool   `json:"is_profile_public,omitempty"`
	Nickname          *string `json:"nickname,omitempty" validate:"omitempty,max=100"`
	PreferredPronouns *string `json:"preferred_pronouns,omitempty" validate:"omitempty,max=50"`
}

// UpdateProfile updates the authenticated user's profile details
// @Summary Update the authenticated user's profile
// @Description Update user details based on provided fields (excluding email, username, and images)
// @Tags Account
// @Accept  json
// @Produce  json
// @Param   body  body  UpdateProfileRequest  true  "Profile update data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /account/profile [put]
func UpdateProfile(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Retrieve the user from the context set by AuthMiddleware
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.Handle(c, nil, errors.New("user not found in context"))
		}

		// Parse the request body
		var req UpdateProfileRequest
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
				case "FirstName":
					errorMessages["first_name"] = "First name must be less than 100 characters"
				case "MiddleName":
					errorMessages["middle_name"] = "Middle name must be less than 100 characters"
				case "LastName":
					errorMessages["last_name"] = "Last name must be less than 100 characters"
				case "Gender":
					errorMessages["gender"] = "Gender must be less than 50 characters"
				case "DateOfBirth":
					errorMessages["date_of_birth"] = "Date of birth must be in YYYY-MM-DD format"
				case "AboutTheUser":
					errorMessages["about_the_user"] = "About the user must be less than 1000 characters"
				case "Nickname":
					errorMessages["nickname"] = "Nickname must be less than 100 characters"
				case "PreferredPronouns":
					errorMessages["preferred_pronouns"] = "Preferred pronouns must be less than 50 characters"
				}
			}
			return responseHandler.Handle(c, map[string]interface{}{
				"error":  "validation failed",
				"fields": errorMessages,
			}, errors.New("validation failed"))
		}

		// Fetch or create UserDetail
		var userDetail models.UserDetail
		if err := db.Where("user_id = ?", user.ID).First(&userDetail).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				userDetail = models.UserDetail{UserID: user.ID} // Initialize with UserID
			} else {
				return responseHandler.Handle(c, nil, err)
			}
		}

		// Update UserDetail fields if provided
		if req.FirstName != "" {
			userDetail.FirstName = req.FirstName
		}
		if req.MiddleName != nil {
			userDetail.MiddleName = req.MiddleName
		}
		if req.LastName != "" {
			userDetail.LastName = req.LastName
		}
		if req.Gender != nil {
			userDetail.Gender = req.Gender
		}
		if req.DateOfBirth != nil {
			if parsedDate, err := time.Parse("2006-01-02", *req.DateOfBirth); err == nil {
				userDetail.DateOfBirth = &parsedDate
			} else {
				return responseHandler.Handle(c, nil, errors.New("invalid date_of_birth format, use YYYY-MM-DD"))
			}
		}
		if req.AboutTheUser != nil {
			userDetail.AboutTheUser = req.AboutTheUser
		}
		if req.IsProfilePublic != nil {
			userDetail.IsProfilePublic = *req.IsProfilePublic
		}
		if req.Nickname != nil {
			userDetail.Nickname = req.Nickname
		}
		if req.PreferredPronouns != nil {
			userDetail.PreferredPronouns = req.PreferredPronouns
		}

		// Save UserDetail (create or update)
		if err := db.Save(&userDetail).Error; err != nil {
			return responseHandler.Handle(c, nil, err)
		}

		// Return success message
		return responseHandler.Handle(c, map[string]string{
			"message": "updated successfully",
		}, nil)
	}
}
