package account

import (
	"fmt"
	"strings"
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
		// Retrieve the user from context with proper type assertion
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Parse request body
		var req UpdateProfileRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))
		}

		// Initialize validator
		// Initialize validator
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			errorMessages := make(map[string]string)

			for _, e := range validationErrors {
				field := strings.ToLower(e.Field())
				var message string

				// Get the actual validation tag that failed (e.g., "max", "required")
				switch e.Tag() {
				case "required":
					message = "This field is required"
				case "max":
					switch field {
					case "firstname":
						message = "First name cannot exceed 100 characters"
					case "middlename":
						message = "Middle name cannot exceed 100 characters"
					case "lastname":
						message = "Last name cannot exceed 100 characters"
					case "gender":
						message = "Gender cannot exceed 50 characters"
					case "abouttheuser":
						message = "About section cannot exceed 1000 characters"
					case "nickname":
						message = "Nickname cannot exceed 100 characters"
					case "preferredpronouns":
						message = "Preferred pronouns cannot exceed 50 characters"
					default:
						message = fmt.Sprintf("Cannot exceed %s characters", e.Param())
					}
				case "datetime":
					if field == "dateofbirth" {
						message = "Must be in YYYY-MM-DD format"
					}
				default:
					message = fmt.Sprintf("Validation failed on %s rule", e.Tag())
				}

				// Convert field names to the JSON naming convention you use
				jsonField := strings.ToLower(e.Field())
				switch jsonField {
				case "firstname":
					jsonField = "first_name"
				case "middlename":
					jsonField = "middle_name"
				case "lastname":
					jsonField = "last_name"
				case "dateofbirth":
					jsonField = "date_of_birth"
				case "abouttheuser":
					jsonField = "about_the_user"
				case "preferredpronouns":
					jsonField = "preferred_pronouns"
				}

				errorMessages[jsonField] = message
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

		// Fetch or create UserDetail
		var userDetail models.UserDetail
		if err := tx.Where("user_id = ?", user.ID).First(&userDetail).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				userDetail = models.UserDetail{UserID: user.ID}
			} else {
				tx.Rollback()
				return responseHandler.HandleResponse(c, nil,
					fmt.Errorf("failed to fetch user details: %w", err))
			}
		}

		// Update fields
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
				tx.Rollback()
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusBadRequest, "Invalid date_of_birth format, use YYYY-MM-DD"))
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

		// Save changes
		if err := tx.Save(&userDetail).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update profile: %w", err))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Profile updated successfully",
		}, nil)
	}
}
