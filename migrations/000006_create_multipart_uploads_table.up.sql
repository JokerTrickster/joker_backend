-- Create multipart_uploads table for tracking large file uploads
CREATE TABLE multipart_uploads (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    upload_id VARCHAR(255) NOT NULL COMMENT 'S3 multipart upload ID',
    file_key VARCHAR(500) NOT NULL COMMENT 'S3 object key',
    file_name VARCHAR(255) NOT NULL COMMENT 'Original file name',
    file_size BIGINT NOT NULL COMMENT 'Total file size in bytes',
    content_type VARCHAR(100) NULL COMMENT 'MIME type',
    part_size INT NOT NULL COMMENT 'Size of each part in bytes',
    total_parts INT NOT NULL COMMENT 'Total number of parts',
    status ENUM('initiated', 'uploading', 'completed', 'aborted') DEFAULT 'initiated' COMMENT 'Upload status',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_upload_id (upload_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
