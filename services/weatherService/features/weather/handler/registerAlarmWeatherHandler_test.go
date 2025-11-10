package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/request"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/repository"
	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/usecase"
	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func setupTestDB(t *testing.T) *gorm.DB {
	// 테스트용 데이터베이스 연결
	dsn := "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

func generateTestToken(t *testing.T, userID uint, email string) string {
	if err := jwt.InitJwt(); err != nil {
		t.Fatalf("Failed to initialize JWT: %v", err)
	}

	accessToken, _, _, _, err := jwt.GenerateToken(email, userID)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}
	return accessToken
}

func TestRegisterAlarmWeatherHandler_Success(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	repo := repository.NewRegisterAlarmWeatherRepository(db)
	uc := usecase.NewRegisterAlarmWeatherUseCase(repo, 10*time.Second)
	handler := &RegisterAlarmWeatherHandler{UseCase: uc}

	// Test data
	userID := uint(1)
	email := "test@example.com"
	token := generateTestToken(t, userID, email)

	reqBody := request.ReqRegisterAlarm{
		AlarmTime: "09:00",
		Region:    "서울시 관악구",
		FCMToken:  "test-fcm-token-12345",
		DeviceID:  "test-device-001",
	}

	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/v0.1/weather/register", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set userID in context (normally done by JWT middleware)
	c.Set("userID", userID)
	c.Set("email", email)

	// Execute
	err = handler.RegisterAlarm(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	t.Logf("Response: %s", rec.Body.String())
}

func TestRegisterAlarmWeatherHandler_InvalidTimeFormat(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	repo := repository.NewRegisterAlarmWeatherRepository(db)
	uc := usecase.NewRegisterAlarmWeatherUseCase(repo, 10*time.Second)
	handler := &RegisterAlarmWeatherHandler{UseCase: uc}

	userID := uint(1)
	email := "test@example.com"

	reqBody := request.ReqRegisterAlarm{
		AlarmTime: "25:00", // Invalid hour
		Region:    "서울시 관악구",
		FCMToken:  "test-fcm-token",
		DeviceID:  "test-device",
	}

	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v0.1/weather/register", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("userID", userID)
	c.Set("email", email)

	// Execute
	err = handler.RegisterAlarm(c)

	// Assert
	assert.Error(t, err)
	t.Logf("Error response: %v", err)
}

func TestRegisterAlarmWeatherHandler_MissingUserID(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	repo := repository.NewRegisterAlarmWeatherRepository(db)
	uc := usecase.NewRegisterAlarmWeatherUseCase(repo, 10*time.Second)
	handler := &RegisterAlarmWeatherHandler{UseCase: uc}

	reqBody := request.ReqRegisterAlarm{
		AlarmTime: "09:00",
		Region:    "서울시 관악구",
		FCMToken:  "test-fcm-token",
		DeviceID:  "test-device",
	}

	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v0.1/weather/register", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Don't set userID - should fail

	// Execute
	err = handler.RegisterAlarm(c)

	// Assert
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}
