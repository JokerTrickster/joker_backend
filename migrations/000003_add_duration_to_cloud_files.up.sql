-- Add duration field to cloud_files table for video length tracking
ALTER TABLE cloud_files
ADD COLUMN duration DECIMAL(10,2) NULL COMMENT 'Video duration in seconds';

-- Add index for duration-based queries if needed
CREATE INDEX idx_cloud_files_duration ON cloud_files(duration) WHERE duration IS NOT NULL;
