CREATE TABLE tokens (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    refresh_expired_at INT,
    user_id INT,
    access_token VARCHAR(255),
    refresh_token VARCHAR(255)
);

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    name VARCHAR(255),
    email VARCHAR(255),
    password VARCHAR(255),
    provider VARCHAR(50),
);
<-- 
   weather service table
-->

create table weather_service_tokens (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    user_id INT NOT NULL,
    fcm_token VARCHAR(500) NOT NULL,
    device_id VARCHAR(255),
    unique key uk_user_device (user_id, device_id),
    foreign key (user_id) references users(id)
);

create table user_alarms (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    user_id INT,
    alarm_time TIME NOT NULL,
    region varchar(255) not null,
    is_enabled boolean default true,
    last_sent datetime default null,
    foreign key (user_id) references users(id),
    index idx_alarm_time (alarm_time, is_enabled, deleted_at),
    index idx_last_sent (last_sent)
);

<--
    cloud repository service tables
-->

-- Update users table (add columns if not exist)
-- ALTER TABLE users ADD COLUMN storage_used BIGINT DEFAULT 0;
-- ALTER TABLE users ADD COLUMN storage_limit BIGINT DEFAULT 16106127360;

CREATE TABLE files (
    id VARCHAR(36) PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(255),
    original_name VARCHAR(255),
    s3_key VARCHAR(512),
    url VARCHAR(1024),
    mime_type VARCHAR(100),
    size BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT FALSE,
    metadata JSON,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE tags (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE file_tags (
    file_id VARCHAR(36),
    tag_id VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (file_id, tag_id),
    FOREIGN KEY (file_id) REFERENCES files(id),
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);

CREATE TABLE activity_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id INT NOT NULL,
    action_type VARCHAR(50),
    target_id VARCHAR(36),
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

<!-- 
   morandoran service tables
-->

CREATE TABLE morandoran_users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    nickname VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NULL,
    role ENUM('user','admin') NOT NULL DEFAULT 'user',
    provider VARCHAR(20) NOT NULL DEFAULT 'email',
    profile_image VARCHAR(512) NULL,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) NULL,
    UNIQUE INDEX idx_morandoran_users_email (email),
    INDEX idx_morandoran_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE news (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    summary VARCHAR(500),
    content TEXT NOT NULL,
    thumbnail VARCHAR(512),
    category VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) NULL,
    INDEX idx_news_category (category),
    INDEX idx_news_date (date),
    INDEX idx_news_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE products (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price INT UNSIGNED NOT NULL,
    original_price INT UNSIGNED NULL,
    description TEXT NOT NULL,
    image VARCHAR(512) NOT NULL,
    category VARCHAR(50) NOT NULL,
    badge VARCHAR(20) NULL,
    in_stock BOOLEAN NOT NULL DEFAULT true,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) NULL,
    INDEX idx_products_category (category),
    INDEX idx_products_in_stock (in_stock),
    INDEX idx_products_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE rankings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    game_type VARCHAR(30) NOT NULL,
    nickname VARCHAR(50) NOT NULL,
    clear_time_ms INT UNSIGNED NOT NULL,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE INDEX idx_rankings_user_game (user_id, game_type),
    INDEX idx_rankings_game_time (game_type, clear_time_ms),
    CONSTRAINT fk_rankings_user FOREIGN KEY (user_id) REFERENCES morandoran_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE gallery_posts (
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

CREATE TABLE gallery_likes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    gallery_id BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE INDEX idx_gallery_likes_user_gallery (user_id, gallery_id),
    CONSTRAINT fk_gallery_likes_user FOREIGN KEY (user_id) REFERENCES morandoran_users(id) ON DELETE CASCADE,
    CONSTRAINT fk_gallery_likes_gallery FOREIGN KEY (gallery_id) REFERENCES gallery_posts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE gallery_comments (
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