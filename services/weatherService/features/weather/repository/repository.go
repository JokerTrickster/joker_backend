package repository

import "gorm.io/gorm"

type RegisterAlarmWeatherRepository struct {
	GormDB *gorm.DB
}
