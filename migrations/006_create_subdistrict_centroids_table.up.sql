-- Migration: Create subdistrict_centroids table for geographic validation
-- Purpose: Store centroid coordinates for Indonesian subdistricts to validate report locations
-- Data Source: BIG (Badan Informasi Geospasial) official geospatial data
-- Validation: FR-006 requires coordinates within 200m of subdistrict centroid

CREATE TABLE IF NOT EXISTS subdistrict_centroids (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subdistrict_code VARCHAR(20) NOT NULL UNIQUE,
    centroid_lat DOUBLE PRECISION NOT NULL,
    centroid_lng DOUBLE PRECISION NOT NULL,
    name VARCHAR(255) NOT NULL,
    province_code VARCHAR(5) NOT NULL,
    district_code VARCHAR(8) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Validation constraints
    CONSTRAINT chk_subdistrict_code_format CHECK (subdistrict_code ~ '^\d{2}\.\d{2}\.\d{2}\.\d{4}$'),
    CONSTRAINT chk_centroid_lat_bounds CHECK (centroid_lat >= -11 AND centroid_lat <= 6),
    CONSTRAINT chk_centroid_lng_bounds CHECK (centroid_lng >= 95 AND centroid_lng <= 141)
);

-- Index for fast subdistrict code lookup during validation
CREATE INDEX idx_subdistrict_centroids_code ON subdistrict_centroids(subdistrict_code);

-- Index for hierarchical queries (province, district level aggregations)
CREATE INDEX idx_subdistrict_centroids_province ON subdistrict_centroids(province_code);
CREATE INDEX idx_subdistrict_centroids_district ON subdistrict_centroids(district_code);

-- Add comment for documentation
COMMENT ON TABLE subdistrict_centroids IS 'Geographic centroids for Indonesian subdistricts used for location validation per FR-006. Data sourced from BIG (Badan Informasi Geospasial).';
COMMENT ON COLUMN subdistrict_centroids.subdistrict_code IS 'Kemendagri format: NN.NN.NN.NNNN (province.district.subdistrict.village)';
COMMENT ON COLUMN subdistrict_centroids.centroid_lat IS 'Latitude of subdistrict centroid in decimal degrees';
COMMENT ON COLUMN subdistrict_centroids.centroid_lng IS 'Longitude of subdistrict centroid in decimal degrees';

-- Seed example data for testing (East Java subdistricts)
-- In production, this should be populated from official BIG geospatial dataset
INSERT INTO subdistrict_centroids (subdistrict_code, centroid_lat, centroid_lng, name, province_code, district_code) VALUES
    ('35.10.02.2005', -7.257472, 112.752090, 'Kelurahan Ketintang, Gayungan, Surabaya', '35', '35.10'),
    ('35.78.01.1001', -7.983908, 112.630892, 'Desa Sukorejo, Sukorejo, Ponorogo', '35', '35.78'),
    ('35.09.01.2001', -7.943893, 112.612766, 'Kelurahan Banjarsari, Buduran, Sidoarjo', '35', '35.09')
ON CONFLICT (subdistrict_code) DO NOTHING;
