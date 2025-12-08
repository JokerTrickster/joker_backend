-- Remove folder_id column from cloud_files table
ALTER TABLE cloud_files
DROP FOREIGN KEY cloud_files_ibfk_3,
DROP INDEX idx_folder_id,
DROP COLUMN folder_id;

-- Drop folders table
DROP TABLE IF EXISTS folders;
