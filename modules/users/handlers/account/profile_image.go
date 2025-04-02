package account

import (
	"fmt"
	"log"

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
		userID := c.Locals("user").(models.User)

		// Validate image type
		imageType := c.FormValue("type") // "profile" or "cover"
		if imageType != "profile" && imageType != "cover" {
			return responseHandler.HandleResponse(c, map[string]interface{}{
				"success": false, "message": "Invalid image type, must be 'profile' or 'cover'",
			}, fmt.Errorf("invalid image type"))
		}

		// Handle image file
		file, err := c.FormFile("image")
		if err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{
				"success": false, "message": "No image file provided",
			}, fmt.Errorf("no image file provided"))
		}

		// Upload to Cloudinary
		fileURL, err := cld.UploadFile(file, userID.ID.String())
		if err != nil {
			log.Println("Failed to upload to Cloudinary:", err)
			return responseHandler.HandleResponse(c, map[string]interface{}{
				"success": false, "message": "Failed to upload image",
			}, fmt.Errorf("failed to upload image"))
		}

		// Check if user details exist
		var userDetail models.UserDetail
		if err := db.Where("user_id = ?", userID.ID).First(&userDetail).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create a new user detail record if not found
				userDetail = models.UserDetail{
					UserID: userID.ID,
				}
				if imageType == "profile" {
					userDetail.ProfileImage = fileURL
				} else {
					userDetail.CoverImage = fileURL
				}

				if err := db.Create(&userDetail).Error; err != nil {
					return responseHandler.HandleResponse(c, map[string]interface{}{
						"success": false, "message": "Failed to create user details",
					}, fmt.Errorf("failed to create user details"))
				}
			} else {
				return responseHandler.HandleResponse(c, map[string]interface{}{
					"success": false, "message": "User details lookup failed",
				}, err)
			}
		} else {
			// Update the correct image field
			if imageType == "profile" {
				userDetail.ProfileImage = fileURL
			} else {
				userDetail.CoverImage = fileURL
			}

			// Save user details to the database
			if err := db.Save(&userDetail).Error; err != nil {
				return responseHandler.HandleResponse(c, map[string]interface{}{
					"success": false, "message": "Failed to update user image",
				}, fmt.Errorf("failed to update user image"))
			}
		}

		// Return success response
		return responseHandler.HandleResponse(c, map[string]interface{}{
			"success": true, "message": "Image updated successfully", "url": fileURL,
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
		userID := c.Locals("user").(models.User)
		imageType := c.Query("type")
		// Validate image type
		if imageType != "profile" && imageType != "cover" {
			return responseHandler.HandleResponse(c, map[string]interface{}{
				"success": false, "message": "Invalid image type, must be 'profile' or 'cover'",
			}, fmt.Errorf("invalid image type, must be 'profile' or 'cover'"))
		}

		// Get user details from the database
		var userDetail models.UserDetail
		err := db.Where("user_id = ?", userID.ID).First(&userDetail).Error

		// If user details are not found, create a new one
		if err == gorm.ErrRecordNotFound {
			userDetail = models.UserDetail{UserID: userID.ID}
			if err := db.Create(&userDetail).Error; err != nil {
				return responseHandler.HandleResponse(c, map[string]interface{}{
					"success": false, "message": "Failed to create user details",
				}, fmt.Errorf("failed to create user details"))
			}
		} else if err != nil {
			// Handle other potential errors
			return responseHandler.HandleResponse(c, map[string]interface{}{
				"success": false, "message": "Error retrieving user details",
			}, err)
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

		// Save user details after removing image
		if err := db.Save(&userDetail).Error; err != nil {
			return responseHandler.HandleResponse(c, map[string]interface{}{
				"success": false, "message": "Failed to update database",
			}, fmt.Errorf("failed to update database"))
		}

		// Remove image from Cloudinary
		if filePath != "" {
			if err := cld.DeleteFile(filePath); err != nil {
				log.Println("Failed to delete from Cloudinary:", err)
				// Continue execution even if delete fails
			}
		}

		// Return success response
		return responseHandler.HandleResponse(c, map[string]interface{}{
			"success": true, "message": "Image removed successfully",
		}, nil)
	}
}
