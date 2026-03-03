package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/request"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubmitHandler_Submit_Unauthorized(t *testing.T) {
	t.Log("Submit: no userID in context -> 401 Unauthorized")
	e := setupRankingTestEcho()
	mockRepo := new(mockRankingRepository)
	uc := usecase.NewSubmitUseCase(mockRepo, rankingDefaultTimeout)
	h := NewSubmitHandler(uc)

	reqBody := &request.ReqSubmitRanking{ClearTimeMs: 5000}
	body := rankingMustJSON(t, reqBody)
	req := httptest.NewRequest(http.MethodPost, "/ranking/puzzle/submit", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/ranking/:gameType/submit")
	c.SetParamNames("gameType")
	c.SetParamValues("puzzle")
	// Do NOT set userID - simulates unauthenticated request

	err := h.Submit(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	assert.Equal(t, "UNAUTHORIZED", he.Message)
	t.Logf("Submit correctly returned 401 when no auth")
}

func TestSubmitHandler_Submit_BadRequest_BindError(t *testing.T) {
	t.Log("Submit: invalid JSON -> 400 Bad Request")
	e := setupRankingTestEcho()
	mockRepo := new(mockRankingRepository)
	uc := usecase.NewSubmitUseCase(mockRepo, rankingDefaultTimeout)
	h := NewSubmitHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/ranking/puzzle/submit", bytes.NewReader([]byte("invalid")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/ranking/:gameType/submit")
	c.SetParamNames("gameType")
	c.SetParamValues("puzzle")
	setupRankingAuthContext(c) // need auth to reach bind

	err := h.Submit(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	t.Logf("Submit bind error -> 400")
}

func TestSubmitHandler_Submit_Success_SkipsDB(t *testing.T) {
	t.Skip("Submit success requires DB for getUserNickname; integration test")
	// getUserNickname uses mysql.GormMysqlDB directly - cannot mock without refactoring
	// Run this test only when TEST_DB_DSN is set and DB has morandoran_users table
}
