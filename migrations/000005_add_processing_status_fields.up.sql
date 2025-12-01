-- Add processing status tracking fields to cloud_files table
ALTER TABLE cloud_files
ADD COLUMN processing_progress INT DEFAULT 0 COMMENT 'Processing progress percentage (0-100)',
ADD COLUMN processing_stage VARCHAR(50) NULL COMMENT 'Current processing stage',
ADD COLUMN processing_started_at TIMESTAMP NULL COMMENT 'When processing started',
ADD COLUMN processing_completed_at TIMESTAMP NULL COMMENT 'When processing completed or failed';

-- Add index for processing status queries
CREATE INDEX idx_cloud_files_processing_status ON cloud_files(processing_status);

-- Add index for processing progress tracking
CREATE INDEX idx_cloud_files_processing_stage ON cloud_files(processing_stage) WHERE processing_stage IS NOT NULL;
