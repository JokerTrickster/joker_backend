ALTER TABLE morandoran_users
    DROP COLUMN provider,
    DROP COLUMN profile_image,
    MODIFY COLUMN password VARCHAR(255) NOT NULL;
