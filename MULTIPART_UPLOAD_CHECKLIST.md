# Multipart Upload Implementation Checklist

## Pre-Deployment Verification

### 1. Code Review
- [x] Entity models created (`multipartUpload.go`)
- [x] Request/Response DTOs created
- [x] AWS S3 functions implemented
- [x] Repository layer implemented
- [x] Use case layer implemented
- [x] Handler layer implemented
- [x] Routes registered
- [x] Interfaces updated
- [x] Build successful (no compilation errors)

### 2. Database
- [ ] Run migration: `migrate -path migrations -database "mysql://user:pass@host/db" up`
- [ ] Verify table created: `SHOW TABLES LIKE 'multipart_uploads';`
- [ ] Verify indexes: `SHOW INDEX FROM multipart_uploads;`

### 3. AWS Configuration
- [ ] Verify S3 bucket permissions allow multipart uploads
- [ ] Verify AWS credentials are configured
- [ ] Test S3 connection in service startup

### 4. Testing
- [ ] Test initiate endpoint with curl/Postman
- [ ] Test presigned URL generation
- [ ] Test actual file upload to presigned URL
- [ ] Test complete endpoint with valid ETags
- [ ] Test abort endpoint
- [ ] Test error cases (invalid upload_id, etc.)

## Quick Test Commands

### 1. Initiate Upload
```bash
curl -X POST http://localhost:8080/api/v1/files/multipart/initiate \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "test-video.mp4",
    "file_size": 104857600,
    "content_type": "video/mp4",
    "file_type": "video"
  }'
```

### 2. Generate Presigned URLs
```bash
curl -X POST http://localhost:8080/api/v1/files/multipart/presigned-urls \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "upload_id": "UPLOAD_ID_FROM_STEP_1",
    "file_key": "FILE_KEY_FROM_STEP_1",
    "part_numbers": [1, 2, 3]
  }'
```

### 3. Upload Part to S3
```bash
curl -X PUT "PRESIGNED_URL_FROM_STEP_2" \
  --upload-file part1.bin \
  -v
```
Note: Save the ETag from response header

### 4. Complete Upload
```bash
curl -X POST http://localhost:8080/api/v1/files/multipart/complete \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "upload_id": "UPLOAD_ID_FROM_STEP_1",
    "file_key": "FILE_KEY_FROM_STEP_1",
    "parts": [
      {"part_number": 1, "etag": "ETAG_FROM_S3_RESPONSE"},
      {"part_number": 2, "etag": "ETAG_FROM_S3_RESPONSE"}
    ]
  }'
```

### 5. Abort Upload (Optional)
```bash
curl -X POST http://localhost:8080/api/v1/files/multipart/abort \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "upload_id": "UPLOAD_ID_TO_ABORT",
    "file_key": "FILE_KEY_TO_ABORT"
  }'
```

## Database Migration Command

```bash
# Up migration
migrate -path /Users/luxrobo/project/joker_backend/migrations \
        -database "mysql://user:password@tcp(host:3306)/database_name" \
        up

# Down migration (if needed)
migrate -path /Users/luxrobo/project/joker_backend/migrations \
        -database "mysql://user:password@tcp(host:3306)/database_name" \
        down 1
```

## Verification Queries

### Check multipart_uploads table
```sql
-- Verify table structure
DESCRIBE multipart_uploads;

-- Check indexes
SHOW INDEX FROM multipart_uploads;

-- View recent uploads
SELECT * FROM multipart_uploads ORDER BY created_at DESC LIMIT 10;

-- Check status distribution
SELECT status, COUNT(*) as count FROM multipart_uploads GROUP BY status;
```

## Integration Verification

### 1. File Creation
After completing multipart upload, verify:
```sql
-- Check if CloudFile was created
SELECT * FROM cloud_files
WHERE s3_key = 'FILE_KEY_FROM_UPLOAD'
ORDER BY created_at DESC;
```

### 2. Video Processing
For video uploads, verify queue:
```sql
-- Check if processing was triggered
SELECT * FROM cloud_files
WHERE file_type = 'video'
AND processing_status = 'pending'
ORDER BY created_at DESC;
```

### 3. Activity Logging
```sql
-- Check activity logs
SELECT * FROM activity_logs
WHERE activity_type = 'upload'
ORDER BY created_at DESC LIMIT 10;
```

## Common Issues & Solutions

### Issue: "AWS S3 client not initialized"
**Solution**:
- Check AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables
- Verify IS_LOCAL setting if in local development

### Issue: "upload not found"
**Solution**:
- Verify upload_id matches the one from initiate response
- Check user_id (uploads are user-specific)
- Ensure upload hasn't been aborted or completed

### Issue: "invalid part number"
**Solution**:
- Part numbers must be 1 to total_parts (from initiate response)
- Verify part_numbers array in request

### Issue: "failed to complete multipart upload"
**Solution**:
- Ensure all parts were uploaded successfully
- Verify ETags are correct (from S3 upload responses)
- Check parts are in ascending order by part_number

## Performance Monitoring

### Metrics to Track
```sql
-- Average upload completion time
SELECT
    AVG(TIMESTAMPDIFF(SECOND, created_at, updated_at)) as avg_seconds
FROM multipart_uploads
WHERE status = 'completed';

-- Success rate
SELECT
    status,
    COUNT(*) as count,
    (COUNT(*) * 100.0 / (SELECT COUNT(*) FROM multipart_uploads)) as percentage
FROM multipart_uploads
GROUP BY status;

-- File size distribution
SELECT
    CASE
        WHEN file_size < 1073741824 THEN '<1GB'
        WHEN file_size < 5368709120 THEN '1-5GB'
        WHEN file_size < 10737418240 THEN '5-10GB'
        ELSE '>10GB'
    END as size_range,
    COUNT(*) as count
FROM multipart_uploads
GROUP BY size_range;
```

## Cleanup Maintenance

### Stale Uploads Cleanup
```sql
-- Find uploads older than 24 hours not completed
SELECT * FROM multipart_uploads
WHERE status IN ('initiated', 'uploading')
AND created_at < DATE_SUB(NOW(), INTERVAL 24 HOUR);

-- Abort stale uploads (manual or via scheduled job)
-- Note: Also need to call S3 AbortMultipartUpload API
```

## Client Implementation Example

### JavaScript/TypeScript Client
```typescript
async function uploadLargeFile(file: File, token: string) {
  // 1. Initiate
  const initResponse = await fetch('/api/v1/files/multipart/initiate', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      file_name: file.name,
      file_size: file.size,
      content_type: file.type,
      file_type: file.type.startsWith('video/') ? 'video' : 'image',
    }),
  });
  const { upload_id, file_key, part_size, total_parts } = await initResponse.json();

  // 2. Split file and upload parts
  const parts = [];
  for (let i = 0; i < total_parts; i++) {
    const start = i * part_size;
    const end = Math.min(start + part_size, file.size);
    const chunk = file.slice(start, end);

    // Get presigned URL
    const urlResponse = await fetch('/api/v1/files/multipart/presigned-urls', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        upload_id,
        file_key,
        part_numbers: [i + 1],
      }),
    });
    const { urls } = await urlResponse.json();

    // Upload part
    const uploadResponse = await fetch(urls[0].url, {
      method: 'PUT',
      body: chunk,
    });
    const etag = uploadResponse.headers.get('ETag')?.replace(/"/g, '');

    parts.push({ part_number: i + 1, etag });
  }

  // 3. Complete
  const completeResponse = await fetch('/api/v1/files/multipart/complete', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ upload_id, file_key, parts }),
  });

  return await completeResponse.json();
}
```

## Rollback Plan

If issues are found in production:

1. **Immediate**: Disable routes by commenting out in `routes.go`
2. **Database**: Run down migration to remove table
3. **Code**: Revert commits related to multipart upload
4. **Client**: Fall back to single upload API for all files

## Sign-off Checklist

- [ ] Code reviewed and approved
- [ ] Database migration tested in staging
- [ ] Integration tests passing
- [ ] API documentation updated
- [ ] Client team notified of new endpoints
- [ ] Monitoring dashboards configured
- [ ] Rollback plan documented and tested
- [ ] Production deployment approved

---

**Last Updated**: December 5, 2025
**Status**: Ready for Deployment
