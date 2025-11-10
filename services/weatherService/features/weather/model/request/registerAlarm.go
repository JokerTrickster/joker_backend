package request

type ReqRegisterAlarm struct {
	AlarmTime string `json:"alarmTime" validate:"required"` // 예: "09:00", "06:30"
	Region    string `json:"region" validate:"required"`    // 예: "서울시 관악구"
	FCMToken  string `json:"fcmToken" validate:"required"`  // FCM 토큰
	DeviceID  string `json:"deviceId"`                      // 디바이스 ID (선택)
}
