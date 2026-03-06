package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/upload/model/response"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/upload/usecase"
	"github.com/JokerTrickster/joker_backend/shared/aws"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimalValidJPEG is the smallest valid JPEG (107 bytes) for usecase validation (extension check)
var minimalValidJPEG = []byte{
	0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01,
	0x00, 0x01, 0x00, 0x00, 0xff, 0xdb, 0x00, 0x43, 0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08,
	0x07, 0x07, 0x07, 0x09, 0x09, 0x08, 0x0a, 0x0c, 0x14, 0x0d, 0x0c, 0x0b, 0x0b, 0x0c, 0x19, 0x12,
	0x13, 0x0f, 0x14, 0x1d, 0x1a, 0x1f, 0x1e, 0x1d, 0x1a, 0x1c, 0x1c, 0x20, 0x24, 0x2e, 0x27, 0x24,
	0x22, 0x2e, 0x1b, 0x1c, 0x1c, 0x28, 0x37, 0x29, 0x2c, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1f, 0x27,
	0x39, 0x3d, 0x38, 0x32, 0x3c, 0x2e, 0x33, 0x34, 0x32, 0xff, 0xc0, 0x00, 0x0b, 0x08, 0x00, 0x01,
	0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xff, 0xc4, 0x00, 0x1f, 0x00, 0x00, 0x01, 0x05, 0x01, 0x01,
	0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04,
	0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0xff, 0xc4, 0x00, 0xb5, 0x10, 0x00, 0x02, 0x01, 0x03,
	0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7d, 0x01, 0x02, 0x03, 0x00,
	0x04, 0x11, 0x05, 0x12, 0x21, 0x31, 0x41, 0x06, 0x13, 0x51, 0x61, 0x07, 0x22, 0x71, 0x14, 0x32,
	0x81, 0x91, 0xa1, 0x08, 0x23, 0x42, 0xb1, 0xc1, 0x15, 0x52, 0xd1, 0xf0, 0x24, 0x33, 0x62, 0x72,
	0x82, 0x09, 0x0a, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x34, 0x35,
	0x36, 0x37, 0x38, 0x39, 0x3a, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x53, 0x54, 0x55,
	0x56, 0x57, 0x58, 0x59, 0x5a, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x73, 0x74, 0x75,
	0x76, 0x77, 0x78, 0x79, 0x7a, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8a, 0x92, 0x93, 0x94,
	0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8, 0xa9, 0xaa, 0xb2,
	0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7, 0xc8, 0xc9,
	0xca, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5, 0xe6,
	0xe7, 0xe8, 0xe9, 0xea, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xfa, 0xff, 0xda,
	0x00, 0x0c, 0x03, 0x01, 0x00, 0x02, 0x11, 0x03, 0x11, 0x00, 0x3f, 0x00, 0xfb, 0x28, 0xff, 0xd9,
}

func setupUploadEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func newMultipartUploadRequest(t *testing.T, filename string, fileContent []byte, fileType string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	if filename != "" && fileContent != nil {
		part, err := w.CreateFormFile("file", filename)
		require.NoError(t, err)
		_, err = part.Write(fileContent)
		require.NoError(t, err)
	}
	if fileType != "" {
		_ = w.WriteField("type", fileType)
	}
	require.NoError(t, w.Close())
	req := httptest.NewRequest(http.MethodPost, "/api/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestNewUploadHandler(t *testing.T) {
	t.Log("NewUploadHandler returns handler with UseCase set")
	uc := usecase.NewUploadUseCase(10 * time.Second)
	h := NewUploadHandler(uc)
	require.NotNil(t, h)
	require.NotNil(t, h.UseCase)
	assert.Same(t, uc, h.UseCase)
}

func TestUploadHandler_Upload_MissingFile(t *testing.T) {
	t.Log("Upload: multipart without file part -> 400 file is required")
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

func TestUploadHandler_Upload_InvalidContentType(t *testing.T) {
	t.Log("Upload: request with application/json instead of multipart -> 400")
	uc := usecase.NewUploadUseCase(10 * time.Second)
	h := NewUploadHandler(uc)
	e := setupUploadEcho()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Upload(c)
	require.Error(t, err)
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusBadRequest, he.Code)
		assert.Equal(t, "file is required", he.Message)
	}
	t.Logf("Invalid content-type returns 400: %v", err)
}

func TestUploadHandler_Upload_UseCaseError_UnsupportedExtension(t *testing.T) {
	t.Log("Upload: file with .pdf extension -> 400 from usecase")
	uc := usecase.NewUploadUseCase(10 * time.Second)
	h := NewUploadHandler(uc)
	e := setupUploadEcho()

	req := newMultipartUploadRequest(t, "document.pdf", []byte("pdf content"), "general")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Upload(c)
	require.Error(t, err)
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusBadRequest, he.Code)
		assert.Contains(t, he.Message, "unsupported", "error message should mention unsupported file type")
	}
	t.Logf("Unsupported extension returns 400: %v", err)
}

func TestUploadHandler_Upload_UseCaseError_FileSizeExceeded(t *testing.T) {
	t.Log("Upload: file > 5MB -> 400 from usecase")
	if testing.Short() {
		t.Skip("Skipping 5MB upload test in short mode")
	}
	uc := usecase.NewUploadUseCase(10 * time.Second)
	h := NewUploadHandler(uc)
	e := setupUploadEcho()

	oversizedContent := bytes.Repeat([]byte{0x00}, 5*1024*1024+1) // 5MB+1
	req := newMultipartUploadRequest(t, "large.jpg", oversizedContent, "general")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Upload(c)
	require.Error(t, err)
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusBadRequest, he.Code)
		assert.Contains(t, he.Message, "5MB", "error message should mention size limit")
	}
	t.Logf("File size exceeded returns 400: %v", err)
}

func TestUploadHandler_Upload_UseCaseError_S3NotInitialized(t *testing.T) {
	t.Log("Upload: valid file but S3 not initialized -> 400")
	if aws.S3Client != nil {
		t.Skip("Skipping: S3 is initialized - run in env without AWS to test S3 nil path")
	}
	uc := usecase.NewUploadUseCase(10 * time.Second)
	h := NewUploadHandler(uc)
	e := setupUploadEcho()

	req := newMultipartUploadRequest(t, "test.jpg", minimalValidJPEG, "general")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Upload(c)
	require.Error(t, err)
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusBadRequest, he.Code)
		assert.Contains(t, he.Message, "S3", "error message should mention S3")
	}
	t.Logf("S3 not initialized returns 400: %v", err)
}

func TestUploadHandler_Upload_DefaultTypeWhenEmpty(t *testing.T) {
	t.Log("Upload: multipart with file but no type field -> defaults to general, passes to usecase")
	if aws.S3Client != nil {
		t.Skip("Skipping: S3 is initialized - this test exercises default type path; usecase would succeed with S3")
	}
	uc := usecase.NewUploadUseCase(10 * time.Second)
	h := NewUploadHandler(uc)
	e := setupUploadEcho()

	req := newMultipartUploadRequest(t, "test.jpg", minimalValidJPEG, "")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Upload(c)
	require.Error(t, err, "expect error when S3 not initialized")
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusBadRequest, he.Code)
	}
	t.Logf("Default type path exercised: %v", err)
}

func TestUploadHandler_Upload_Success(t *testing.T) {
	t.Log("Upload: valid file with S3 initialized -> 200 with URL")
	if aws.S3Client == nil {
		t.Skip("Skipping: S3 not initialized - run with AWS configured for full E2E test")
	}
	uc := usecase.NewUploadUseCase(10 * time.Second)
	h := NewUploadHandler(uc)
	e := setupUploadEcho()

	req := newMultipartUploadRequest(t, "test.jpg", minimalValidJPEG, "avatar")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Upload(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var res response.ResUpload
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.NotEmpty(t, res.URL)
	assert.Contains(t, res.URL, "avatar/", "URL should include file type in path")
	t.Logf("Success: url=%s", res.URL)
}
