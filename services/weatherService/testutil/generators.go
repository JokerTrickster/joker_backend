package testutil

import (
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
)

// GenerateTestAlarms creates N test alarms with specified base time
func GenerateTestAlarms(count int, baseTime time.Time) []entity.UserAlarm {
	alarms := make([]entity.UserAlarm, count)
	regions := []string{
		"서울시 강남구",
		"서울시 서초구",
		"경기도 성남시",
		"부산시 해운대구",
		"대구시 중구",
		"인천시 남동구",
		"광주시 북구",
		"대전시 서구",
		"울산시 남구",
		"경기도 수원시",
	}

	for i := 0; i < count; i++ {
		alarms[i] = entity.UserAlarm{
			UserID:    i + 1,
			AlarmTime: baseTime.Format("15:04:05"),
			Region:    regions[i%len(regions)],
			IsEnabled: true,
			LastSent:  nil,
		}
	}

	return alarms
}

// GenerateTestAlarmsWithLastSent creates alarms with last_sent timestamps
func GenerateTestAlarmsWithLastSent(count int, baseTime time.Time, lastSent time.Time) []entity.UserAlarm {
	alarms := GenerateTestAlarms(count, baseTime)
	for i := range alarms {
		alarms[i].LastSent = &lastSent
	}
	return alarms
}

// GenerateTestWeatherData creates test weather data for a region
func GenerateTestWeatherData(region string) *entity.WeatherData {
	return &entity.WeatherData{
		Temperature:   20.0 + float64(len(region)%10),
		Humidity:      50.0 + float64(len(region)%20),
		Precipitation: float64(len(region) % 5),
		WindSpeed:     2.0 + float64(len(region)%3),
		CachedAt:      time.Now(),
	}
}

// GenerateTestWeatherDataWithValues creates weather data with specific values
func GenerateTestWeatherDataWithValues(temp, humidity, precip, wind float64) *entity.WeatherData {
	return &entity.WeatherData{
		Temperature:   temp,
		Humidity:      humidity,
		Precipitation: precip,
		WindSpeed:     wind,
		CachedAt:      time.Now(),
	}
}

// GenerateTestFCMTokens creates N FCM tokens for a user
func GenerateTestFCMTokens(userID int, count int) []entity.WeatherServiceToken {
	tokens := make([]entity.WeatherServiceToken, count)
	for i := 0; i < count; i++ {
		tokens[i] = entity.WeatherServiceToken{
			UserID:   userID,
			FCMToken: fmt.Sprintf("fcm_token_%d_%d", userID, i),
			DeviceID: fmt.Sprintf("device_%d_%d", userID, i),
		}
	}
	return tokens
}

// GenerateTestFCMTokensWithPrefix creates tokens with custom prefix
func GenerateTestFCMTokensWithPrefix(userID int, count int, prefix string) []entity.WeatherServiceToken {
	tokens := make([]entity.WeatherServiceToken, count)
	for i := 0; i < count; i++ {
		tokens[i] = entity.WeatherServiceToken{
			UserID:   userID,
			FCMToken: fmt.Sprintf("%s_token_%d_%d", prefix, userID, i),
			DeviceID: fmt.Sprintf("%s_device_%d_%d", prefix, userID, i),
		}
	}
	return tokens
}

// GenerateVariedAlarms creates alarms with different times and regions
func GenerateVariedAlarms(count int) []entity.UserAlarm {
	alarms := make([]entity.UserAlarm, count)
	baseTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.Local)
	regions := []string{
		"서울시 강남구", "서울시 서초구", "경기도 성남시", "부산시 해운대구",
		"대구시 중구", "인천시 남동구", "광주시 북구", "대전시 서구",
	}

	for i := 0; i < count; i++ {
		// Vary alarm times
		alarmTime := baseTime.Add(time.Duration(i%12) * time.Hour)

		alarms[i] = entity.UserAlarm{
			UserID:    (i % 10) + 1, // Reuse users
			AlarmTime: alarmTime.Format("15:04:05"),
			Region:    regions[i%len(regions)],
			IsEnabled: i%5 != 0, // 80% enabled
			LastSent:  nil,
		}
	}

	return alarms
}

// GenerateDisabledAlarms creates disabled alarms for testing
func GenerateDisabledAlarms(count int, baseTime time.Time) []entity.UserAlarm {
	alarms := GenerateTestAlarms(count, baseTime)
	for i := range alarms {
		alarms[i].IsEnabled = false
	}
	return alarms
}

// GenerateAlarmsWithSentToday creates alarms already sent today
func GenerateAlarmsWithSentToday(count int, baseTime time.Time) []entity.UserAlarm {
	alarms := GenerateTestAlarms(count, baseTime)
	now := time.Now()
	for i := range alarms {
		alarms[i].LastSent = &now
	}
	return alarms
}

// GenerateAlarmsWithSentYesterday creates alarms sent yesterday
func GenerateAlarmsWithSentYesterday(count int, baseTime time.Time) []entity.UserAlarm {
	alarms := GenerateTestAlarms(count, baseTime)
	yesterday := time.Now().AddDate(0, 0, -1)
	for i := range alarms {
		alarms[i].LastSent = &yesterday
	}
	return alarms
}

// GenerateMixedAlarms creates a mix of enabled/disabled and sent/unsent alarms
func GenerateMixedAlarms(count int, baseTime time.Time) []entity.UserAlarm {
	alarms := make([]entity.UserAlarm, count)
	regions := []string{
		"서울시 강남구", "서울시 서초구", "경기도 성남시", "부산시 해운대구",
	}

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	for i := 0; i < count; i++ {
		alarm := entity.UserAlarm{
			UserID:    (i % 10) + 1,
			AlarmTime: baseTime.Format("15:04:05"),
			Region:    regions[i%len(regions)],
			IsEnabled: i%4 != 0, // 75% enabled
		}

		// Vary last_sent status
		switch i % 3 {
		case 0:
			alarm.LastSent = nil // Never sent
		case 1:
			alarm.LastSent = &yesterday // Sent yesterday
		case 2:
			alarm.LastSent = &now // Sent today
		}

		alarms[i] = alarm
	}

	return alarms
}

// GenerateInvalidRegionNames creates alarms with edge case region names
func GenerateInvalidRegionNames() []string {
	return []string{
		"",                                    // Empty
		"서울시",                                // Short
		string(make([]byte, 1000)),            // Very long
		"서울시 !@#$%^&*()",                     // Special chars
		"서울시 😀🎉",                            // Emoji
		"서울시'; DROP TABLE alarms; --",       // SQL injection
		"서울시\x00강남구",                        // Null byte
		"   서울시   ",                          // Whitespace
		"SEOUL GANGNAM-GU",                    // English
		"123 456 789",                         // Numbers only
		"서울" + string(make([]byte, 500)),      // Mixed valid + long
		"../../../etc/passwd",                 // Path traversal
		"<script>alert('xss')</script>",       // XSS attempt
		"서울\r\n강남\r\n",                        // Line breaks
		"서울\t강남\t",                           // Tabs
	}
}

// GenerateTimeEdgeCases creates test cases for time boundaries
func GenerateTimeEdgeCases() []time.Time {
	loc := time.Local

	return []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, loc),      // Midnight
		time.Date(2024, 1, 1, 0, 0, 1, 0, loc),      // One second after midnight
		time.Date(2024, 1, 1, 23, 59, 59, 0, loc),   // Last second of day
		time.Date(2024, 12, 31, 23, 59, 59, 0, loc), // End of year
		time.Date(2024, 2, 29, 12, 0, 0, 0, loc),    // Leap day
		time.Date(2024, 3, 10, 2, 30, 0, 0, loc),    // DST transition (varies by region)
		time.Date(2024, 11, 3, 2, 30, 0, 0, loc),    // DST end (varies by region)
	}
}

// GenerateAlarmTimes creates varied alarm time strings
func GenerateAlarmTimes() []string {
	return []string{
		"00:00:00", // Midnight
		"06:00:00", // Early morning
		"09:00:00", // Morning
		"12:00:00", // Noon
		"15:00:00", // Afternoon
		"18:00:00", // Evening
		"21:00:00", // Night
		"23:59:59", // End of day
	}
}

// GenerateStressTestData creates large dataset for stress testing
func GenerateStressTestData(alarmCount, tokensPerUser int) ([]entity.UserAlarm, []entity.WeatherServiceToken) {
	baseTime := time.Now().Add(10 * time.Second)
	alarms := GenerateTestAlarms(alarmCount, baseTime)

	totalTokens := alarmCount * tokensPerUser
	tokens := make([]entity.WeatherServiceToken, 0, totalTokens)

	for i := 0; i < alarmCount; i++ {
		userID := i + 1
		userTokens := GenerateTestFCMTokens(userID, tokensPerUser)
		tokens = append(tokens, userTokens...)
	}

	return alarms, tokens
}
