package account

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

type ProfileResponse struct {
	ID                uuid.UUID  `json:"id"`
	Email             string     `json:"email"`
	Username          string     `json:"username,omitempty"`
	Status            string     `json:"status,omitempty"`
	IsActive          bool       `json:"is_active"`
	FirstName         string     `json:"first_name"`
	LastName          string     `json:"last_name"`
	MiddleName        *string    `json:"middle_name,omitempty"`
	Gender            *string    `json:"gender,omitempty"`
	DateOfBirth       *time.Time `json:"date_of_birth,omitempty"`
	ProfileImage      string     `json:"profile_image"`
	CoverImage        string     `json:"cover_image"`
	AboutTheUser      *string    `json:"about_the_user,omitempty"`
	IsProfilePublic   bool       `json:"is_profile_public"`
	Nickname          *string    `json:"nickname,omitempty"`
	PreferredPronouns *string    `json:"preferred_pronouns,omitempty"`
	Location          struct {
		Country string `json:"country,omitempty"`
		State   string `json:"state,omitempty"`
		City    string `json:"city,omitempty"`
		Zip     string `json:"zip,omitempty"`
	} `json:"location,omitempty"`
	Roles []struct {
		RoleName string `json:"role_name"`
		IsActive bool   `json:"is_active"`
	} `json:"roles,omitempty"`
}

// ProfileHandler retrieves the profile information of the logged-in user
// @Summary Get the authenticated user's profile details
// @Description Fetch user profile details based on JWT authentication
// @Tags Account
// @Accept  json
// @Produce  json
// @Success 200 {object} ProfileResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /account/profile [get]
func ProfileHandler(db *gorm.DB, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authentication check with proper type assertion
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Use separate database connections for parallel queries
		dbUserDetail := db.Session(&gorm.Session{})
		dbUserLocation := db.Session(&gorm.Session{})
		dbSocialLinks := db.Session(&gorm.Session{})
		dbUserRoles := db.Session(&gorm.Session{})

		// Parallel execution with error channel
		errChan := make(chan error, 4)
		var wg sync.WaitGroup
		wg.Add(4)

		// Fetch user details
		var userDetail models.UserDetail
		go func() {
			defer wg.Done()
			if err := dbUserDetail.Model(&models.UserDetail{}).
				Select("first_name", "middle_name", "last_name", "gender", "date_of_birth",
					"profile_image", "cover_image", "about_the_user", "is_profile_public",
					"nickname", "preferred_pronouns").
				Where("user_id = ?", user.ID).
				First(&userDetail).Error; err != nil && err != gorm.ErrRecordNotFound {
				errChan <- fmt.Errorf("failed to fetch user details: %w", err)
			}
		}()

		// Fetch user location
		var userLocation models.UserLocation
		go func() {
			defer wg.Done()
			if err := dbUserLocation.Model(&models.UserLocation{}).
				Select("country", "state", "city", "zip").
				Where("user_id = ?", user.ID).
				First(&userLocation).Error; err != nil && err != gorm.ErrRecordNotFound {
				errChan <- fmt.Errorf("failed to fetch user location: %w", err)
			}
		}()

		// Fetch social links
		var socialLinks []models.SocialLink
		go func() {
			defer wg.Done()
			if err := dbSocialLinks.Model(&models.SocialLink{}).
				Select("platform", "link").
				Where("user_id = ?", user.ID).
				Find(&socialLinks).Error; err != nil {
				errChan <- fmt.Errorf("failed to fetch social links: %w", err)
			}
		}()

		// Fetch user roles with role names
		var userRoles []struct {
			RoleID   uuid.UUID
			RoleName string
			IsActive bool
		}
		go func() {
			defer wg.Done()
			if err := dbUserRoles.Model(&models.UserRole{}).
				Select("user_roles.role_id", "roles.role_name", "user_roles.is_active").
				Joins("JOIN roles ON roles.id = user_roles.role_id").
				Where("user_roles.user_id = ?", user.ID).
				Find(&userRoles).Error; err != nil {
				errChan <- fmt.Errorf("failed to fetch user roles: %w", err)
			}
		}()

		wg.Wait()
		close(errChan)

		// Check for errors
		for e := range errChan {
			if e != nil {
				return responseHandler.HandleResponse(c, nil, e)
			}
		}

		// Construct the profile response
		profileResponse := fiber.Map{
			"id":                 user.ID,
			"email":              user.Email,
			"username":           user.Username,
			"status":             user.Status,
			"is_active":          user.IsActive,
			"first_name":         userDetail.FirstName,
			"middle_name":        userDetail.MiddleName,
			"last_name":          userDetail.LastName,
			"gender":             userDetail.Gender,
			"date_of_birth":      userDetail.DateOfBirth,
			"profile_image":      userDetail.ProfileImage,
			"cover_image":        userDetail.CoverImage,
			"about_the_user":     userDetail.AboutTheUser,
			"is_profile_public":  userDetail.IsProfilePublic,
			"nickname":           userDetail.Nickname,
			"preferred_pronouns": userDetail.PreferredPronouns,
			"location": fiber.Map{
				"country": userLocation.Country,
				"state":   userLocation.State,
				"city":    userLocation.City,
				"zip":     userLocation.Zip,
			},
			"roles": make([]fiber.Map, 0, len(userRoles)),
		}

		// Populate roles
		for _, ur := range userRoles {
			profileResponse["roles"] = append(profileResponse["roles"].([]fiber.Map), fiber.Map{
				"role_name": ur.RoleName,
				"is_active": ur.IsActive,
			})
		}

		return responseHandler.HandleResponse(c, profileResponse, nil)
	}
}
