package usecase

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/shared/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUploadUseCase(t *testing.T) {
	t.Log("NewUploadUseCase returns usecase with ContextTimeout set")
	timeout := 15 * time.Second
	uc := NewUploadUseCase(timeout)
	require.NotNil(t, uc)
	assert.Equal(t, timeout, uc.ContextTimeout)
}

func TestUploadUseCase_Upload_FileSizeExceedsLimit(t *testing.T) {
	t.Log("Upload: file Size > 5MB -> error")
	uc := NewUploadUseCase(10 * time.Second)
	ctx := context.Background()

	fh := &multipart.FileHeader{
		Filename: "large.jpg",
		Size:     6 * 1024 * 1024, // 6MB
	}

	res, err := uc.Upload(ctx, fh, "general")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "5MB", "error should mention 5MB limit")
	t.Logf("File size validation: %v", err)
}

func TestUploadUseCase_Upload_UnsupportedExtension(t *testing.T) {
	t.Log("Upload: .pdf extension -> error")
	uc := NewUploadUseCase(10 * time.Second)
	ctx := context.Background()

	fh := &multipart.FileHeader{
		Filename: "document.pdf",
		Size:     1024,
	}

	res, err := uc.Upload(ctx, fh, "general")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "unsupported", "error should mention unsupported type")
	t.Logf("Extension validation: %v", err)
}

func TestUploadUseCase_Upload_UnsupportedExtension_EmptyExt(t *testing.T) {
	t.Log("Upload: filename with no extension -> error")
	uc := NewUploadUseCase(10 * time.Second)
	ctx := context.Background()

	fh := &multipart.FileHeader{
		Filename: "noextension",
		Size:     1024,
	}

	res, err := uc.Upload(ctx, fh, "general")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "unsupported")
	t.Logf("Empty extension validation: %v", err)
}

func TestUploadUseCase_Upload_UnsupportedExtension_Gif(t *testing.T) {
	t.Log("Upload: .gif not in allowed list -> error")
	uc := NewUploadUseCase(10 * time.Second)
	ctx := context.Background()

	fh := &multipart.FileHeader{
		Filename: "image.gif",
		Size:     1024,
	}

	res, err := uc.Upload(ctx, fh, "general")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "unsupported")
	t.Logf("GIF extension validation: %v", err)
}

func TestUploadUseCase_Upload_S3ClientNil(t *testing.T) {
	t.Log("Upload: valid file but S3 client nil -> error")
	if aws.S3Client != nil {
		t.Skip("Skipping: S3 is initialized - run without AWS to test S3 nil path")
	}
	uc := NewUploadUseCase(10 * time.Second)
	ctx := context.Background()

	// Create real multipart request to get FileHeader with working Open()
	mpBody := &bytes.Buffer{}
	mpWriter := multipart.NewWriter(mpBody)
	part, err := mpWriter.CreateFormFile("file", "test.jpg")
	require.NoError(t, err)
	_, err = part.Write([]byte{0xff, 0xd8, 0xff}) // minimal jpg prefix
	require.NoError(t, err)
	contentType := mpWriter.FormDataContentType()
	require.NoError(t, mpWriter.Close())

	req := httptest.NewRequest(http.MethodPost, "/upload", mpBody)
	req.Header.Set("Content-Type", contentType)

	_, fh, err := req.FormFile("file")
	require.NoError(t, err)

	res, err := uc.Upload(ctx, fh, "general")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "S3", "error should mention S3")
	t.Logf("S3 nil validation: %v", err)
}
