package collection

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/config"
	collectionModels "github.com/muga20/artsMarket/modules/artwork-management/models/collection"
	services "github.com/muga20/artsMarket/modules/artwork-management/services"
	userModels "github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// UpdateCollectionImagesHandler handles updating collection images
// @Summary Update collection images
// @Description Upload cover or primary image for a collection
// @Tags Collections
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Collection ID"
// @Param type formData string true "Image type (cover or primary)"
// @Param image formData file true "Image file"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections/{id}/images [put]
func UpdateCollectionImagesHandler(db *gorm.DB, cld *config.CloudinaryClient, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authenticated user
		user, ok := c.Locals("user").(userModels.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		collectionID := c.Params("id")
		if collectionID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Collection ID is required"))
		}

		// Validate collection ownership
		var collection collectionModels.Collection
		if err := db.Where("id = ? AND user_id = ?", collectionID, user.ID).First(&collection).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Collection not found or you don't have permission"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to retrieve collection: %w", err))
		}

		// Parse image type from form data
		var req CollectionImageRequest
		if err := c.BodyParser(&req); err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid request payload"))
		}

		// Get uploaded file
		file, err := c.FormFile("image")
		if err != nil {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "No image file provided"))
		}

		// Validate file
		if err := services.ValidateImageFile(file); err != nil {
			return responseHandler.HandleResponse(c, nil, err)
		}

		// Upload to Cloudinary
		fileURL, err := cld.UploadFile(file, fmt.Sprintf("collections/%s", collectionID))
		if err != nil {
			log.Printf("Cloudinary upload failed: %v", err)
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to upload image"))
		}

		// Update collection with new image URL
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Update appropriate image field based on type
		switch req.ImageType {
		case "cover":
			collection.CoverImageURL = fileURL
		case "primary":
			collection.PrimaryImageURL = fileURL
		}
		collection.UpdatedAt = time.Now()

		if err := tx.Save(&collection).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update collection image: %w", err))
		}

		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": fmt.Sprintf("%s image updated successfully", req.ImageType),
		}, nil)
	}
}

// RemoveCollectionImageHandler handles removing collection images
// @Summary Remove collection image
// @Description Completely removes either cover or primary image from a collection
// @Tags Collections
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Collection ID"
// @Param type query string true "Image type to remove (cover or primary)"
// @Success 200 {object} map[string]interface{} "Returns success message"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections/{id}/images [delete]
func RemoveCollectionImageHandler(db *gorm.DB, cld *config.CloudinaryClient, responseHandler *handlers.ResponseHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authenticated user
		user, ok := c.Locals("user").(userModels.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		collectionID := c.Params("id")
		if collectionID == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Collection ID is required"))
		}

		imageType := c.Query("type")
		if imageType == "" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Image type is required (cover or primary)"))
		}

		if imageType != "cover" && imageType != "primary" {
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, "Invalid image type. Must be 'cover' or 'primary'"))
		}

		// Validate collection ownership
		var collection collectionModels.Collection
		if err := db.Where("id = ? AND user_id = ?", collectionID, user.ID).First(&collection).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return responseHandler.HandleResponse(c, nil,
					fiber.NewError(fiber.StatusNotFound, "Collection not found or you don't have permission"))
			}
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to retrieve collection: %w", err))
		}

		// Start transaction
		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to start transaction: %w", tx.Error))
		}

		// Get current image URL before deletion
		var imageURL string
		if imageType == "cover" {
			imageURL = collection.CoverImageURL
			collection.CoverImageURL = ""
		} else {
			imageURL = collection.PrimaryImageURL
			collection.PrimaryImageURL = ""
		}

		// If there was no image to begin with
		if imageURL == "" {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Collection has no %s image to remove", imageType)))
		}

		// Update collection in database
		collection.UpdatedAt = time.Now()
		if err := tx.Save(&collection).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to update collection: %w", err))
		}

		// Delete from Cloudinary
		if err := cld.DeleteFile(imageURL); err != nil {
			tx.Rollback()
			log.Printf("Failed to delete image from Cloudinary: %v", err)
			return responseHandler.HandleResponse(c, nil,
				fiber.NewError(fiber.StatusInternalServerError, "Failed to remove image from storage"))
		}

		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil,
				fmt.Errorf("failed to commit transaction: %w", err))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": fmt.Sprintf("%s image removed successfully", imageType),
		}, nil)
	}
}
