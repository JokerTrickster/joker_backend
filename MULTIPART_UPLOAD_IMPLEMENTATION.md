# S3 Multipart Upload Implementation Summary

## Overview
This document describes the complete implementation of S3 multipart upload functionality for handling large file uploads (>5GB) in the joker_backend project.

## Implementation Date
December 5, 2025

## Features Implemented
- Multipart upload initiation
- Presigned URL generation for parts (with batching support)
- Multipart upload completion
- Multipart upload abortion
- Database tracking of upload progress
- Automatic video processing queue integration

## File Structure

### 1. Database Migration
**Location**: `/Users/luxrobo/project/joker_backend/migrations/`

- `000006_create_multipart_uploads_table.up.sql` - Creates multipart_uploads table
- `000006_create_multipart_uploads_table.down.sql` - Rollback migration

**Table Schema**:
```sql
CREATE TABLE multipart_uploads (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    upload_id VARCHAR(255) NOT NULL,
    file_key VARCHAR(500) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    content_type VARCHAR(100) NULL,
    part_size INT NOT NULL,
    total_parts INT NOT NULL,
    status ENUM('initiated', 'uploading', 'completed', 'aborted') DEFAULT 'initiated',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_upload_id (upload_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status)
);
```

### 2. Entity Models
**Location**: `/Users/luxrobo/project/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity/`

- `multipartUpload.go` - Entity definition for multipart upload tracking

**Key Fields**:
- `UploadID` - S3 multipart upload ID
- `FileKey` - S3 object key
- `PartSize` - Size of each part (default 10MB)
- `TotalParts` - Total number of parts
- `Status` - Upload status (initiated, uploading, completed, aborted)

### 3. Request/Response DTOs
**Location**: `/Users/luxrobo/project/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/`

#### Request DTOs (`request/multipartUploadRequest.go`):
- `InitiateMultipartUploadRequestDTO` - Initiate upload session
- `GeneratePresignedURLsRequestDTO` - Request presigned URLs for parts
- `CompleteMultipartUploadRequestDTO` - Complete upload with ETags
- `AbortMultipartUploadRequestDTO` - Abort upload session

#### Response DTOs (`response/multipartUploadResponse.go`):
- `InitiateMultipartUploadResponseDTO` - Upload ID and part info
- `GeneratePresignedURLsResponseDTO` - Array of presigned URLs
- `CompleteMultipartUploadResponseDTO` - File ID and download URL
- `AbortMultipartUploadResponseDTO` - Confirmation message

### 4. AWS S3 Functions
**Location**: `/Users/luxrobo/project/joker_backend/shared/aws/s3.go`

New functions added:
- `CreateMultipartUpload()` - Initiates S3 multipart upload
- `GeneratePresignedUploadPartURL()` - Generates presigned URL for a part
- `CompleteMultipartUpload()` - Completes S3 multipart upload
- `AbortMultipartUpload()` - Aborts S3 multipart upload

### 5. Repository Layer
**Location**: `/Users/luxrobo/project/joker_backend/services/cloudRepositoryService/features/cloudRepository/repository/`

- `multipartUploadRepository.go` - Repository implementation
  - Database operations for upload tracking
  - S3 operations wrapper

### 6. Use Case Layer
**Location**: `/Users/luxrobo/project/joker_backend/services/cloudRepositoryService/features/cloudRepository/usecase/`

- `multipartUploadUseCase.go` - Business logic implementation

**Key Features**:
- Part size calculation (default 10MB, adjustable based on file size)
- Maximum 10,000 parts validation
- File type validation (image/video)
- Automatic video processing queue integration
- Activity logging
- S3 key generation

### 7. Handler Layer
**Location**: `/Users/luxrobo/project/joker_backend/services/cloudRepositoryService/features/cloudRepository/handler/`

- `multipartUploadHandler.go` - HTTP request handlers
- Updated `routes.go` - Route registration

### 8. Interface Definitions
**Location**: `/Users/luxrobo/project/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface/`

Updated files:
- `ICloudRepositoryRepository.go` - Added `IMultipartUploadRepository`
- `ICloudRepositoryUseCase.go` - Added `IMultipartUploadUseCase`
- `ICloudRepositoryHandler.go` - Added `IMultipartUploadHandler`

## API Endpoints

### 1. Initiate Multipart Upload
```
POST /api/v1/files/multipart/initiate
Authorization: Bearer <JWT_TOKEN>

Request:
{
  "file_name": "large-video.mp4",
  "file_size": 10737418240,
  "content_type": "video/mp4",
  "file_type": "video"
}

Response:
{
  "upload_id": "xxx-upload-id-xxx",
  "file_key": "users/1/files/uuid-filename.mp4",
  "part_size": 10485760,
  "total_parts": 1024
}
```

### 2. Generate Presigned URLs for Parts
```
POST /api/v1/files/multipart/presigned-urls
Authorization: Bearer <JWT_TOKEN>

Request:
{
  "upload_id": "xxx-upload-id-xxx",
  "file_key": "users/1/files/uuid-filename.mp4",
  "part_numbers": [1, 2, 3, 4, 5]
}

Response:
{
  "urls": [
    {
      "part_number": 1,
      "url": "https://s3.amazonaws.com/..."
    },
    {
      "part_number": 2,
      "url": "https://s3.amazonaws.com/..."
    }
  ],
  "expires_in": 900
}
```

### 3. Complete Multipart Upload
```
POST /api/v1/files/multipart/complete
Authorization: Bearer <JWT_TOKEN>

Request:
{
  "upload_id": "xxx-upload-id-xxx",
  "file_key": "users/1/files/uuid-filename.mp4",
  "parts": [
    {"part_number": 1, "etag": "etag-1"},
    {"part_number": 2, "etag": "etag-2"}
  ]
}

Response:
{
  "file_id": 123,
  "url": "https://bucket.s3.amazonaws.com/...",
  "size": 10737418240
}
```

### 4. Abort Multipart Upload
```
POST /api/v1/files/multipart/abort
Authorization: Bearer <JWT_TOKEN>

Request:
{
  "upload_id": "xxx-upload-id-xxx",
  "file_key": "users/1/files/uuid-filename.mp4"
}

Response:
{
  "success": true,
  "message": "Upload aborted successfully"
}
```

## Technical Specifications

### Constants
- **Default Part Size**: 10MB (10,485,760 bytes)
- **Min Part Size**: 5MB (S3 minimum, except last part)
- **Max Part Size**: 5GB
- **Max Parts**: 10,000 (S3 limit)
- **Presigned URL Expiration**: 15 minutes

### Part Size Calculation
The system automatically calculates optimal part size:
1. Default: 10MB per part
2. If file requires >10,000 parts, part size is increased
3. Maximum part size is 5GB (S3 limit)
4. Files exceeding 5TB (5GB × 10,000) are rejected

### Upload Flow
1. **Client calls /initiate** → Receives upload_id, file_key, part_size
2. **Client splits file** into parts of part_size bytes
3. **Client requests presigned URLs** (can batch multiple part numbers)
4. **Client uploads parts** directly to S3 using presigned URLs
5. **Client tracks ETags** from S3 responses for each part
6. **Client calls /complete** with all part numbers and ETags
7. **Server creates file record** and triggers video processing (if video)

### Error Handling
- Invalid file types are rejected
- Files >5TB are rejected
- Mismatched upload_id/file_key are rejected
- Already completed uploads cannot be completed again
- Already aborted uploads cannot be completed
- Completed uploads cannot be aborted

### Integration Points
- **JWT Authentication**: All endpoints require valid JWT token
- **User Stats**: Upload activity is logged
- **Video Processing**: Videos automatically queue for processing
- **File Records**: CloudFile entity created on completion

## Testing Recommendations

### Unit Tests
1. Test part size calculation for various file sizes
2. Test validation rules (file type, size limits)
3. Test status transitions (initiated → uploading → completed)
4. Test error cases (invalid upload_id, part number out of range)

### Integration Tests
1. Test complete upload flow (initiate → URLs → complete)
2. Test abort flow (initiate → URLs → abort)
3. Test presigned URL expiration
4. Test concurrent uploads by same user
5. Test video processing queue integration

### Load Tests
1. Test multiple concurrent multipart uploads
2. Test large number of parts (close to 10,000)
3. Test very large files (approaching 5TB)

## Migration Steps

### To Deploy
1. Run database migration: `migrate -path migrations -database "mysql://..." up`
2. Build and deploy service: `go build && deploy`
3. Update API documentation (Swagger)
4. Update client applications to use multipart upload for files >5GB

### Rollback
1. Run down migration: `migrate -path migrations -database "mysql://..." down 1`
2. Remove route registration from `routes.go`
3. Revert to previous deployment

## Backward Compatibility

The existing single-file upload API (`POST /api/v1/files/upload`) remains unchanged and functional for files <5GB. Clients can:
- Continue using single upload for small files (<5GB)
- Use multipart upload for large files (>5GB)
- Optionally use multipart for all files for consistency

## Performance Considerations

### Advantages
- Supports files up to 5TB
- Parallel part uploads improve speed
- Resume capability (client can track uploaded parts)
- Reduced memory footprint (no need to load entire file)

### Resource Usage
- Database: One record per multipart upload session
- S3: Standard multipart upload storage
- Bandwidth: Same as file size (no overhead)

## Security Considerations

1. **Authentication**: All endpoints require valid JWT token
2. **Authorization**: Users can only access their own uploads
3. **Presigned URL Expiration**: 15-minute window limits exposure
4. **File Type Validation**: Only allowed image/video types
5. **User Isolation**: File keys include user ID for separation

## Monitoring & Logging

### Logged Events
- Multipart upload initiated (upload_id, user_id, total_parts)
- Presigned URLs generated (upload_id, parts_count)
- Multipart upload completed (upload_id, file_id)
- Multipart upload aborted (upload_id)
- Video processing enqueued (file_id)

### Metrics to Track
- Average upload completion time
- Success/failure rates
- Aborted upload percentage
- Part size distribution
- File size distribution

## Future Enhancements

### Potential Improvements
1. **Resume Support**: API to list uploaded parts
2. **Progress Tracking**: Real-time upload progress
3. **Auto-cleanup**: Scheduled job to abort stale uploads
4. **Bandwidth Optimization**: Adaptive part sizing
5. **Multi-region**: Cross-region replication support
6. **Checksums**: MD5 validation for parts

## References

### AWS Documentation
- [S3 Multipart Upload Overview](https://docs.aws.amazon.com/AmazonS3/latest/userguide/mpuoverview.html)
- [AWS SDK for Go V2 - S3](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3)

### Project Files
- Entity: `/services/cloudRepositoryService/features/cloudRepository/model/entity/multipartUpload.go`
- Handler: `/services/cloudRepositoryService/features/cloudRepository/handler/multipartUploadHandler.go`
- UseCase: `/services/cloudRepositoryService/features/cloudRepository/usecase/multipartUploadUseCase.go`
- Repository: `/services/cloudRepositoryService/features/cloudRepository/repository/multipartUploadRepository.go`
- AWS S3: `/shared/aws/s3.go`
- Migration: `/migrations/000006_create_multipart_uploads_table.up.sql`

## Support

For questions or issues:
1. Check error logs for detailed error messages
2. Verify S3 credentials and permissions
3. Ensure database migration has been applied
4. Validate JWT token is valid and not expired
5. Check file size and type constraints

---

**Implementation Status**: Complete
**Build Status**: Passing
**Migration Status**: Ready to apply
**Documentation**: Complete
