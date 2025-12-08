-- Create folders table
CREATE TABLE IF NOT EXISTS folders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    folder_name VARCHAR(255) NOT NULL,
    parent_folder_id BIGINT UNSIGNED,
    created_at TIMESTAMP NULL DEFAULT NULL,
    updated_at TIMESTAMP NULL DEFAULT NULL,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_parent_folder_id (parent_folder_id),
    INDEX idx_deleted_at (deleted_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_folder_id) REFERENCES folders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add folder_id column to cloud_files table
ALTER TABLE cloud_files
ADD COLUMN folder_id BIGINT UNSIGNED,
ADD INDEX idx_folder_id (folder_id),
ADD FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL;
