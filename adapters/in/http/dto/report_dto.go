package dto

import "github.com/nicklaros/jalanrusak-be/core/domain/entities"

// PointDTO represents a coordinate point in the request
type PointDTO struct {
	Lat float64 `json:"lat" binding:"required,gte=-11,lte=6" example:"-7.2575"`
	Lng float64 `json:"lng" binding:"required,gte=95,lte=141" example:"112.7521"`
}

// CreateDamagedRoadRequest represents the request to create a damaged road report
type CreateDamagedRoadRequest struct {
	Title           string     `json:"title" binding:"required,min=3,max=100" example:"Jalan berlubang di depan SDN 01"`
	SubDistrictCode string     `json:"subdistrict_code" binding:"required" example:"35.10.02.2005"`
	PathPoints      []PointDTO `json:"path_points" binding:"required,min=1,max=100"`
	PhotoURLs       []string   `json:"photo_urls" binding:"required,min=1,max=10"`
	Description     *string    `json:"description,omitempty" binding:"omitempty,max=500" example:"Jalan berlubang sepanjang 50 meter"`
}

// GeometryDTO represents a PostGIS geometry in the response
type GeometryDTO struct {
	Type        string      `json:"type" example:"LineString"`
	Coordinates [][]float64 `json:"coordinates"`
}

// DamagedRoadResponse represents a damaged road report in the response
type DamagedRoadResponse struct {
	ID              string      `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title           string      `json:"title" example:"Jalan berlubang di depan SDN 01"`
	SubDistrictCode string      `json:"subdistrict_code" example:"35.10.02.2005"`
	Path            GeometryDTO `json:"path"`
	Description     *string     `json:"description,omitempty" example:"Jalan berlubang sepanjang 50 meter"`
	PhotoURLs       []string    `json:"photo_urls"`
	AuthorID        string      `json:"author_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status          string      `json:"status" example:"submitted"`
	CreatedAt       string      `json:"created_at" example:"2025-10-20T10:00:00Z"`
	UpdatedAt       string      `json:"updated_at" example:"2025-10-20T10:00:00Z"`
}

// DamagedRoadListResponse represents a paginated list of damaged road reports
type DamagedRoadListResponse struct {
	Data       []DamagedRoadResponse `json:"data"`
	Pagination PaginationMeta        `json:"pagination"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Total  int `json:"total" example:"100"`
	Limit  int `json:"limit" example:"20"`
	Offset int `json:"offset" example:"0"`
	Page   int `json:"page" example:"1"`
}

// UpdateStatusRequest represents the request to update report status
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required" example:"under_verification"`
}

// ToEntity converts CreateDamagedRoadRequest to domain entities
func (r *CreateDamagedRoadRequest) ToEntity() (
	entities.Title,
	entities.SubDistrictCode,
	[]entities.Point,
	*entities.Description,
	error,
) {
	title, err := entities.NewTitle(r.Title)
	if err != nil {
		return "", "", nil, nil, err
	}

	subdistrictCode, err := entities.NewSubDistrictCode(r.SubDistrictCode)
	if err != nil {
		return "", "", nil, nil, err
	}

	points := make([]entities.Point, len(r.PathPoints))
	for i, p := range r.PathPoints {
		point, err := entities.NewPoint(p.Lat, p.Lng)
		if err != nil {
			return "", "", nil, nil, err
		}
		points[i] = *point
	}

	var description *entities.Description
	if r.Description != nil && *r.Description != "" {
		desc, err := entities.NewDescription(*r.Description)
		if err != nil {
			return "", "", nil, nil, err
		}
		description = &desc
	}

	return title, subdistrictCode, points, description, nil
}

// FromDamagedRoad converts a DamagedRoad entity to a response DTO
func FromDamagedRoad(road *entities.DamagedRoad) DamagedRoadResponse {
	var description *string
	if road.Description != nil {
		desc := road.Description.String()
		description = &desc
	}

	return DamagedRoadResponse{
		ID:              road.ID.String(),
		Title:           road.Title.String(),
		SubDistrictCode: road.SubDistrictCode.String(),
		Path: GeometryDTO{
			Type:        road.Path.Type,
			Coordinates: road.Path.Coordinates,
		},
		Description: description,
		PhotoURLs:   road.PhotoURLs,
		AuthorID:    road.AuthorID.String(),
		Status:      road.Status.String(),
		CreatedAt:   road.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   road.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
