# Cloud Repository Service

Image and video file management service using AWS S3 with presigned URLs for direct client uploads/downloads.

## Features

- 📤 **Presigned Upload URLs**: Front-end directly uploads files to S3
- 📦 **Batch Upload**: Upload up to 30 files at once
- 📥 **Presigned Download URLs**: Secure temporary download links
- 🖼️ **Image Support**: All image/* MIME types (JPEG, PNG, GIF, WebP, HEIC, HEIF, etc.)
- 🎥 **Video Support**: All video/* MIME types (MP4, WebM, AVI, MOV, QuickTime, M4V, etc.)
- 📱 **iPhone Compatible**: Supports HEIC/HEIF photos and MOV/M4V videos natively
- 📊 **File Management**: List, delete files with pagination
- 🔐 **User Isolation**: Each user can only access their own files
- 🗄️ **Database Tracking**: Metadata stored in MySQL

## Architecture

```
features/cloudRepository/
├── handler/           # HTTP handlers (Echo)
│   ├── cloudRepository.go
│   └── routes.go
├── model/            # Domain models & DTOs
│   ├── cloudFile.go
│   └── dto.go
├── repository/       # S3 & database access
│   └── cloudRepository.go
└── usecase/         # Business logic
    └── cloudRepository.go
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/files/upload` | Request presigned upload URL (single file) |
| POST | `/api/v1/files/upload/batch` | Request presigned upload URLs (batch, max 30) |
| GET | `/api/v1/files` | List user's files (filtering & pagination) |
| GET | `/api/v1/files/:id/download` | Get presigned download URL |
| DELETE | `/api/v1/files/:id` | Delete file (soft delete) |

## Filtering & Sorting

The `GET /api/v1/files` endpoint supports the following query parameters:

| Parameter | Description | Example |
|-----------|-------------|---------|
| `keyword` | Search in filename or tags | `?keyword=vacation` |
| `tags` | Filter by specific tags (multiple allowed) | `?tags=travel&tags=2023` |
| `file_type` | Filter by type (`image` or `video`) | `?file_type=image` |
| `sort` | Sort order (`latest`, `oldest`, `name`, `size`) | `?sort=size` |
| `start_date` | Filter by start date (YYYY-MM-DD) | `?start_date=2023-01-01` |
| `end_date` | Filter by end date (YYYY-MM-DD) | `?end_date=2023-12-31` |
| `page` | Page number (default: 1) | `?page=2` |
| `page_size` | Page size (default: 20, max: 100) | `?page_size=50` |

## Upload Flow

### Single File Upload
1. **Client** → `POST /api/v1/files/upload` with file metadata
2. **Server** → Returns presigned upload URL + file ID
3. **Client** → Directly uploads file to S3 using presigned URL
4. **Client** → (Optional) Call download endpoint to get file

### Batch Upload (Max 30 files)
1. **Client** → `POST /api/v1/files/upload/batch` with array of file metadata
   ```json
   {
     "files": [
       {
         "file_name": "photo1.jpg",
         "content_type": "image/jpeg",
         "file_type": "image",
         "file_size": 1024000
       },
       {
         "file_name": "iphone_photo.heic",
         "content_type": "image/heic",
         "file_type": "image",
         "file_size": 2048000
       },
       {
         "file_name": "video1.mp4",
         "content_type": "video/mp4",
         "file_type": "video",
         "file_size": 5024000
       },
       {
         "file_name": "iphone_video.mov",
         "content_type": "video/quicktime",
         "file_type": "video",
         "file_size": 8192000
       }
     ]
   }
   ```
2. **Server** → Returns array of presigned upload URLs
3. **Client** → Uploads each file to S3 in parallel using presigned URLs

### Supported MIME Types

**Images** (all `image/*` MIME types):
- Standard formats: JPEG, PNG, GIF, WebP, BMP, TIFF
- iPhone formats: HEIC, HEIF, HEIC-sequence, HEIF-sequence
- Any other image/* MIME type

**Videos** (all `video/*` MIME types):
- Standard formats: MP4, WebM, AVI
- iPhone formats: MOV (QuickTime), M4V
- Any other video/* MIME type

**Validation Logic**:
- Uses prefix-based validation (`strings.HasPrefix`)
- All `image/*` content types are accepted for image uploads
- All `video/*` content types are accepted for video uploads
- Invalid MIME types (PDF, text, JSON, etc.) are rejected with clear error messages

## Download Flow

1. **Client** → `GET /api/v1/files/:id/download`
2. **Server** → Returns presigned download URL
3. **Client** → Directly downloads from S3

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
CLOUD_REPOSITORY_BUCKET=cloud-repository-dev
AWS_REGION=ap-south-1
DB_HOST=localhost
DB_PORT=3306
DB_NAME=cloud_repository
PORT=8080
```

## Quick Start

```bash
# Install dependencies
make deps

# Run locally
make run

# Build binary
make build

# Run tests
make test

# Docker
make docker-build
make docker-run
```

## S3 Bucket Structure

```
cloud-repository-dev/
├── image/
│   └── {userID}/
│       └── {uuid}_{random}.{ext}
└── video/
    └── {userID}/
        └── {uuid}_{random}.{ext}
```

## Database Schema

```sql
CREATE TABLE cloud_files (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    s3_key VARCHAR(512) NOT NULL UNIQUE,
    file_type VARCHAR(20) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_file_type (file_type),
    INDEX idx_deleted_at (deleted_at)
);
```

## Shared AWS Helpers

Located in `shared/aws/s3.go`:
- `GeneratePresignedUploadURL()` - Upload URL generation
- `GeneratePresignedDownloadURL()` - Download URL generation
- `DeleteObject()` - S3 object deletion

## TODO

- [ ] Add file size validation on upload completion
- [ ] Implement file virus scanning
- [ ] Add file thumbnail generation for images
- [ ] Implement file sharing between users
- [ ] Add batch upload/delete operations
- [ ] Implement file versioning
