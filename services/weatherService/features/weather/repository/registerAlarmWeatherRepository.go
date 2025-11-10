package repository

import (
	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"gorm.io/gorm"
)

func NewRegisterAlarmWeatherRepository(gormDB *gorm.DB) _interface.IRegisterAlarmWeatherRepository {
	return &RegisterAlarmWeatherRepository{GormDB: gormDB}
}
