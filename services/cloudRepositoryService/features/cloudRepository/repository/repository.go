package repository

import (
	"gorm.io/gorm"
)

type CloudRepositoryRepository struct {
	db     *gorm.DB
	bucket string
}

func NewCloudRepositoryRepository(db *gorm.DB, bucket string) *CloudRepositoryRepository {
	return &CloudRepositoryRepository{
		db:     db,
		bucket: bucket,
	}
}
