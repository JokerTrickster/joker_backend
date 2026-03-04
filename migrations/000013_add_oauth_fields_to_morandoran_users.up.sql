ALTER TABLE morandoran_users
    ADD COLUMN provider VARCHAR(20) NOT NULL DEFAULT 'email' AFTER role,
    ADD COLUMN profile_image VARCHAR(512) NULL AFTER provider,
    MODIFY COLUMN password VARCHAR(255) NULL;
