package aws

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"image/png"
	"math/rand"
	"mime/multipart"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/disintegration/imaging"
)

var imgMeta = map[ImgType]imgMetaStruct{
	ImgCloudRepository: {
		bucket:     func() string { return "joker-cloud-repository-dev" },
		domain:     func() string { return "board-game-app.s3.ap-south-1.amazonaws.com" },
		path:       "images",
		expireTime: 10 * time.Hour,
	},
	ImgGallery: {
		bucket:     func() string { return "molandolan-gallery" },
		domain:     func() string { return "molandolan-gallery.s3.ap-south-1.amazonaws.com" },
		path:       "gallery",
		expireTime: 24 * time.Hour,
	},
}

func ImageUpload(ctx context.Context, file *multipart.FileHeader, filename string, imgType ImgType) error {
	meta, ok := imgMeta[imgType]
	if !ok {
		return fmt.Errorf("not available meta info for imgType - %v", imgType)
	}
	bucket := meta.bucket()

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("fail to open file - %v", err)
	}
	defer src.Close()

	img, err := imaging.Decode(src)
	if err != nil {
		return fmt.Errorf("fail to load image - %v", err)
	}

	// 비율 유지하며 리사이즈
	if meta.width > 0 && meta.height > 0 {
		img = imaging.Fit(img, meta.width, meta.height, imaging.Lanczos)
	} else if meta.width > 0 {
		img = imaging.Resize(img, meta.width, 0, imaging.Lanczos)
	} else if meta.height > 0 {
		img = imaging.Resize(img, 0, meta.height, imaging.Lanczos)
	}

	buf := new(bytes.Buffer)
	if err := imaging.Encode(buf, img, imaging.PNG, imaging.PNGCompressionLevel(png.BestCompression)); err != nil {
		return fmt.Errorf("fail to encode png image - %v", err)
	}

	_, err = awsClientS3Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fmt.Sprintf("%s/%s", meta.path, filename)),
		Body:        buf,
		ContentType: aws.String("image/png"),
	})
	if err != nil {
		return fmt.Errorf("fail to upload image to s3 - bucket:%s / key:%s/%s", bucket, meta.path, filename)
	}
	return nil
}

func ImageGetSignedURL(ctx context.Context, fileName string, imgType ImgType) (string, error) {
	meta, ok := imgMeta[imgType]
	if !ok {
		return "", fmt.Errorf("not available meta info for imgType - %v", imgType)
	}
	presignClient := s3.NewPresignClient(awsClientS3)

	key := fmt.Sprintf("%s/%s", meta.path, fileName)
	presignParams := &s3.GetObjectInput{
		Bucket: aws.String(meta.bucket()),
		Key:    aws.String(key),
	}

	presignResult, err := presignClient.PresignGetObject(ctx, presignParams, s3.WithPresignExpires(meta.expireTime))
	if err != nil {
		return "", err
	}

	return html.UnescapeString(presignResult.URL), nil
}

func ImageDelete(ctx context.Context, fileName string, imgType ImgType) error {
	meta, ok := imgMeta[imgType]
	if !ok {
		return fmt.Errorf("not available meta info for imgType - %v", imgType)
	}

	bucket := meta.bucket()
	key := fmt.Sprintf("%s/%s", meta.path, fileName)

	if _, err := awsClientS3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}); err != nil {
		return fmt.Errorf("fail to delete image from s3 - bucket:%s, key:%s", bucket, key)
	}

	return nil
}

func RawFileUpload(ctx context.Context, file *multipart.FileHeader, key, contentType string, imgType ImgType) (string, error) {
	meta, ok := imgMeta[imgType]
	if !ok {
		return "", fmt.Errorf("not available meta info for imgType - %v", imgType)
	}
	bucket := meta.bucket()
	fullKey := fmt.Sprintf("%s/%s", meta.path, key)

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	_, err = awsClientS3Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fullKey),
		Body:        src,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	url := fmt.Sprintf("https://%s/%s", meta.domain(), fullKey)
	return url, nil
}

func ThumbnailUpload(ctx context.Context, file *multipart.FileHeader, key string, imgType ImgType) (string, error) {
	meta, ok := imgMeta[imgType]
	if !ok {
		return "", fmt.Errorf("not available meta info for imgType - %v", imgType)
	}
	bucket := meta.bucket()
	thumbKey := fmt.Sprintf("%s/thumb_%s", meta.path, key)

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	img, err := imaging.Decode(src)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	thumb := imaging.Fit(img, 400, 400, imaging.Lanczos)
	buf := new(bytes.Buffer)
	if err := imaging.Encode(buf, thumb, imaging.PNG, imaging.PNGCompressionLevel(png.BestCompression)); err != nil {
		return "", fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	_, err = awsClientS3Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(thumbKey),
		Body:        buf,
		ContentType: aws.String("image/png"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload thumbnail to S3: %w", err)
	}

	url := fmt.Sprintf("https://%s/%s", meta.domain(), thumbKey)
	return url, nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func FileNameGenerateRandom() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// GeneratePresignedUploadURL generates a presigned URL for uploading to S3
func GeneratePresignedUploadURL(ctx context.Context, bucket, key, contentType string, expiration time.Duration) (string, error) {
	if awsClientS3 == nil {
		return "", fmt.Errorf("AWS S3 client not initialized - check AWS configuration or IS_LOCAL setting")
	}
	presignClient := s3.NewPresignClient(awsClientS3)

	presignParams := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}

	presignResult, err := presignClient.PresignPutObject(ctx, presignParams, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return html.UnescapeString(presignResult.URL), nil
}

// GeneratePresignedDownloadURL generates a presigned URL for downloading from S3
func GeneratePresignedDownloadURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	if awsClientS3 == nil {
		return "", fmt.Errorf("AWS S3 client not initialized - check AWS configuration or IS_LOCAL setting")
	}
	presignClient := s3.NewPresignClient(awsClientS3)

	presignParams := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	presignResult, err := presignClient.PresignGetObject(ctx, presignParams, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return html.UnescapeString(presignResult.URL), nil
}

// GeneratePresignedDownloadURLWithFilename generates a presigned URL for downloading from S3 with Content-Disposition header
func GeneratePresignedDownloadURLWithFilename(ctx context.Context, bucket, key, filename string, expiration time.Duration) (string, error) {
	if awsClientS3 == nil {
		return "", fmt.Errorf("AWS S3 client not initialized - check AWS configuration or IS_LOCAL setting")
	}
	presignClient := s3.NewPresignClient(awsClientS3)

	presignParams := &s3.GetObjectInput{
		Bucket:                     aws.String(bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String(fmt.Sprintf(`attachment; filename="%s"`, filename)),
	}

	presignResult, err := presignClient.PresignGetObject(ctx, presignParams, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return html.UnescapeString(presignResult.URL), nil
}

// DeleteObject deletes an object from S3
func DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := awsClientS3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3 - bucket: %s, key: %s: %w", bucket, key, err)
	}

	return nil
}

// CreateMultipartUpload initiates a multipart upload
func CreateMultipartUpload(ctx context.Context, bucket, key, contentType string) (string, error) {
	if awsClientS3 == nil {
		return "", fmt.Errorf("AWS S3 client not initialized - check AWS configuration or IS_LOCAL setting")
	}

	output, err := awsClientS3.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create multipart upload: %w", err)
	}

	return *output.UploadId, nil
}

// GeneratePresignedUploadPartURL generates a presigned URL for uploading a part
func GeneratePresignedUploadPartURL(ctx context.Context, bucket, key, uploadID string, partNumber int, expiration time.Duration) (string, error) {
	if awsClientS3 == nil {
		return "", fmt.Errorf("AWS S3 client not initialized - check AWS configuration or IS_LOCAL setting")
	}
	presignClient := s3.NewPresignClient(awsClientS3)

	presignParams := &s3.UploadPartInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(key),
		PartNumber: aws.Int32(int32(partNumber)),
		UploadId:   aws.String(uploadID),
	}

	presignResult, err := presignClient.PresignUploadPart(ctx, presignParams, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload part URL: %w", err)
	}

	return html.UnescapeString(presignResult.URL), nil
}

// CompleteMultipartUpload completes a multipart upload
func CompleteMultipartUpload(ctx context.Context, bucket, key, uploadID string, parts []CompletedPart) error {
	if awsClientS3 == nil {
		return fmt.Errorf("AWS S3 client not initialized - check AWS configuration or IS_LOCAL setting")
	}

	// Convert to AWS SDK types
	var completedParts []types.CompletedPart
	for _, part := range parts {
		completedParts = append(completedParts, types.CompletedPart{
			ETag:       aws.String(part.ETag),
			PartNumber: aws.Int32(int32(part.PartNumber)),
		})
	}

	_, err := awsClientS3.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	return nil
}

// AbortMultipartUpload aborts a multipart upload
func AbortMultipartUpload(ctx context.Context, bucket, key, uploadID string) error {
	if awsClientS3 == nil {
		return fmt.Errorf("AWS S3 client not initialized - check AWS configuration or IS_LOCAL setting")
	}

	_, err := awsClientS3.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
	if err != nil {
		return fmt.Errorf("failed to abort multipart upload: %w", err)
	}

	return nil
}

// CompletedPart represents a completed multipart upload part
type CompletedPart struct {
	PartNumber int
	ETag       string
}
