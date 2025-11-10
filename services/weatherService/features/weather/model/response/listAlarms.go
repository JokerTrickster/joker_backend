package response

type AlarmItem struct {
	AlarmID   int    `json:"alarmId"`
	AlarmTime string `json:"alarmTime"`
	Region    string `json:"region"`
	IsEnabled bool   `json:"isEnabled"`
	LastSent  string `json:"lastSent,omitempty"` // ISO 8601 format or empty
}

type ResListAlarms struct {
	Alarms []AlarmItem `json:"alarms"`
	Total  int         `json:"total"`
}
