---
name: favorites
description: File favorites feature for quick access to frequently used images and videos
status: backlog
created: 2025-11-26T07:47:09Z
---

# PRD: File Favorites Feature

## Executive Summary

Add favorites (bookmarking) functionality to the cloud repository service, enabling users to mark frequently accessed images and videos for quick retrieval. Users can toggle favorites from the file list/detail views and access a dedicated favorites view with filtering and sorting capabilities.

**Value Proposition**: Reduce time to access frequently used files by 80% through instant favorites access, improving user productivity and engagement.

## Problem Statement

### Current Pain Points
Users managing large media libraries (200+ files average) face difficulties:
- **Time-consuming navigation**: Scrolling through entire file lists to find frequently used images/videos
- **No quick access patterns**: No way to mark or prioritize important files
- **Repeated searches**: Users perform the same searches multiple times for commonly accessed files

### Why Now?
- User feedback indicates 60%+ of file access is concentrated on ~20% of files (Pareto principle)
- Frontend team has completed UI/UX implementation, backend implementation is the blocker
- Competitive parity: Standard feature in cloud storage solutions (Google Drive, Dropbox, etc.)

## User Stories

### Primary Persona: Content Creator
**Demographics**: Professional/hobbyist creating content with frequent media reuse
**Goals**: Quick access to commonly used assets, efficient workflow
**Pain Points**: Time wasted finding frequently used files

### User Journey 1: Adding to Favorites
```
AS A user viewing my file list
WHEN I click the star icon on a file card
THEN the file is immediately marked as favorited (optimistic update)
AND the star icon fills/highlights
AND if the request fails, the icon reverts and shows error message
AND I can retry the operation

Acceptance Criteria:
✓ Star icon visible on all file cards in list view
✓ Optimistic update completes within 100ms
✓ Failed requests revert state with user-friendly error
✓ Idempotent: clicking already-favorited file returns 200 OK
✓ Works for both images and videos
```

### User Journey 2: Viewing Favorites List
```
AS A user with favorited files
WHEN I navigate to "Favorites" tab/filter
THEN I see only my favorited files in thumbnail grid/list
AND they are sorted by upload date (newest first) by default
AND I can filter by filename, extension, or tags
AND each item shows: thumbnail, name, size, type, tags, favorited date

Acceptance Criteria:
✓ Pagination support (page, size parameters)
✓ Sort options: upload date, filename
✓ Filter by: filename (q), extension (ext), tags (tag)
✓ Default sort: file upload date DESC
✓ Response includes metadata: total count, current page, page size
```

### User Journey 3: Removing from Favorites
```
AS A user viewing favorites or file list
WHEN I click the filled star icon on a favorited file
THEN the file is immediately unfavorited (optimistic update)
AND the star icon becomes outline/unfilled
AND if the request fails, the icon reverts and shows error message

Acceptance Criteria:
✓ Idempotent: DELETE returns 204 even if favorite doesn't exist
✓ Immediate visual feedback (<100ms)
✓ Failed requests show retry option
```

## Requirements

### Functional Requirements

#### FR1: Add to Favorites (POST /api/v1/favorites)
- **Input**: `{ fileId: uint }`
- **Authentication**: Required (JWT)
- **Behavior**:
  - Validate fileId exists in cloud_files table
  - Verify user owns the file (user_id match)
  - Create favorite record with current timestamp
  - If already favorited, return 200 OK (idempotent)
- **Response**:
  - Success: 200/201 with `{ success: true, favoritedAt: timestamp }`
  - File not found: 404 `{ error: "file not found" }`
  - Unauthorized: 403 `{ error: "access denied" }`

#### FR2: Remove from Favorites (DELETE /api/v1/favorites/:fileId)
- **Input**: fileId as URL parameter
- **Authentication**: Required (JWT)
- **Behavior**:
  - Delete favorite record matching (user_id, file_id)
  - Idempotent: return 204 even if record doesn't exist
- **Response**: 204 No Content

#### FR3: List Favorites (GET /api/v1/favorites)
- **Query Parameters**:
  - `page` (int, default: 1)
  - `size` (int, default: 20, max: 100)
  - `sort` (enum: "uploadDate" | "fileName", default: "uploadDate")
  - `order` (enum: "asc" | "desc", default: "desc")
  - `q` (string): filename search
  - `ext` (string): file extension filter (e.g., "jpg", "mp4")
  - `tag` (string): tag filter
- **Response Schema**:
```json
{
  "data": [
    {
      "fileId": 123,
      "fileName": "sunset.jpg",
      "fileSize": 2048576,
      "mimeType": "image/jpeg",
      "thumbnailUrl": "https://...",
      "downloadUrl": "https://...",
      "createdAt": "2025-11-20T10:30:00Z",
      "favoritedAt": "2025-11-25T15:45:00Z",
      "tags": ["nature", "portfolio"]
    }
  ],
  "pagination": {
    "total": 47,
    "page": 1,
    "size": 20,
    "totalPages": 3
  }
}
```

#### FR4: Favorite Status in File List
- Existing file list endpoint should include `isFavorited: boolean` field
- No breaking changes to existing response structure

### Non-Functional Requirements

#### NFR1: Performance
- **Add/Remove latency**: < 200ms p95
- **List query**: < 500ms p95 for 200 favorites
- **Database indexes**: Ensure idx_user_favorited_at and idx_user_fileid perform well
- **Optimistic updates**: Frontend can update UI before server response

#### NFR2: Scalability
- Support unlimited favorites per user
- Design for 200 favorites average, 1000+ favorites edge case
- Pagination required for list endpoint (max 100 per page)

#### NFR3: Data Integrity
- **Unique constraint**: (user_id, file_id) prevents duplicate favorites
- **Cascade deletion**: When file deleted, favorites auto-deleted (ON DELETE CASCADE)
- **Referential integrity**: Foreign keys enforce valid user_id and file_id

#### NFR4: Security
- All endpoints require JWT authentication
- Users can only favorite their own files (ownership check)
- Rate limiting: 100 requests/minute per user (prevent abuse)

#### NFR5: Reliability
- Idempotent operations (safe to retry)
- Graceful degradation: if favorites service fails, file list still works
- Transaction safety: favorites creation/deletion in DB transactions

## Success Criteria

### Quantitative Metrics
1. **Adoption Rate**: 40%+ of active users create at least one favorite within 30 days
2. **Engagement**: Users with favorites have 25%+ higher session duration
3. **Performance**: p95 latency < 200ms for add/remove, < 500ms for list
4. **Reliability**: 99.9% uptime, < 0.1% error rate

### Qualitative Metrics
1. **User Feedback**: Positive sentiment in user surveys (>4/5 rating)
2. **Workflow Improvement**: Reduced clicks to access frequently used files
3. **Feature Discovery**: 60%+ of users discover favorites within first week

### Launch Criteria
- [ ] All API endpoints functional and tested
- [ ] Database migration completed without data loss
- [ ] Performance benchmarks met (load testing)
- [ ] Integration tests passing (>95% coverage)
- [ ] Frontend integration complete and verified
- [ ] Documentation updated (API docs, user guide)

## Constraints & Assumptions

### Technical Constraints
- **Database**: MySQL (existing), must use provided schema
- **Architecture**: Follow existing cloudRepositoryService patterns (handler → usecase → repository)
- **Authentication**: Existing JWT system, no changes allowed
- **API Versioning**: Use /api/v1 prefix (existing pattern)

### Business Constraints
- **Timeline**: Backend implementation must complete before frontend launch
- **Resources**: Single backend developer, 1-2 week timeline
- **Breaking Changes**: No breaking changes to existing APIs allowed

### Assumptions
1. **User Behavior**: Users will favorite 10-20% of their total files
2. **Access Patterns**: Favorites list accessed less frequently than main file list
3. **Data Retention**: Favorites persist indefinitely (no auto-cleanup)
4. **Frontend Ready**: Frontend implementation already complete and tested
5. **Database Schema**: `users` and `cloud_files` tables exist with correct structure

## Out of Scope

### Explicitly NOT Included
1. **Shared Favorites**: No collaboration/sharing of favorite lists
2. **Folders/Collections**: No nested organization beyond flat favorites list
3. **Smart Favorites**: No auto-favorite based on usage patterns
4. **Favorite Limits**: No quota/limit enforcement (unlimited)
5. **Bulk Operations**: No bulk add/remove (one file at a time)
6. **Analytics Dashboard**: No favorites usage analytics for users
7. **Export/Import**: No favorites backup/restore functionality
8. **Mobile Optimization**: Focus on web, mobile optimization is Phase 2
9. **Offline Support**: No offline favorites access
10. **Notification**: No notifications when favorited files are modified

### Future Considerations (Phase 2)
- Favorites folders/collections for better organization
- Bulk add/remove operations
- Analytics: most favorited files, favorites trends
- Smart suggestions based on favorites patterns

## Dependencies

### External Dependencies
1. **Database**: MySQL 8.0+ with InnoDB engine
2. **AWS S3**: For presigned URLs (thumbnail/download)
3. **JWT Service**: Existing authentication system
4. **Frontend**: React application (already implemented)

### Internal Dependencies
1. **cloudRepositoryService**: Base service for file operations
2. **Shared AWS Module**: For S3 presigned URL generation
3. **Migration System**: golang-migrate for DB schema changes

### Team Dependencies
1. **Frontend Team**: Integration testing and deployment coordination
2. **DevOps**: Database migration execution in production
3. **QA**: End-to-end testing and performance validation

## Technical Specification

### Database Schema

```sql
-- New table: favorites
CREATE TABLE favorites (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  file_id BIGINT UNSIGNED NOT NULL,
  favorited_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  -- Constraints
  UNIQUE KEY uniq_user_file (user_id, file_id),
  INDEX idx_user_favorited_at (user_id, favorited_at),
  INDEX idx_user_fileid (user_id, file_id),

  -- Foreign keys with cascade delete
  CONSTRAINT fk_fav_user FOREIGN KEY (user_id)
    REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_fav_file FOREIGN KEY (file_id)
    REFERENCES cloud_files(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### API Endpoints Summary

| Method | Endpoint | Purpose | Auth |
|--------|----------|---------|------|
| POST | /api/v1/favorites | Add to favorites | Required |
| DELETE | /api/v1/favorites/:fileId | Remove from favorites | Required |
| GET | /api/v1/favorites | List all favorites | Required |

### Architecture Pattern

Follow existing cloudRepositoryService structure:
```
handler/favoriteHandler.go
  ↓
usecase/favoriteUseCase.go
  ↓
repository/favoriteRepository.go
  ↓
model/
  - entity/favorite.go
  - request/favoriteRequest.go
  - response/favoriteResponse.go
  - interface/IFavoriteRepository.go
```

### Error Handling Strategy

| Scenario | HTTP Code | Response |
|----------|-----------|----------|
| Success (add) | 200/201 | `{ success: true, favoritedAt: "..." }` |
| Already favorited | 200 | `{ success: true, favoritedAt: "..." }` |
| File not found | 404 | `{ error: "file not found" }` |
| Unauthorized file | 403 | `{ error: "access denied" }` |
| Invalid input | 400 | `{ error: "invalid file ID" }` |
| Success (remove) | 204 | No content |
| Remove non-existent | 204 | No content (idempotent) |

## Implementation Checklist

### Phase 1: Database Setup
- [ ] Create migration: `000004_create_favorites_table.up.sql`
- [ ] Create rollback: `000004_create_favorites_table.down.sql`
- [ ] Test migration locally
- [ ] Verify indexes created correctly

### Phase 2: Backend Implementation
- [ ] Entity: `model/entity/favorite.go`
- [ ] Requests: `model/request/favoriteRequest.go`
- [ ] Responses: `model/response/favoriteResponse.go`
- [ ] Interface: `model/interface/IFavoriteRepository.go`
- [ ] Repository: `repository/favoriteRepository.go`
- [ ] UseCase: `usecase/favoriteUseCase.go`
- [ ] Handler: `handler/favoriteHandler.go`
- [ ] Routes: Update `handler/routes.go`

### Phase 3: Testing
- [ ] Unit tests: repository layer (>90% coverage)
- [ ] Unit tests: usecase layer (>90% coverage)
- [ ] Integration tests: API endpoints
- [ ] Performance tests: 200 favorites load test
- [ ] Edge cases: concurrent add/remove, duplicate handling

### Phase 4: Documentation
- [ ] API documentation (Swagger/OpenAPI)
- [ ] Code comments and examples
- [ ] Migration guide for production

### Phase 5: Deployment
- [ ] Code review and approval
- [ ] Run migration in staging
- [ ] Integration testing with frontend
- [ ] Production deployment
- [ ] Monitor error rates and performance

---

## Appendix

### API Request/Response Examples

#### Add to Favorites
```bash
POST /api/v1/favorites
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "fileId": 123
}

# Response 201 Created
{
  "success": true,
  "favoritedAt": "2025-11-26T07:45:00Z"
}
```

#### Remove from Favorites
```bash
DELETE /api/v1/favorites/123
Authorization: Bearer <jwt_token>

# Response 204 No Content
```

#### List Favorites
```bash
GET /api/v1/favorites?page=1&size=20&sort=uploadDate&order=desc&tag=nature
Authorization: Bearer <jwt_token>

# Response 200 OK
{
  "data": [...],
  "pagination": {
    "total": 47,
    "page": 1,
    "size": 20,
    "totalPages": 3
  }
}
```

### Performance Benchmarks

Expected query times (200 favorites):
- Add favorite: < 50ms (INSERT with indexes)
- Remove favorite: < 30ms (DELETE with indexes)
- List favorites (no filter): < 100ms (SELECT with pagination)
- List favorites (with filters): < 200ms (SELECT with WHERE + indexes)

### Security Considerations

1. **SQL Injection**: Use parameterized queries (GORM provides this)
2. **Authorization**: Verify user_id from JWT matches file ownership
3. **Rate Limiting**: Prevent abuse with request throttling
4. **Input Validation**: Sanitize fileId input (uint type check)
