package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/product/model/entity"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/product/model/request"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/product/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

func TestUpdateHandler_Update_Success(t *testing.T) {
	t.Log("Update: valid request -> 200 OK")
	e := setupProductTestEcho()
	mockRepo := new(mockProductRepository)
	uc := usecase.NewUpdateUseCase(mockRepo, productDefaultTimeout)
	h := NewUpdateHandler(uc)

	name := "Updated Name"
	reqBody := &request.ReqUpdateProduct{Name: &name}
	body := productMustJSON(t, reqBody)
	mockRepo.On("Update", mock.Anything, uint(1), mock.AnythingOfType("map[string]interface {}")).Return(&entity.Product{
		ID: 1, Name: "Updated Name", Price: 100, Description: "D", Image: "i", Category: "c", InStock: true,
	}, nil)

	req := httptest.NewRequest(http.MethodPut, "/products/1", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/products/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := h.Update(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "product-001")
	mockRepo.AssertExpectations(t)
}

func TestUpdateHandler_Update_InvalidID(t *testing.T) {
	t.Log("Update: invalid ID -> 400")
	e := setupProductTestEcho()
	mockRepo := new(mockProductRepository)
	uc := usecase.NewUpdateUseCase(mockRepo, productDefaultTimeout)
	h := NewUpdateHandler(uc)

	reqBody := &request.ReqUpdateProduct{}
	body := productMustJSON(t, reqBody)
	req := httptest.NewRequest(http.MethodPut, "/products/xyz", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/products/:id")
	c.SetParamNames("id")
	c.SetParamValues("xyz")

	err := h.Update(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdateHandler_Update_NotFound(t *testing.T) {
	t.Log("Update: not found -> 404")
	e := setupProductTestEcho()
	mockRepo := new(mockProductRepository)
	uc := usecase.NewUpdateUseCase(mockRepo, productDefaultTimeout)
	h := NewUpdateHandler(uc)

	name := "X"
	reqBody := &request.ReqUpdateProduct{Name: &name}
	body := productMustJSON(t, reqBody)
	mockRepo.On("Update", mock.Anything, uint(999), mock.Anything).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodPut, "/products/999", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/products/:id")
	c.SetParamNames("id")
	c.SetParamValues("999")

	err := h.Update(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, he.Code)
	mockRepo.AssertExpectations(t)
}
