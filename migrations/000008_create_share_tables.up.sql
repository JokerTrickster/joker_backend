-- Create folder_shares table
CREATE TABLE IF NOT EXISTS folder_shares (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    folder_id BIGINT UNSIGNED NOT NULL,
    owner_id INT NOT NULL,
    shared_with_id INT NOT NULL,
    created_at TIMESTAMP NULL DEFAULT NULL,
    updated_at TIMESTAMP NULL DEFAULT NULL,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_folder_id (folder_id),
    INDEX idx_shared_with_id (shared_with_id),
    INDEX idx_owner_id (owner_id),
    INDEX idx_deleted_at (deleted_at),
    UNIQUE KEY uniq_folder_user (folder_id, shared_with_id, deleted_at),
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (shared_with_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create file_shares table
CREATE TABLE IF NOT EXISTS file_shares (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    file_id BIGINT UNSIGNED NOT NULL,
    owner_id INT NOT NULL,
    shared_with_id INT NOT NULL,
    created_at TIMESTAMP NULL DEFAULT NULL,
    updated_at TIMESTAMP NULL DEFAULT NULL,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_file_id (file_id),
    INDEX idx_shared_with_id (shared_with_id),
    INDEX idx_owner_id (owner_id),
    INDEX idx_deleted_at (deleted_at),
    UNIQUE KEY uniq_file_user (file_id, shared_with_id, deleted_at),
    FOREIGN KEY (file_id) REFERENCES cloud_files(id) ON DELETE CASCADE,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (shared_with_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
