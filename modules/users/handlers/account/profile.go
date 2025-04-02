package account

import (
	"errors"
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
	SocialLinks []struct {
		Platform string `json:"platform"`
		Link     string `json:"link"`
	} `json:"social_links,omitempty"`
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
		// Retrieve the user from the context set by the AuthMiddleware
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.Handle(c, nil, errors.New("user not found in context"))
		}

		// Fetch UserDetail, UserLocation, SocialLinks, and UserRoles with selective attributes
		var userDetail models.UserDetail
		if err := db.Model(&models.UserDetail{}).
			Select("id", "user_id", "first_name", "middle_name", "last_name", "gender", "date_of_birth",
				"profile_image", "cover_image", "about_the_user", "is_profile_public", "nickname", "preferred_pronouns").
			Where("user_id = ?", user.ID).
			First(&userDetail).Error; err != nil && err != gorm.ErrRecordNotFound {
			return responseHandler.Handle(c, nil, err)
		}

		var userLocation models.UserLocation
		if err := db.Model(&models.UserLocation{}).
			Select("country", "state", "city", "zip").
			Where("user_id = ?", user.ID).
			First(&userLocation).Error; err != nil && err != gorm.ErrRecordNotFound {
			return responseHandler.Handle(c, nil, err)
		}

		var socialLinks []models.SocialLink
		if err := db.Model(&models.SocialLink{}).
			Select("platform", "link").
			Where("user_id = ?", user.ID).
			Find(&socialLinks).Error; err != nil {
			return responseHandler.Handle(c, nil, err)
		}

		var userRoles []models.UserRole
		if err := db.Model(&models.UserRole{}).
			Select("user_roles.role_id", "user_roles.is_active", "roles.role_name").
			Joins("JOIN roles ON roles.id = user_roles.role_id").
			Where("user_roles.user_id = ?", user.ID).
			Find(&userRoles).Error; err != nil {
			return responseHandler.Handle(c, nil, err)
		}

		// Construct the profile response
		profileResponse := ProfileResponse{
			ID:                user.ID,
			Email:             user.Email,
			Username:          user.Username,
			Status:            user.Status,
			IsActive:          user.IsActive,
			FirstName:         userDetail.FirstName,
			MiddleName:        userDetail.MiddleName,
			LastName:          userDetail.LastName,
			Gender:            userDetail.Gender,
			DateOfBirth:       userDetail.DateOfBirth,
			ProfileImage:      userDetail.ProfileImage,
			CoverImage:        userDetail.CoverImage,
			AboutTheUser:      userDetail.AboutTheUser,
			IsProfilePublic:   userDetail.IsProfilePublic,
			Nickname:          userDetail.Nickname,
			PreferredPronouns: userDetail.PreferredPronouns,
			Location: struct {
				Country string `json:"country,omitempty"`
				State   string `json:"state,omitempty"`
				City    string `json:"city,omitempty"`
				Zip     string `json:"zip,omitempty"`
			}{
				Country: userLocation.Country,
				State:   userLocation.State,
				City:    userLocation.City,
				Zip:     userLocation.Zip,
			},
		}

		// Populate social links
		for _, sl := range socialLinks {
			profileResponse.SocialLinks = append(profileResponse.SocialLinks, struct {
				Platform string `json:"platform"`
				Link     string `json:"link"`
			}{
				Platform: sl.Platform,
				Link:     sl.Link,
			})
		}

		// Populate roles
		for _, ur := range userRoles {
			var role models.Role
			if err := db.Model(&models.Role{}).
				Select("role_name").
				Where("id = ?", ur.RoleID).
				First(&role).Error; err != nil {
				return responseHandler.Handle(c, nil, err)
			}
			profileResponse.Roles = append(profileResponse.Roles, struct {
				RoleName string `json:"role_name"`
				IsActive bool   `json:"is_active"`
			}{
				RoleName: role.RoleName,
				IsActive: ur.IsActive,
			})
		}

		// Return the profile response
		return responseHandler.Handle(c, profileResponse, nil)
	}
}
