-- Remove processing status indexes
DROP INDEX IF EXISTS idx_cloud_files_processing_stage ON cloud_files;
DROP INDEX IF EXISTS idx_cloud_files_processing_status ON cloud_files;

-- Remove processing status tracking fields from cloud_files table
ALTER TABLE cloud_files
DROP COLUMN IF EXISTS processing_completed_at,
DROP COLUMN IF EXISTS processing_started_at,
DROP COLUMN IF EXISTS processing_stage,
DROP COLUMN IF EXISTS processing_progress;
