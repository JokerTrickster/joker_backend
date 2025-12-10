-- Remove index for permission
DROP INDEX IF EXISTS idx_file_shares_permission;
DROP INDEX IF EXISTS idx_folder_shares_permission;

-- Remove permission column from file_shares table
ALTER TABLE file_shares
DROP COLUMN permission;

-- Remove permission column from folder_shares table
ALTER TABLE folder_shares
DROP COLUMN permission;
