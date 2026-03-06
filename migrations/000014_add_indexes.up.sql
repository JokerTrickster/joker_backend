-- idx_news_category already exists in 000011; gallery indexes use alternate names for GORM alignment
CREATE INDEX idx_gallery_comments_gallery_id ON gallery_comments (gallery_id);
CREATE INDEX idx_gallery_comments_author_id ON gallery_comments (author_id);
