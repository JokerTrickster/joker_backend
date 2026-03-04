package usecase

import (
	"context"
	"mime/multipart"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadUseCase_Upload_FileSizeExceedsLimit(t *testing.T) {
	uc := NewUploadUseCase(10 * time.Second)
	ctx := context.Background()

	// FileHeader with Size > 5MB (validation runs before Open/S3)
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

func TestUploadUseCase_Upload_S3Required(t *testing.T) {
	t.Skip("Skipping: actual S3 upload test requires S3 client - run as integration test when AWS configured")
}
