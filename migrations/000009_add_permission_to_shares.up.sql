-- Add permission column to folder_shares table
ALTER TABLE folder_shares
ADD COLUMN permission VARCHAR(10) NOT NULL DEFAULT 'read';

-- Add permission column to file_shares table
ALTER TABLE file_shares
ADD COLUMN permission VARCHAR(10) NOT NULL DEFAULT 'read';

-- Add index for permission-based queries
CREATE INDEX idx_folder_shares_permission ON folder_shares(permission);
CREATE INDEX idx_file_shares_permission ON file_shares(permission);
