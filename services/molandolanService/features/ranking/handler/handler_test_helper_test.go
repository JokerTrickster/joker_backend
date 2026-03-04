package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/model/interface"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const rankingTestUserID = uint(1)

const rankingDefaultTimeout = 5 * time.Second

type mockRankingRepository struct {
	mock.Mock
}

func (m *mockRankingRepository) List(ctx context.Context, gameType string, page, limit int) ([]entity.Ranking, int64, error) {
	args := m.Called(ctx, gameType, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Ranking), args.Get(1).(int64), args.Error(2)
}

func (m *mockRankingRepository) FindByUserAndGame(ctx context.Context, userID uint, gameType string) (*entity.Ranking, error) {
	args := m.Called(ctx, userID, gameType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Ranking), args.Error(1)
}

func (m *mockRankingRepository) Create(ctx context.Context, ranking *entity.Ranking) error {
	args := m.Called(ctx, ranking)
	return args.Error(0)
}

func (m *mockRankingRepository) Update(ctx context.Context, ranking *entity.Ranking) error {
	args := m.Called(ctx, ranking)
	return args.Error(0)
}

func (m *mockRankingRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRankingRepository) GetRank(ctx context.Context, gameType string, clearTimeMs uint) (int, error) {
	args := m.Called(ctx, gameType, clearTimeMs)
	return args.Int(0), args.Error(1)
}

var _ _interface.IRankingRepository = (*mockRankingRepository)(nil)

func setupRankingTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func setupRankingAuthContext(c echo.Context) {
	c.Set("userID", rankingTestUserID)
}

func rankingMustJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func rankingNewRequest(t *testing.T, method, path string, body []byte) *http.Request {
	t.Helper()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return req
}
