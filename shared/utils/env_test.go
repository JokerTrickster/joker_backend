package utils

import (
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestGetEnv(t *testing.T) {
	const key = "TEST_GETENV_KEY"
	os.Unsetenv(key)
	defer os.Unsetenv(key)

	tests := []struct {
		name         string
		setValue     string
		useSetenv    bool
		defaultVal   string
		want         string
	}{
		{
			name:       "returns env value when set",
			setValue:   "from-env",
			useSetenv:  true,
			defaultVal: "default",
			want:       "from-env",
		},
		{
			name:       "returns default when env empty",
			setValue:   "",
			useSetenv:  true,
			defaultVal: "fallback",
			want:       "fallback",
		},
		{
			name:       "returns default when env unset",
			setValue:   "",
			useSetenv:  false,
			defaultVal: "my-default",
			want:       "my-default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(key)
			if tt.useSetenv {
				os.Setenv(key, tt.setValue)
			}
			got := GetEnv(key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("GetEnv(%q, %q) = %q, want %q", key, tt.defaultVal, got, tt.want)
			}
		})
	}
}

func TestTimeToEpochMillis(t *testing.T) {
	// Fixed time: 2024-01-15 10:30:45 UTC = 1705314645000 ms
	utcTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	got := TimeToEpochMillis(utcTime)
	want := int64(1705314645000)
	if got != want {
		t.Errorf("TimeToEpochMillis() = %d, want %d", got, want)
	}
}

func TestEpochToTime(t *testing.T) {
	epoch := int64(1705314645) // 2024-01-15 10:30:45 UTC
	got := EpochToTime(epoch)
	want := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	if got.Unix() != want.Unix() {
		t.Errorf("EpochToTime(%d) = %v (unix=%d), want unix=%d", epoch, got, got.Unix(), want.Unix())
	}
}

func TestEpochToTimeMillis(t *testing.T) {
	// 1705314645123 ms = 2024-01-15 10:30:45.123 UTC
	epochMillis := int64(1705314645123)
	got := EpochToTimeMillis(epochMillis)
	want := time.Date(2024, 1, 15, 10, 30, 45, 123000000, time.UTC)
	if got.Unix() != want.Unix() || got.Nanosecond()/1000000 != 123 {
		t.Errorf("EpochToTimeMillis(%d) = %v, want %v", epochMillis, got, want)
	}
}

func TestCtxGenerate(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	startTime := time.Now()
	c.Set("uID", uint(99))
	c.Set("rID", "req-123")
	c.Set("startTime", startTime)
	c.Set("email", "ctx@test.com")

	ctx, userID, email := CtxGenerate(c)

	if userID != 99 {
		t.Errorf("CtxGenerate() userID = %d, want 99", userID)
	}
	if email != "ctx@test.com" {
		t.Errorf("CtxGenerate() email = %q, want %q", email, "ctx@test.com")
	}

	vals := ctx.Value("key")
	if vals == nil {
		t.Fatal("CtxGenerate() context should have key")
	}
	cv, ok := vals.(*CtxValues)
	if !ok {
		t.Fatalf("CtxGenerate() context value type = %T, want *CtxValues", vals)
	}
	if cv.Method != "GET" {
		t.Errorf("CtxValues.Method = %q, want GET", cv.Method)
	}
	if cv.Url != "/api/test" {
		t.Errorf("CtxValues.Url = %q, want /api/test", cv.Url)
	}
	if cv.UserID != 99 {
		t.Errorf("CtxValues.UserID = %d, want 99", cv.UserID)
	}
	if cv.RequestID != "req-123" {
		t.Errorf("CtxValues.RequestID = %q, want req-123", cv.RequestID)
	}
	if cv.Email != "ctx@test.com" {
		t.Errorf("CtxValues.Email = %q, want ctx@test.com", cv.Email)
	}
	if !cv.StartTime.Equal(startTime) {
		t.Errorf("CtxValues.StartTime = %v, want %v", cv.StartTime, startTime)
	}
}
