package account

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// UpdateUserImage uploads a new profile or cover image to Cloudinary
// @Summary Update user's profile or cover image
// @Description Uploads a new profile or cover image to Cloudinary and updates the database
// @Tags Account
// @Accept  multipart/form-data
// @Produce  json
// @Param type formData string true "Image type (profile or cover)"
// @Param image formData file true "User image"
// @Success 200 {object} map[string]interface{} "Image updated successfully"
// @Failure 400 {object} map[string]string "Invalid image type or no image file provided"
// @Failure 404 {object} map[string]string "User details not found"
// @Failure 500 {object} map[string]string "Failed to upload or update image"
// @Router /account/image [put]
func UpdateUserImage(db *gorm.DB, cld *config.CloudinaryClient, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authentication check with proper type assertion
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Validate image type
		imageType := c.FormValue("type") // "profile" or "cover"
		if imageType != "profile" && imageType != "cover" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid image type, must be 'profile' or 'cover'"))
		}

		// Handle image file
		file, err := c.FormFile("image")
		if err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "No image file provided"))
		}

		// Validate file size and type
		if file.Size > 5<<20 { // 5MB limit
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Image too large, maximum size is 5MB"))
		}

		if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Only image files are allowed"))
		}

		// Upload to Cloudinary
		fileURL, err := cld.UploadFile(file, user.ID.String())
		if err != nil {
			log.Printf("Cloudinary upload failed: %v", err)
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to upload image"))
		}

		// Start database transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Check if user details exist
		var userDetail models.UserDetail
		err = tx.Where("user_id = ?", user.ID).First(&userDetail).Error

		if err == gorm.ErrRecordNotFound {
			// Create new record if not found
			userDetail = models.UserDetail{
				UserID: user.ID,
			}
			if imageType == "profile" {
				userDetail.ProfileImage = fileURL
			} else {
				userDetail.CoverImage = fileURL
			}

			if err := tx.Create(&userDetail).Error; err != nil {
				tx.Rollback()
				return responseHandler.HandleResponse(c, nil,
					fmt.Errorf("failed to create user details: %w", err))
			}
		} else if err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to fetch user details: %w", err))
		} else {
			// Update existing record
			if imageType == "profile" {
				userDetail.ProfileImage = fileURL
			} else {
				userDetail.CoverImage = fileURL
			}

			if err := tx.Save(&userDetail).Error; err != nil {
				tx.Rollback()
				return responseHandler.HandleResponse(c, nil,
					fmt.Errorf("failed to update user image: %w", err))
			}
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Image updated successfully",
			"url":     fileURL,
		}, nil)
	}
}

// RemoveUserImage deletes a user's image from Cloudinary
// @Summary Remove user's profile or cover image
// @Description Deletes a user's profile or cover image from Cloudinary and updates the database
// @Tags Account
// @Accept  json
// @Produce  json
// @Param type query string true "Image type (profile or cover)"
// @Success 200 {object} map[string]interface{} "Image removed successfully"
// @Failure 400 {object} map[string]string "Invalid image type"
// @Failure 404 {object} map[string]string "User details not found"
// @Failure 500 {object} map[string]string "Failed to delete image"
// @Router /account/image [delete]
func RemoveUserImage(db *gorm.DB, cld *config.CloudinaryClient, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authentication check with proper type assertion
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		// Validate image type
		imageType := c.Query("type")
		if imageType != "profile" && imageType != "cover" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid image type, must be 'profile' or 'cover'"))
		}

		// Start database transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Get or create user details
		var userDetail models.UserDetail
		err := tx.Where("user_id = ?", user.ID).First(&userDetail).Error

		if err == gorm.ErrRecordNotFound {
			// Create new record if not found
			userDetail = models.UserDetail{UserID: user.ID}
			if err := tx.Create(&userDetail).Error; err != nil {
				tx.Rollback()
				return responseHandler.HandleResponse(c, nil,
					fmt.Errorf("failed to create user details: %w", err))
			}
		} else if err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to fetch user details: %w", err))
		}

		// Determine which image field to clear
		var filePath string
		if imageType == "profile" {
			filePath = userDetail.ProfileImage
			userDetail.ProfileImage = ""
		} else {
			filePath = userDetail.CoverImage
			userDetail.CoverImage = ""
		}

		// Save updated user details
		if err := tx.Save(&userDetail).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update user details: %w", err))
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		// Delete from Cloudinary (non-blocking)
		if filePath != "" {
			go func() {
				if err := cld.DeleteFile(filePath); err != nil {
					log.Printf("Failed to delete image from Cloudinary: %v (path: %s)", err, filePath)
				}
			}()
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Image removed successfully",
		}, nil)
	}
}
