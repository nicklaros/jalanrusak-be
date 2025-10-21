package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/nicklaros/jalanrusak-be/core/domain/errors"
)

// Status represents the lifecycle state of a damaged road report
type Status string

const (
	// StatusSubmitted indicates the report has been submitted by a user
	StatusSubmitted Status = "submitted"
	// StatusUnderVerification indicates the report is being verified
	StatusUnderVerification Status = "under_verification"
	// StatusVerified indicates the report has been verified as valid
	StatusVerified Status = "verified"
	// StatusPendingResolved indicates repair is scheduled but not yet complete
	StatusPendingResolved Status = "pending_resolved"
	// StatusResolved indicates the road damage has been repaired
	StatusResolved Status = "resolved"
	// StatusArchived indicates the report has been archived
	StatusArchived Status = "archived"
)

// AllStatuses returns all valid status values
func AllStatuses() []Status {
	return []Status{
		StatusSubmitted,
		StatusUnderVerification,
		StatusVerified,
		StatusPendingResolved,
		StatusResolved,
		StatusArchived,
	}
}

// IsValid checks if the status is valid
func (s Status) IsValid() bool {
	for _, validStatus := range AllStatuses() {
		if s == validStatus {
			return true
		}
	}
	return false
}

// CanTransitionTo checks if transition to another status is allowed
func (s Status) CanTransitionTo(newStatus Status) bool {
	// Define allowed transitions (strictly forward)
	allowedTransitions := map[Status][]Status{
		StatusSubmitted:         {StatusUnderVerification},
		StatusUnderVerification: {StatusVerified},
		StatusVerified:          {StatusPendingResolved},
		StatusPendingResolved:   {StatusResolved},
		StatusResolved:          {StatusArchived},
		StatusArchived:          {}, // Terminal state - no transitions allowed
	}

	allowedTargets, exists := allowedTransitions[s]
	if !exists {
		return false
	}

	for _, allowed := range allowedTargets {
		if newStatus == allowed {
			return true
		}
	}
	return false
}

// String returns the string representation of the status
func (s Status) String() string {
	return string(s)
}

// DamagedRoad represents a damaged road report entity
type DamagedRoad struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	Title           Title           `json:"title" db:"title"`
	SubDistrictCode SubDistrictCode `json:"subdistrict_code" db:"subdistrict_code"`
	Path            Geometry        `json:"path" db:"path"`
	Description     *Description    `json:"description,omitempty" db:"description"`
	PhotoURLs       []string        `json:"photo_urls" db:"photo_urls"`
	AuthorID        uuid.UUID       `json:"author_id" db:"author_id"`
	Status          Status          `json:"status" db:"status"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// NewDamagedRoad creates a new DamagedRoad with validation
func NewDamagedRoad(
	title Title,
	subdistrictCode SubDistrictCode,
	path Geometry,
	photoURLs []string,
	authorID uuid.UUID,
	description *Description,
) (*DamagedRoad, error) {
	now := time.Now()

	road := &DamagedRoad{
		ID:              uuid.New(),
		Title:           title,
		SubDistrictCode: subdistrictCode,
		Path:            path,
		Description:     description,
		PhotoURLs:       photoURLs,
		AuthorID:        authorID,
		Status:          StatusSubmitted,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := road.Validate(); err != nil {
		return nil, err
	}

	return road, nil
}

// Validate validates the DamagedRoad entity
func (d *DamagedRoad) Validate() error {
	// Validate title
	if err := d.Title.Validate(); err != nil {
		return err
	}

	// Validate subdistrict code
	if err := d.SubDistrictCode.Validate(); err != nil {
		return err
	}

	// Validate path
	if err := d.Path.Validate(); err != nil {
		return err
	}

	// Validate description if provided
	if d.Description != nil {
		if err := d.Description.Validate(); err != nil {
			return err
		}
	}

	// Validate photo URLs
	if len(d.PhotoURLs) < 1 {
		return errors.NewValidationError("photo_urls", "at least 1 photo URL required", errors.ErrInvalidPhotoURLs)
	}
	if len(d.PhotoURLs) > 10 {
		return errors.NewValidationError("photo_urls", "cannot have more than 10 photo URLs", errors.ErrInvalidPhotoURLs)
	}

	// Validate status
	if !d.Status.IsValid() {
		return errors.NewValidationError("status", "invalid status value", errors.ErrInvalidStatus)
	}

	// Validate author ID
	if d.AuthorID == uuid.Nil {
		return errors.NewValidationError("author_id", "author ID is required", errors.ErrRequired)
	}

	return nil
}

// UpdateStatus updates the status with transition validation
func (d *DamagedRoad) UpdateStatus(newStatus Status) error {
	if !newStatus.IsValid() {
		return errors.NewValidationError("status", "invalid status value", errors.ErrInvalidStatus)
	}

	if !d.Status.CanTransitionTo(newStatus) {
		return errors.NewValidationError(
			"status",
			"cannot transition from "+d.Status.String()+" to "+newStatus.String(),
			errors.ErrInvalidStatusTransition,
		)
	}

	d.Status = newStatus
	d.UpdatedAt = time.Now()
	return nil
}

// CanBeEditedBy checks if the damaged road can be edited by the given user
func (d *DamagedRoad) CanBeEditedBy(userID uuid.UUID) bool {
	// Only the author can edit their own report
	return d.AuthorID == userID
}

// DamagedRoadFilters represents filters for querying damaged road reports
type DamagedRoadFilters struct {
	Status          *Status    `json:"status,omitempty"`
	SubDistrictCode *string    `json:"subdistrict_code,omitempty"`
	AuthorID        *uuid.UUID `json:"author_id,omitempty"`
	Limit           int        `json:"limit"`
	Offset          int        `json:"offset"`
}

// NewDamagedRoadFilters creates filters with defaults
func NewDamagedRoadFilters() *DamagedRoadFilters {
	return &DamagedRoadFilters{
		Limit:  20,
		Offset: 0,
	}
}
