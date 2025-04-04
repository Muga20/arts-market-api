package repository

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	models "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	"gorm.io/gorm"
)

type ArtworkRepository interface {
	Create(artwork *models.Artwork) error
	GetByID(id uuid.UUID) (*models.Artwork, error)
	GetBySlug(slug string) (*models.Artwork, error)
	GetByUserID(userID uuid.UUID) ([]models.Artwork, error)
	Update(artwork *models.Artwork) error
	Delete(id uuid.UUID) error
}

type artworkRepository struct {
	db *gorm.DB
}

func NewArtworkRepository(db *gorm.DB) ArtworkRepository {
	return &artworkRepository{db: db}
}

func (r *artworkRepository) Create(artwork *models.Artwork) error {
	// Generate slug if empty
	if artwork.Slug == "" {
		artwork.Slug = slug.Make(artwork.Title)
	}

	// Ensure slug is unique
	var count int64
	r.db.Model(&models.Artwork{}).
		Where("slug = ?", artwork.Slug).
		Count(&count)

	if count > 0 {
		artwork.Slug = slug.Make(artwork.Title + "-" + strings.Split(artwork.ID.String(), "-")[0])
	}

	return r.db.Create(artwork).Error
}

func (r *artworkRepository) GetByID(id uuid.UUID) (*models.Artwork, error) {
	var artwork models.Artwork
	err := r.db.
		Preload("User").
		Preload("Collection").
		Preload("Images").
		Preload("Medium").
		Preload("Technique").
		Preload("Editions").
		First(&artwork, "id = ?", id).Error

	if err != nil {
		return nil, err
	}
	return &artwork, nil
}

func (r *artworkRepository) GetBySlug(slug string) (*models.Artwork, error) {
	var artwork models.Artwork
	err := r.db.
		Preload("User").
		Preload("Collection").
		Preload("Images").
		Preload("Medium").
		Preload("Technique").
		Preload("Editions").
		First(&artwork, "slug = ?", slug).Error

	if err != nil {
		return nil, err
	}
	return &artwork, nil
}

func (r *artworkRepository) GetByUserID(userID uuid.UUID) ([]models.Artwork, error) {
	var artworks []models.Artwork
	err := r.db.
		Preload("User").
		Preload("Collection").
		Preload("Images").
		Preload("Medium").
		Preload("Technique").
		Preload("Editions").
		Where("user_id = ?", userID).
		Find(&artworks).Error

	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func (r *artworkRepository) Update(artwork *models.Artwork) error {
	// If title changed, update slug
	if r.db.Statement.Changed("Title") {
		newSlug := slug.Make(artwork.Title)
		if newSlug != artwork.Slug {
			artwork.Slug = newSlug

			// Check for uniqueness
			var count int64
			r.db.Model(&models.Artwork{}).
				Where("slug = ? AND id != ?", artwork.Slug, artwork.ID).
				Count(&count)

			if count > 0 {
				artwork.Slug = slug.Make(artwork.Title + "-" + strings.Split(artwork.ID.String(), "-")[0])
			}
		}
	}

	return r.db.Save(artwork).Error
}

func (r *artworkRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&models.Artwork{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no artwork found with that ID")
	}
	return nil
}
