CREATE TABLE IF NOT EXISTS gallery_posts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    author_id BIGINT UNSIGNED NOT NULL,
    media_type ENUM('image','video') NOT NULL,
    media_url VARCHAR(512) NOT NULL,
    thumbnail_url VARCHAR(512) NOT NULL,
    caption VARCHAR(500),
    like_count INT UNSIGNED NOT NULL DEFAULT 0,
    comment_count INT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) NULL,
    INDEX idx_gallery_posts_author (author_id),
    INDEX idx_gallery_posts_deleted_at (deleted_at),
    CONSTRAINT fk_gallery_posts_author FOREIGN KEY (author_id) REFERENCES morandoran_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS gallery_likes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    gallery_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE INDEX idx_gallery_likes_user_gallery (user_id, gallery_id),
    CONSTRAINT fk_gallery_likes_user FOREIGN KEY (user_id) REFERENCES morandoran_users(id) ON DELETE CASCADE,
    CONSTRAINT fk_gallery_likes_gallery FOREIGN KEY (gallery_id) REFERENCES gallery_posts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS gallery_comments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    gallery_id BIGINT UNSIGNED NOT NULL,
    author_id BIGINT UNSIGNED NOT NULL,
    content VARCHAR(300) NOT NULL,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) NULL,
    INDEX idx_gallery_comments_gallery (gallery_id),
    INDEX idx_gallery_comments_author (author_id),
    INDEX idx_gallery_comments_deleted_at (deleted_at),
    CONSTRAINT fk_gallery_comments_gallery FOREIGN KEY (gallery_id) REFERENCES gallery_posts(id) ON DELETE CASCADE,
    CONSTRAINT fk_gallery_comments_author FOREIGN KEY (author_id) REFERENCES morandoran_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
