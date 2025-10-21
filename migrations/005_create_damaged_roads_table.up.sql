-- Create PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create damaged_roads table
CREATE TABLE IF NOT EXISTS damaged_roads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(100) NOT NULL,
    subdistrict_code VARCHAR(13) NOT NULL,
    path GEOMETRY(LINESTRING, 4326) NOT NULL,
    description TEXT,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'submitted',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Basic data integrity constraints
    CONSTRAINT valid_title_length CHECK (LENGTH(title) >= 3 AND LENGTH(title) <= 100),
    CONSTRAINT valid_subdistrict_code_format CHECK (subdistrict_code ~ '^\d{2}\.\d{2}\.\d{2}\.\d{4}$'),
    CONSTRAINT valid_path_geometry CHECK (ST_IsValid(path) AND ST_GeometryType(path) = 'ST_LineString'),
    CONSTRAINT valid_status CHECK (status IN ('submitted', 'under_verification', 'verified', 'pending_resolved', 'resolved', 'archived')),
    CONSTRAINT valid_description_length CHECK (description IS NULL OR LENGTH(description) <= 500)
);

-- Create indexes for performance
CREATE INDEX idx_damaged_roads_author ON damaged_roads(author_id);
CREATE INDEX idx_damaged_roads_status ON damaged_roads(status);
CREATE INDEX idx_damaged_roads_subdistrict ON damaged_roads(subdistrict_code);
CREATE INDEX idx_damaged_roads_created_at ON damaged_roads(created_at DESC);
CREATE INDEX idx_damaged_roads_path_geom ON damaged_roads USING GIST (path);

-- Create damaged_road_photos table
CREATE TABLE IF NOT EXISTS damaged_road_photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    road_id UUID NOT NULL REFERENCES damaged_roads(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    content_type VARCHAR(50),
    file_size BIGINT,
    validation_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    validated_at TIMESTAMP WITH TIME ZONE,
    validation_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_photo_url UNIQUE(road_id, url),
    CONSTRAINT valid_validation_status CHECK (validation_status IN ('pending', 'valid', 'invalid', 'error'))
);

-- Create indexes for photos
CREATE INDEX idx_damaged_road_photos_road ON damaged_road_photos(road_id);
CREATE INDEX idx_damaged_road_photos_validation ON damaged_road_photos(validation_status);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_damaged_roads_updated_at BEFORE UPDATE ON damaged_roads
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
