package request

type ReqDeleteAlarm struct {
	AlarmID int `json:"alarmId" validate:"required,gt=0"`
}
