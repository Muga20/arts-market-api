package models

import (
	"github.com/google/uuid"
	collection "github.com/muga20/artsMarket/modules/artwork-management/models/collection"
	medium "github.com/muga20/artsMarket/modules/artwork-management/models/medium"
	technique "github.com/muga20/artsMarket/modules/artwork-management/models/technique"
	user "github.com/muga20/artsMarket/modules/users/models"
	"gorm.io/gorm"
	"time"
)

type ArtworkStatus string
type ArtworkType string
type ConditionType string
type LicenseType string

const (
	// Artwork Statuses
	PendingStatus  ArtworkStatus = "pending"
	ApprovedStatus ArtworkStatus = "approved"
	RejectedStatus ArtworkStatus = "rejected"

	// Artwork Types
	TraditionalType ArtworkType = "traditional" // Physical artworks like paintings, sculptures, etc.
	DigitalType     ArtworkType = "digital"     // Digital artworks created on a computer or digital platform
	PhotographyType ArtworkType = "photography" // Photographic artworks, including digital and analog photography
	MixedMediaType  ArtworkType = "mixed_media" // Artworks that combine different materials and mediums
	SculptureType   ArtworkType = "sculpture"   // 3D artworks, including sculptures and installations
	PerformanceType ArtworkType = "performance" // Artworks that involve live actions or performances

	// Artwork Conditions
	PristineCondition   ConditionType = "pristine"   // Artwork is in perfect, untouched condition
	ExcellentCondition  ConditionType = "excellent"  // Artwork shows minimal signs of wear
	GoodCondition       ConditionType = "good"       // Artwork is in good condition with minor wear
	AcceptableCondition ConditionType = "acceptable" // Artwork shows noticeable wear but is still presentable
	RestoredCondition   ConditionType = "restored"   // Artwork has been professionally restored
	DamagedCondition    ConditionType = "damaged"    // Artwork is damaged and may need repair or restoration

	// License Types
	AllRightsReserved   LicenseType = "all_rights_reserved"   // All rights reserved to the artist
	CreativeCommons     LicenseType = "creative_commons"      // Creative Commons license for sharing
	PublicDomain        LicenseType = "public_domain"         // Artwork is in the public domain, free to use
	ExclusiveLicense    LicenseType = "exclusive_license"     // Exclusive rights granted to a specific entity
	NonExclusiveLicense LicenseType = "non_exclusive_license" // Non-exclusive rights for multiple entities
	CustomLicense       LicenseType = "custom"                // Custom licensing terms agreed upon
)

type Artwork struct {
	ID             uuid.UUID     `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	UserID         uuid.UUID     `gorm:"type:char(36);not null" json:"user_id"`
	CollectionID   *uuid.UUID    `gorm:"type:char(36);not null" json:"collection_id"`
	Title          string        `gorm:"type:varchar(255);not null" json:"title"`
	Description    string        `gorm:"type:text" json:"description"`
	CreationDate   *time.Time    `gorm:"type:date" json:"creation_date"`
	Price          *float64      `gorm:"type:decimal(10,2)" json:"price"`
	IsForSale      bool          `gorm:"type:boolean;not null;default:false" json:"is_for_sale"`
	Status         ArtworkStatus `gorm:"type:enum('pending','approved','rejected');default:'pending'" json:"status"`
	Type           ArtworkType   `gorm:"type:enum('traditional','digital','photography','mixed_media','sculpture','performance');not null" json:"type"`
	Dimensions     string        `gorm:"type:varchar(255)" json:"dimensions"`
	Weight         *float64      `gorm:"type:decimal(6,2)" json:"weight"`
	IsFramed       bool          `gorm:"type:boolean;not null;default:false" json:"is_framed"`
	Condition      ConditionType `gorm:"type:enum('pristine','excellent','good','acceptable','restored','damaged');default:'pristine'" json:"condition"`
	MediumID       *uuid.UUID    `gorm:"type:char(36)" json:"medium_id"`
	TechniqueID    *uuid.UUID    `gorm:"type:char(36)" json:"technique_id"`
	LicenseType    LicenseType   `gorm:"type:enum('all_rights_reserved','creative_commons','public_domain','exclusive_license','non_exclusive_license','custom');default:'all_rights_reserved'" json:"license_type"`
	LicenseDetails string        `gorm:"type:text" json:"license_details"`
	ViewCount      int           `gorm:"type:int;not null;default:0" json:"view_count"`

	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	User       user.User             `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Collection collection.Collection `gorm:"foreignKey:CollectionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Images     []ArtworkImage        `gorm:"foreignKey:ArtworkID;constraint:OnDelete:CASCADE"`
	Medium     medium.Medium         `gorm:"foreignKey:MediumID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Technique  technique.Technique   `gorm:"foreignKey:TechniqueID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Editions   []Edition             `gorm:"foreignKey:ArtworkID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID if not set
func (a *Artwork) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}

	if a.Status == "" {
		a.Status = PendingStatus
	}

	if a.Type == "" {
		a.Type = TraditionalType
	}

	if a.Condition == "" {
		a.Condition = PristineCondition
	}

	if a.LicenseType == "" {
		a.LicenseType = AllRightsReserved
	}

	return
}
