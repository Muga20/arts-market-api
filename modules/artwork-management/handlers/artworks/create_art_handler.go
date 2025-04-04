package artworks

import (
	"context"
	"fmt"
	"mime/multipart"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/muga20/artsMarket/config"
	models "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	attributes "github.com/muga20/artsMarket/modules/artwork-management/models/arttributes"
	category "github.com/muga20/artsMarket/modules/artwork-management/models/category"
	collection "github.com/muga20/artsMarket/modules/artwork-management/models/collection"
	tag "github.com/muga20/artsMarket/modules/artwork-management/models/tags"
	"github.com/muga20/artsMarket/modules/artwork-management/repository"
	image_handler "github.com/muga20/artsMarket/modules/artwork-management/services"
	user_details "github.com/muga20/artsMarket/modules/users/models"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// AttributeInput represents a single attribute with ID and value
type AttributeInput struct {
	ID    string `form:"id"`    // Attribute ID as string
	Value string `form:"value"` // Attribute value
}

// CreateArtworkRequest represents the form fields for creating an artwork
type CreateArtworkRequest struct {
	// Basic Information
	Title       string `form:"title" validate:"required"`
	Description string `form:"description"`

	// Classification
	Type       string   `form:"type" validate:"required,oneof=traditional digital photography mixed_media sculpture performance"`
	Categories []string `form:"categories"` // Comma-separated category IDs
	Tags       []string `form:"tags"`       // Comma-separated tag IDs

	// Ownership & Collection
	CollectionID string `form:"collection_id"`

	// Physical Details
	Dimensions string  `form:"dimensions"`
	Weight     float64 `form:"weight"`
	IsFramed   bool    `form:"is_framed"`
	Condition  string  `form:"condition" validate:"omitempty,oneof=pristine excellent good acceptable restored damaged"`

	// Creation Details
	CreationDate string `form:"creation_date" validate:"omitempty,datetime=2006-01-02"`
	MediumID     string `form:"medium_id"`
	TechniqueID  string `form:"technique_id"`

	// Pricing & Availability
	Price     float64 `form:"price"`
	IsForSale bool    `form:"is_for_sale"`

	// Licensing
	LicenseType    string `form:"license_type" validate:"omitempty,oneof=all_rights_reserved creative_commons public_domain exclusive_license non_exclusive_license custom"`
	LicenseDetails string `form:"license_details"`

	// Edition Information
	EditionNumber int `form:"edition_number"`
	TotalEditions int `form:"total_editions"`

	// Custom Attributes
	Attributes []AttributeInput `form:"attributes"` // Handled specially in parser

	// Images - handled separately in multipart form
}

// ArtworkResponse represents the API response for artwork creation
type ArtworkResponse struct {
	Message string         `json:"message"`
	Artwork models.Artwork `json:"artwork"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateArtworkHandler godoc
// @Summary Create a new artwork
// @Description Creates a new artwork with all related data including images, tags, categories, and attributes
// @Tags Artworks
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param title formData string true "Artwork title"
// @Param description formData string false "Artwork description"
// @Param type formData string true "Artwork type" Enums(traditional,digital,photography,mixed_media,sculpture,performance)
// @Param categories formData string false "Comma-separated category IDs (e.g., 'id1,id2,id3')"
// @Param tags formData []string false "Array of tag IDs (format: tags[0]=id1, tags[1]=id2)"
// @Param collection_id formData string false "Collection ID"
// @Param dimensions formData string false "Dimensions"
// @Param weight formData number false "Weight in kg"
// @Param is_framed formData boolean false "Is framed"
// @Param condition formData string false "Condition" Enums(pristine,excellent,good,acceptable,restored,damaged)
// @Param creation_date formData string false "Creation date (YYYY-MM-DD)"
// @Param medium_id formData string false "Medium ID"
// @Param technique_id formData string false "Technique ID"
// @Param price formData number false "Price"
// @Param is_for_sale formData boolean false "Is for sale"
// @Param license_type formData string false "License type" Enums(all_rights_reserved,creative_commons,public_domain,exclusive_license,non_exclusive_license,custom)
// @Param license_details formData string false "License details"
// @Param edition_number formData integer false "Edition number"
// @Param total_editions formData integer false "Total editions"
// @Param attributes formData []string false "Array of attribute objects"
// @Param images formData file false "Artwork images (multiple allowed)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /artworks [post]
func CreateArtworkHandler(db *gorm.DB, cld *config.CloudinaryClient, responseHandler *handlers.ResponseHandler, artworkRepo repository.ArtworkRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithCancel(c.Context())
		defer cancel()

		user, ok := c.Locals("user").(user_details.User)
		if !ok {
			return responseHandler.HandleResponse(c, nil, fiber.NewError(fiber.StatusUnauthorized, "Authentication required"))
		}

		form, err := c.MultipartForm()
		if err != nil {
			return responseHandler.HandleResponse(c, nil, fiber.NewError(fiber.StatusBadRequest, "Invalid form data"))
		}

		var req CreateArtworkRequest
		if err := parseFormData(form, &req); err != nil {
			return responseHandler.HandleResponse(c, nil, fiber.NewError(fiber.StatusBadRequest, err.Error()))
		}

		if err := validateArtworkRequest(&req); err != nil {
			return responseHandler.HandleResponse(c, nil, fiber.NewError(fiber.StatusBadRequest, err.Error()))
		}

		tx := db.Begin()
		if tx.Error != nil {
			return responseHandler.HandleResponse(c, nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to start transaction"))
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Check existing artworks count
		var existingArtworkCount int64
		if err := tx.Model(&models.Artwork{}).Where("user_id = ?", user.ID).Count(&existingArtworkCount).Error; err != nil {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to check existing artworks"))
		}

		artwork, err := createArtwork(ctx, tx, user.ID, req, responseHandler)
		if err != nil {
			tx.Rollback()
			return err
		}

		var wg sync.WaitGroup
		errChan := make(chan error, 5)

		// Process editions
		if req.EditionNumber > 0 && req.TotalEditions > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := processEditions(ctx, tx, artwork.ID, req.EditionNumber, req.TotalEditions); err != nil {
					errChan <- err
				}
			}()
		}

		// Process tags in bulk
		if len(req.Tags) > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := processTags(ctx, tx, artwork.ID, req.Tags); err != nil {
					errChan <- err
				}
			}()
		}

		// Process categories in bulk
		if len(req.Categories) > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := processCategories(ctx, tx, artwork.ID, req.Categories); err != nil {
					errChan <- err
				}
			}()
		}

		// Process attributes in bulk
		if len(req.Attributes) > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := processAttributes(ctx, tx, artwork.ID, req.Attributes); err != nil {
					errChan <- err
				}
			}()
		}

		// Process images concurrently with bulk insert
		if images := form.File["images"]; len(images) > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := processImages(ctx, tx, cld, artwork.ID, images); err != nil {
					errChan <- err
				}
			}()
		}

		go func() {
			wg.Wait()
			close(errChan)
		}()

		for err := range errChan {
			tx.Rollback()
			return responseHandler.HandleResponse(c, nil, fiber.NewError(fiber.StatusInternalServerError, err.Error()))
		}

		if err := tx.Commit().Error; err != nil {
			return responseHandler.HandleResponse(c, nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to commit transaction"))
		}

		return responseHandler.HandleResponse(c, fiber.Map{
			"message": "Artwork created successfully",
		}, nil)
	}
}

// parseFormData parses the multipart form data into the request struct
func parseFormData(form *multipart.Form, req *CreateArtworkRequest) error {
	// Handle simple string values
	if titles := form.Value["title"]; len(titles) > 0 {
		req.Title = titles[0]
	}
	if descriptions := form.Value["description"]; len(descriptions) > 0 {
		req.Description = descriptions[0]
	}
	if types := form.Value["type"]; len(types) > 0 {
		req.Type = types[0]
	}
	if collectionIDs := form.Value["collection_id"]; len(collectionIDs) > 0 {
		req.CollectionID = collectionIDs[0]
	}
	if dimensions := form.Value["dimensions"]; len(dimensions) > 0 {
		req.Dimensions = dimensions[0]
	}
	if conditions := form.Value["condition"]; len(conditions) > 0 {
		req.Condition = conditions[0]
	}
	if creationDates := form.Value["creation_date"]; len(creationDates) > 0 {
		req.CreationDate = creationDates[0]
	}
	if mediumIDs := form.Value["medium_id"]; len(mediumIDs) > 0 {
		req.MediumID = mediumIDs[0]
	}
	if techniqueIDs := form.Value["technique_id"]; len(techniqueIDs) > 0 {
		req.TechniqueID = techniqueIDs[0]
	}
	if licenseTypes := form.Value["license_type"]; len(licenseTypes) > 0 {
		req.LicenseType = licenseTypes[0]
	}
	if licenseDetails := form.Value["license_details"]; len(licenseDetails) > 0 {
		req.LicenseDetails = licenseDetails[0]
	}

	// Handle comma-separated lists
	if categories := form.Value["categories"]; len(categories) > 0 {
		req.Categories = strings.Split(categories[0], ",")
	}
	if tags := form.Value["tags"]; len(tags) > 0 {
		req.Tags = strings.Split(tags[0], ",")
	}

	// Handle numeric values
	if weights := form.Value["weight"]; len(weights) > 0 {
		if weight, err := strconv.ParseFloat(weights[0], 64); err == nil {
			req.Weight = weight
		}
	}
	if prices := form.Value["price"]; len(prices) > 0 {
		if price, err := strconv.ParseFloat(prices[0], 64); err == nil {
			req.Price = price
		}
	}
	if isForSales := form.Value["is_for_sale"]; len(isForSales) > 0 {
		req.IsForSale = strings.ToLower(isForSales[0]) == "true"
	}
	if isFrameds := form.Value["is_framed"]; len(isFrameds) > 0 {
		req.IsFramed = strings.ToLower(isFrameds[0]) == "true"
	}
	if editionNumbers := form.Value["edition_number"]; len(editionNumbers) > 0 {
		if editionNumber, err := strconv.Atoi(editionNumbers[0]); err == nil {
			req.EditionNumber = editionNumber
		}
	}
	if totalEditions := form.Value["total_editions"]; len(totalEditions) > 0 {
		if totalEdition, err := strconv.Atoi(totalEditions[0]); err == nil {
			req.TotalEditions = totalEdition
		}
	}

	// Handle attributes (format: attributes[0][id]=xxx&attributes[0][value]=yyy)
	for i := 0; ; i++ {
		idKey := fmt.Sprintf("attributes[%d][id]", i)
		valueKey := fmt.Sprintf("attributes[%d][value]", i)

		if ids, ok := form.Value[idKey]; !ok || len(ids) == 0 {
			break
		}

		id := form.Value[idKey][0]
		value := ""
		if values, ok := form.Value[valueKey]; ok && len(values) > 0 {
			value = values[0]
		}

		req.Attributes = append(req.Attributes, AttributeInput{
			ID:    id,
			Value: value,
		})
	}

	return nil
}

// validateArtworkRequest validates the artwork request
func validateArtworkRequest(req *CreateArtworkRequest) error {
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if req.Type == "" {
		return fmt.Errorf("type is required")
	}
	if req.CreationDate != "" {
		if _, err := time.Parse("2006-01-02", req.CreationDate); err != nil {
			return fmt.Errorf("invalid creation date format, use YYYY-MM-DD")
		}
	}
	return nil
}

// createArtwork creates the base artwork record
func createArtwork(ctx context.Context, tx *gorm.DB, userID uuid.UUID, req CreateArtworkRequest, responseHandler *handlers.ResponseHandler) (*models.Artwork, error) {
	collectionID, err := parseCollectionID(ctx, tx, req.CollectionID, userID)
	if err != nil {
		return nil, err
	}

	var creationDate *time.Time
	if req.CreationDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.CreationDate)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid creation date format")
		}
		creationDate = &parsedDate
	}

	artwork := &models.Artwork{
		UserID:         userID,
		Title:          req.Title,
		Description:    req.Description,
		CollectionID:   collectionID,
		CreationDate:   creationDate,
		Price:          &req.Price,
		IsForSale:      req.IsForSale,
		Type:           models.ArtworkType(req.Type),
		Dimensions:     req.Dimensions,
		Weight:         &req.Weight,
		IsFramed:       req.IsFramed,
		Condition:      models.ConditionType(req.Condition),
		MediumID:       parseUUIDPointer(req.MediumID),
		TechniqueID:    parseUUIDPointer(req.TechniqueID),
		LicenseType:    models.LicenseType(req.LicenseType),
		LicenseDetails: req.LicenseDetails,
	}

	if err := tx.Create(artwork).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to create artwork")
	}

	return artwork, nil
}

// processEditions handles edition creation
func processEditions(ctx context.Context, tx *gorm.DB, artworkID uuid.UUID, editionNumber, totalEditions int) error {
	edition := models.Edition{
		ArtworkID:     artworkID,
		EditionNumber: editionNumber,
		TotalEditions: totalEditions,
		Status:        "available",
	}

	return tx.Create(&edition).Error
}

func processTags(ctx context.Context, tx *gorm.DB, artworkID uuid.UUID, tagIDs []string) error {
	artworkTags := make([]tag.ArtworkTag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		tagUUID, err := uuid.Parse(tagID)
		if err != nil {
			return fmt.Errorf("invalid tag ID format: %s", tagID)
		}
		artworkTags = append(artworkTags, tag.ArtworkTag{
			ArtworkID: artworkID,
			TagID:     tagUUID,
		})
	}
	if len(artworkTags) > 0 {
		return tx.Create(&artworkTags).Error
	}
	return nil
}

func processCategories(ctx context.Context, tx *gorm.DB, artworkID uuid.UUID, categoryIDs []string) error {
	artworkCategories := make([]category.ArtworkCategory, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		categoryUUID, err := uuid.Parse(categoryID)
		if err != nil {
			return fmt.Errorf("invalid category ID format: %s", categoryID)
		}
		artworkCategories = append(artworkCategories, category.ArtworkCategory{
			ArtworkID:  artworkID,
			CategoryID: categoryUUID,
		})
	}
	if len(artworkCategories) > 0 {
		return tx.Create(&artworkCategories).Error
	}
	return nil
}

func processAttributes(ctx context.Context, tx *gorm.DB, artworkID uuid.UUID, attrs []AttributeInput) error {
	artworkAttributes := make([]attributes.ArtworkAttribute, 0, len(attrs))
	for _, attr := range attrs {
		attrID, err := uuid.Parse(attr.ID)
		if err != nil {
			return fmt.Errorf("invalid attribute ID format: %s", attr.ID)
		}
		artworkAttributes = append(artworkAttributes, attributes.ArtworkAttribute{
			ArtworkID:   artworkID,
			AttributeID: attrID,
			Value:       attr.Value,
		})
	}
	if len(artworkAttributes) > 0 {
		return tx.Create(&artworkAttributes).Error
	}
	return nil
}

func processImages(ctx context.Context, tx *gorm.DB, cld *config.CloudinaryClient, artworkID uuid.UUID, files []*multipart.FileHeader) error {
	imageURLs, err := uploadImagesConcurrently(ctx, cld, artworkID, files)
	if err != nil {
		return err
	}

	artworkImages := make([]models.ArtworkImage, len(imageURLs))
	for i, url := range imageURLs {
		artworkImages[i] = models.ArtworkImage{
			ArtworkID: artworkID,
			ImageURL:  url,
			IsPrimary: i == 0,
		}
	}
	return tx.Create(&artworkImages).Error
}

func uploadImagesConcurrently(ctx context.Context, cld *config.CloudinaryClient, artworkID uuid.UUID, files []*multipart.FileHeader) ([]string, error) {
	numWorkers := runtime.NumCPU()
	if numWorkers > 4 {
		numWorkers = 4
	}

	type job struct {
		index int
		file  *multipart.FileHeader
	}

	type result struct {
		index int
		url   string
		err   error
	}

	jobs := make(chan job, len(files))
	results := make(chan result, len(files))

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				if err := image_handler.ValidateImageFile(j.file); err != nil {
					results <- result{index: j.index, err: err}
					continue
				}
				url, err := cld.UploadFile(j.file, fmt.Sprintf("artworks/%s", artworkID.String()))
				if err != nil {
					results <- result{index: j.index, err: fmt.Errorf("failed to upload image: %v", err)}
					continue
				}
				results <- result{index: j.index, url: url}
			}
		}()
	}

	for i, file := range files {
		jobs <- job{index: i, file: file}
	}
	close(jobs)
	wg.Wait()
	close(results)

	urls := make([]string, len(files))
	for res := range results {
		if res.err != nil {
			return nil, res.err
		}
		urls[res.index] = res.url
	}
	return urls, nil
}

// parseCollectionID validates and parses collection ID
func parseCollectionID(ctx context.Context, tx *gorm.DB, collectionIDStr string, userID uuid.UUID) (*uuid.UUID, error) {
	if collectionIDStr == "" {
		return nil, nil
	}

	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid collection ID format")
	}

	var count int64
	if err := tx.Model(&collection.Collection{}).
		Where("id = ? AND user_id = ?", collectionID, userID).
		Count(&count).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to verify collection")
	}

	if count == 0 {
		return nil, fiber.NewError(fiber.StatusForbidden, "Collection does not belong to you")
	}

	return &collectionID, nil
}

func parseUUIDPointer(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}
