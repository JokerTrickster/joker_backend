package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/product/model/entity"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/product/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

func TestListHandler_List_Success(t *testing.T) {
	t.Log("List: valid request -> 200 OK")
	e := setupProductTestEcho()
	mockRepo := new(mockProductRepository)
	uc := usecase.NewListUseCase(mockRepo, productDefaultTimeout)
	h := NewListHandler(uc)

	items := []entity.Product{
		{ID: 1, Name: "P1", Price: 100, Description: "D", Image: "i", Category: "c", InStock: true},
	}
	mockRepo.On("List", mock.Anything, 1, 20, "", (*bool)(nil)).Return(items, int64(1), nil)

	req := httptest.NewRequest(http.MethodGet, "/products?page=1&limit=20", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.List(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "product-001")
	mockRepo.AssertExpectations(t)
}

func TestListHandler_List_UseCaseError(t *testing.T) {
	t.Log("List: usecase error -> 500")
	e := setupProductTestEcho()
	mockRepo := new(mockProductRepository)
	uc := usecase.NewListUseCase(mockRepo, productDefaultTimeout)
	h := NewListHandler(uc)

	mockRepo.On("List", mock.Anything, 1, 20, "", (*bool)(nil)).Return(nil, int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.List(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	mockRepo.AssertExpectations(t)
}
