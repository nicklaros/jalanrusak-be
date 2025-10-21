package errors

import (
	"errors"
	"fmt"
)

// Validation errors
var (
	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrRequired is returned when a required field is missing
	ErrRequired = errors.New("required field is missing")

	// ErrInvalidFormat is returned when a field has invalid format
	ErrInvalidFormat = errors.New("invalid format")

	// ErrInvalidLength is returned when a field exceeds length constraints
	ErrInvalidLength = errors.New("invalid length")
)

// Damaged road report errors
var (
	// ErrReportNotFound is returned when a damaged road report is not found
	ErrReportNotFound = errors.New("damaged road report not found")

	// ErrInvalidTitle is returned when report title is invalid
	ErrInvalidTitle = errors.New("title must be between 3 and 100 characters")

	// ErrInvalidSubDistrictCode is returned when administrative code is invalid
	ErrInvalidSubDistrictCode = errors.New("invalid subdistrict code format (expected: NN.NN.NN.NNNN)")

	// ErrInvalidCoordinates is returned when coordinates are invalid
	ErrInvalidCoordinates = errors.New("invalid coordinates")

	// ErrCoordinatesOutOfBounds is returned when coordinates are outside Indonesian boundaries
	ErrCoordinatesOutOfBounds = errors.New("coordinates outside Indonesian boundaries (lat: -11 to 6, lng: 95 to 141)")

	// ErrInvalidPath is returned when path points are invalid
	ErrInvalidPath = errors.New("path must have at least 1 coordinate point")

	// ErrTooManyPathPoints is returned when path has too many points
	ErrTooManyPathPoints = errors.New("path cannot have more than 100 coordinate points")

	// ErrInvalidPhotoURLs is returned when photo URLs are invalid
	ErrInvalidPhotoURLs = errors.New("at least 1 and at most 10 photo URLs required")

	// ErrPhotoURLNotAccessible is returned when photo URL is not accessible
	ErrPhotoURLNotAccessible = errors.New("photo URL is not accessible")

	// ErrInvalidPhotoURL is returned when photo URL format is invalid
	ErrInvalidPhotoURL = errors.New("invalid photo URL format")

	// ErrInvalidDescription is returned when description exceeds max length
	ErrInvalidDescription = errors.New("description cannot exceed 500 characters")

	// ErrInvalidStatus is returned when status is invalid
	ErrInvalidStatus = errors.New("invalid status")

	// ErrInvalidStatusTransition is returned when status transition is not allowed
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	// ErrUnauthorizedAccess is returned when user tries to access unauthorized resource
	ErrUnauthorizedAccess = errors.New("unauthorized access to resource")
)

// Geospatial errors
var (
	// ErrInvalidGeometry is returned when geometry is invalid
	ErrInvalidGeometry = errors.New("invalid geometry")

	// ErrLocationNotInBoundary is returned when location is not within expected boundary
	ErrLocationNotInBoundary = errors.New("location is not within expected administrative boundary")

	// ErrSubDistrictNotFound is returned when subdistrict code does not exist
	ErrSubDistrictNotFound = errors.New("subdistrict code not found")

	// ErrLocationMismatch is returned when coordinate and subdistrict don't match
	ErrLocationMismatch = errors.New("coordinates do not match the specified subdistrict area")
)

// Repository errors
var (
	// ErrDatabaseConnection is returned when database connection fails
	ErrDatabaseConnection = errors.New("database connection error")

	// ErrDatabaseQuery is returned when database query fails
	ErrDatabaseQuery = errors.New("database query error")

	// ErrDatabaseTransaction is returned when database transaction fails
	ErrDatabaseTransaction = errors.New("database transaction error")

	// ErrRecordNotFound is returned when database record is not found
	ErrRecordNotFound = errors.New("record not found")

	// ErrDuplicateRecord is returned when trying to create duplicate record
	ErrDuplicateRecord = errors.New("duplicate record")
)

// ValidationError wraps a validation error with field information
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// DatabaseError wraps a database error with context
type DatabaseError struct {
	Operation string
	Err       error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s: %v", e.Operation, e.Err)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// NewDatabaseError creates a new database error
func NewDatabaseError(operation string, err error) *DatabaseError {
	return &DatabaseError{
		Operation: operation,
		Err:       err,
	}
}
