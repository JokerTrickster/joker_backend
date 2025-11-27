-- Create favorites table for bookmarking files
CREATE TABLE favorites (
  id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  file_id BIGINT UNSIGNED NOT NULL,
  favorited_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  -- Prevent duplicate favorites
  UNIQUE KEY uniq_user_file (user_id, file_id),

  -- Optimize queries for user's favorites list
  INDEX idx_user_favorited_at (user_id, favorited_at),
  INDEX idx_user_fileid (user_id, file_id),

  -- Auto-cleanup when user or file is deleted
  CONSTRAINT fk_fav_user FOREIGN KEY (user_id)
    REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_fav_file FOREIGN KEY (file_id)
    REFERENCES cloud_files(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
