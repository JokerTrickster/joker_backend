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

func TestDetailHandler_Detail_Success(t *testing.T) {
	t.Log("Detail: valid ID -> 200 OK")
	e := setupProductTestEcho()
	mockRepo := new(mockProductRepository)
	uc := usecase.NewDetailUseCase(mockRepo, productDefaultTimeout)
	h := NewDetailHandler(uc)

	mockRepo.On("FindByID", mock.Anything, uint(1)).Return(&entity.Product{
		ID: 1, Name: "Product A", Price: 1000, Description: "D", Image: "i", Category: "c", InStock: true,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/products/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := h.Detail(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "product-001")
	mockRepo.AssertExpectations(t)
}

func TestDetailHandler_Detail_InvalidID(t *testing.T) {
	t.Log("Detail: invalid ID -> 400")
	e := setupProductTestEcho()
	mockRepo := new(mockProductRepository)
	uc := usecase.NewDetailUseCase(mockRepo, productDefaultTimeout)
	h := NewDetailHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/products/x", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/products/:id")
	c.SetParamNames("id")
	c.SetParamValues("x")

	err := h.Detail(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	mockRepo.AssertNotCalled(t, "FindByID")
}

func TestDetailHandler_Detail_NotFound(t *testing.T) {
	t.Log("Detail: not found -> 404")
	e := setupProductTestEcho()
	mockRepo := new(mockProductRepository)
	uc := usecase.NewDetailUseCase(mockRepo, productDefaultTimeout)
	h := NewDetailHandler(uc)

	mockRepo.On("FindByID", mock.Anything, uint(999)).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/products/999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/products/:id")
	c.SetParamNames("id")
	c.SetParamValues("999")

	err := h.Detail(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, he.Code)
	mockRepo.AssertExpectations(t)
}
