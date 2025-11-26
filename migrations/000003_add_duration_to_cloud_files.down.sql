-- Remove duration index
DROP INDEX IF EXISTS idx_cloud_files_duration ON cloud_files;

-- Remove duration column from cloud_files table
ALTER TABLE cloud_files
DROP COLUMN IF EXISTS duration;
