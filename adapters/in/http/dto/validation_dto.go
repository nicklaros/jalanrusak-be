package dto

// ValidateLocationRequest represents the request to validate coordinates before report submission
type ValidateLocationRequest struct {
	SubDistrictCode string     `json:"subdistrict_code" binding:"required" example:"35.10.02.2005"`
	PathPoints      []PointDTO `json:"path_points" binding:"required,min=1,max=50,dive"`
}

// ValidateLocationResponse represents the validation result
type ValidateLocationResponse struct {
	Valid               bool    `json:"valid" example:"true"`
	Message             string  `json:"message" example:"Coordinates are valid"`
	SubDistrictExists   bool    `json:"subdistrict_exists" example:"true"`
	WithinBoundaries    bool    `json:"within_boundaries" example:"true"`
	NearCentroid        bool    `json:"near_centroid" example:"true"`
	MinDistanceToCenter float64 `json:"min_distance_to_center_meters,omitempty" example:"45.3"`
	CentroidLat         float64 `json:"centroid_lat,omitempty" example:"-7.257472"`
	CentroidLng         float64 `json:"centroid_lng,omitempty" example:"112.752090"`
}

// ValidatePhotosRequest represents the request to validate photo URLs
type ValidatePhotosRequest struct {
	PhotoURLs []string `json:"photo_urls" binding:"required,min=1,max=10,dive,url" example:"https://example.com/photo1.jpg"`
}

// ValidatePhotosResponse represents the photo validation results
type ValidatePhotosResponse struct {
	AllValid bool                    `json:"all_valid" example:"true"`
	Results  []PhotoValidationResult `json:"results"`
}

// PhotoValidationResult represents individual photo URL validation result
type PhotoValidationResult struct {
	URL         string `json:"url" example:"https://example.com/photo1.jpg"`
	Valid       bool   `json:"valid" example:"true"`
	Error       string `json:"error,omitempty" example:""`
	ContentType string `json:"content_type,omitempty" example:"image/jpeg"`
	SizeBytes   int64  `json:"size_bytes,omitempty" example:"524288"`
}
