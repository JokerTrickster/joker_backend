package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

const testGalleryUserID = uint(1)

func setupGalleryTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func setupGalleryAuthContext(c echo.Context) {
	c.Set("userID", testGalleryUserID)
}

func mustJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func newJSONRequest(t *testing.T, method, path string, body []byte) *http.Request {
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

func newMultipartGalleryRequest(t *testing.T, method, path, filename string, fileContent []byte, caption string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	if filename != "" && fileContent != nil {
		part, err := w.CreateFormFile("file", filename)
		require.NoError(t, err)
		_, err = part.Write(fileContent)
		require.NoError(t, err)
	}
	if caption != "" {
		_ = w.WriteField("caption", caption)
	}
	require.NoError(t, w.Close())
	req := httptest.NewRequest(method, path, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}
