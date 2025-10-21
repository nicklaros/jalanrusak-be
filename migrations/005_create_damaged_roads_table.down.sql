-- Drop trigger
DROP TRIGGER IF EXISTS update_damaged_roads_updated_at ON damaged_roads;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS damaged_road_photos CASCADE;
DROP TABLE IF EXISTS damaged_roads CASCADE;

-- Note: We don't drop the PostGIS extension as it may be used by other features
