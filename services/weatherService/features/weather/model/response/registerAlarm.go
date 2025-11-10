package response

type ResRegisterAlarm struct {
	AlarmID   int    `json:"alarmId"`
	AlarmTime string `json:"alarmTime"`
	Region    string `json:"region"`
	IsEnabled bool   `json:"isEnabled"`
}
