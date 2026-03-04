package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/upload/usecase"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUploadEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func TestUploadHandler_Upload_MissingFile(t *testing.T) {
	uc := usecase.NewUploadUseCase(10 * time.Second)
	h := NewUploadHandler(uc)
	e := setupUploadEcho()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("type", "general")
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Upload(c)
	require.Error(t, err)
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusBadRequest, he.Code)
		assert.Equal(t, "file is required", he.Message)
	}
	t.Logf("Missing file returns 400: %v", err)
}

func TestUploadHandler_Upload_S3Required(t *testing.T) {
	t.Skip("Skipping: Upload handler requires S3 - actual upload test needs S3 client initialized")
}
