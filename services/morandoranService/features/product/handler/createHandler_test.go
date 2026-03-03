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

func TestCreateHandler_Create_Success(t *testing.T) {
	t.Log("Create: valid request -> 201 Created")
	e := setupProductTestEcho()
	mockRepo := new(mockProductRepository)
	uc := usecase.NewCreateUseCase(mockRepo, productDefaultTimeout)
	h := NewCreateHandler(uc)

	reqBody := &request.ReqCreateProduct{
		Name: "Product A", Price: 1000, Description: "Desc", Image: "https://img.png", Category: "merch", InStock: boolPtr(true),
	}
	body := productMustJSON(t, reqBody)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Product")).Return(&entity.Product{
		ID: 1, Name: reqBody.Name, Price: reqBody.Price, Description: reqBody.Description,
		Image: reqBody.Image, Category: reqBody.Category, InStock: true,
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Create(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "product-001")
	mockRepo.AssertExpectations(t)
}

func TestCreateHandler_Create_BadRequest_BindError(t *testing.T) {
	t.Log("Create: invalid JSON -> 400 Bad Request")
	e := setupProductTestEcho()
	mockRepo := new(mockProductRepository)
	uc := usecase.NewCreateUseCase(mockRepo, productDefaultTimeout)
	h := NewCreateHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader([]byte("invalid")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Create(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateHandler_Create_InternalError(t *testing.T) {
	t.Log("Create: usecase error -> 500")
	e := setupProductTestEcho()
	mockRepo := new(mockProductRepository)
	uc := usecase.NewCreateUseCase(mockRepo, productDefaultTimeout)
	h := NewCreateHandler(uc)

	reqBody := &request.ReqCreateProduct{
		Name: "P", Price: 100, Description: "D", Image: "i", Category: "c",
	}
	body := productMustJSON(t, reqBody)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Product")).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Create(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	mockRepo.AssertExpectations(t)
}

func boolPtr(b bool) *bool { return &b }
