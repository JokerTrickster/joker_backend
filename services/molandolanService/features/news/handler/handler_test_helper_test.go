package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/interface"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testUserID = uint(1)

const defaultTimeout = 5 * time.Second

type mockNewsRepository struct {
	mock.Mock
}

func (m *mockNewsRepository) List(ctx context.Context, page, limit int, category string) ([]entity.News, int64, error) {
	args := m.Called(ctx, page, limit, category)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.News), args.Get(1).(int64), args.Error(2)
}

func (m *mockNewsRepository) FindByID(ctx context.Context, id uint) (*entity.News, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockNewsRepository) Create(ctx context.Context, news *entity.News) (*entity.News, error) {
	args := m.Called(ctx, news)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockNewsRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.News, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockNewsRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

var _ _interface.INewsRepository = (*mockNewsRepository)(nil)

func setupNewsTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func setupNewsAuthContext(c echo.Context) {
	c.Set("userID", testUserID)
}

func newsMustJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func newsNewJSONRequest(t *testing.T, method, path string, body []byte) *http.Request {
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
