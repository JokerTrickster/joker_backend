# Backend Code Audit Report

**Date:** 2026-03-03
**Scope:** authService, cloudRepositoryService, lottoDefenseService, tdService, molandolanService, shared module
**Total issues:** 37

## Executive Summary

| Severity | Count |
|----------|-------|
| Critical | 7 |
| High | 10 |
| Medium | 12 |
| Low | 8 |

---

## 1. Critical Issues

### 1.1 Double Close Panic -- cloudRepositoryService WebSocket Hub

- **File:** `services/cloudRepositoryService/pkg/websocket/hub.go:102`
- **Description:** `broadcastMessage` closes `client.send` when the channel buffer is full but does not remove the client from the hub's map. When `readPump` exits and triggers `unregisterClient`, it closes the same channel again, causing a panic.
- **Impact:** Server crash on WebSocket broadcast under load.
- **Fix:** Remove the client from the map before closing the channel in `broadcastMessage`, mirroring the tdService hub pattern.

### 1.2 OAuth CSRF -- molandolanService

- **File:** `services/molandolanService/features/auth/handler/oauthHandler.go:88`
- **Description:** OAuth state parameter is hardcoded as `"state"`. The state should be a random per-session token stored in a cookie or session, then verified on callback.
- **Impact:** Cross-site request forgery on OAuth login flow.
- **Fix:** Generate a cryptographically random state, store it in a secure cookie, and verify it in the callback handler.

### 1.3 Race Condition -- authService Signup

- **File:** `services/authService/features/auth/repository/signupAuthRepository.go:67-91`
- **Description:** User creation uses a check-then-create pattern (SELECT then INSERT) without a transaction or database-level unique constraint enforcement. Two concurrent signups with the same email can both pass the check.
- **Impact:** Duplicate user records.
- **Fix:** Wrap in a serializable transaction or rely on the DB unique constraint and handle the duplicate key error.

### 1.4 Race Condition -- authService Google Signin

- **File:** `services/authService/features/auth/repository/googleSigninAuthRepository.go:120-135`
- **Description:** Same check-then-create pattern for Google OAuth find-or-create flow.
- **Impact:** Duplicate user records for the same Google account.
- **Fix:** Same as 1.3 -- use a transaction or handle duplicate key errors.

### 1.5 JWT Signing Algorithm Not Enforced

- **File:** `shared/jwt/jwt.go:141,155,172`
- **Description:** The key functions passed to `jwt.Parse` do not verify that `token.Method` is `*jwt.SigningMethodHMAC`. An attacker could craft a token with `alg: none` or an RSA algorithm to bypass verification.
- **Impact:** Token forgery.
- **Fix:** Add `if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { return nil, fmt.Errorf("unexpected signing method") }` in each key function.

### 1.6 Credential Logging -- Redis Init

- **File:** `shared/db/redis/redis.go:29,36`
- **Description:** `fmt.Println(dbInfos)` and `fmt.Println(connectionString)` print Redis credentials (user, password, host) to stdout.
- **Impact:** Credentials may appear in container logs, monitoring systems, or log aggregation.
- **Fix:** Remove the `fmt.Println` calls or replace with structured logging that redacts sensitive fields.

### 1.7 S3 Content-Disposition Header Injection

- **File:** `shared/aws/s3.go:247`
- **Description:** The `filename` parameter is interpolated directly into the `Content-Disposition` header value (`attachment; filename="%s"`). A user-controlled filename containing `"` or CRLF characters can inject additional headers.
- **Impact:** HTTP response splitting / header injection.
- **Fix:** Sanitize the filename (strip `"`, `\r`, `\n`) or use RFC 5987 `filename*` encoding.

---

## 2. High-Priority Issues

### 2.1 WebSocket CheckOrigin Always True -- tdService

- **File:** `services/tdService/features/td/handler/websocketHandler.go:17-20`
- **Description:** `CheckOrigin` returns `true` for all origins.
- **Impact:** Cross-site WebSocket hijacking.
- **Fix:** Restrict to allowed origins from configuration.

### 2.2 DB Connection Pool Not Configured

- **File:** `shared/db/mysql/mysql.go:60-64`
- **Description:** `sql.DB` is used without calling `SetMaxOpenConns`, `SetMaxIdleConns`, or `SetConnMaxLifetime`.
- **Impact:** Under load, the default unlimited connection pool can exhaust MySQL connections.
- **Fix:** Configure pool limits based on expected load (e.g., `MaxOpenConns=25`, `MaxIdleConns=10`, `ConnMaxLifetime=5m`).

### 2.3 DB Connection Not Closed on Shutdown

- **File:** `shared/init.go:196`
- **Description:** `Cleanup()` does not close the MySQL connection (`MysqlDB.Close()`).
- **Impact:** Connection pool not released on graceful shutdown.
- **Fix:** Add `mysql.MysqlDB.Close()` to `Cleanup()`.

### 2.4 Refresh Token Race -- authService

- **File:** `services/authService/features/auth/usecase/refreshTokenUseCase.go:38-49`
- **Description:** Old tokens are deleted before new ones are created. If `CreateToken` fails, the user has no valid tokens and cannot authenticate.
- **Impact:** User locked out on token refresh failure.
- **Fix:** Use a transaction: create new tokens first, then delete old ones, or wrap both in a single transaction.

### 2.5 Uncapped PageSize -- cloudRepositoryService

- **File:** `services/cloudRepositoryService/features/cloudRepository/repository/listCloudRepository.go:96-98`
- **Description:** `PageSize` is clamped for `< 1` (default 20) but has no upper limit.
- **Impact:** A request with `pageSize=1000000` can cause high memory usage and DB load.
- **Fix:** Cap `PageSize` (e.g., max 100).

### 2.6 Direct DB Access in Handlers -- molandolanService

- **Files:** `services/molandolanService/features/gallery/handler/deleteHandler.go:33-34`, `commentDeleteHandler.go:33-34`
- **Description:** Handlers query `mysql.GormMysqlDB` directly with raw table names (`morandoran_users`) instead of going through the repository layer.
- **Impact:** Breaks layered architecture, makes testing difficult, bypasses any repository-level logic.
- **Fix:** Move these queries to the repository and call them from the usecase.

### 2.7 Ignored UpdateColumn Results -- molandolanService Gallery

- **File:** `services/molandolanService/features/gallery/repository/galleryRepository.go:83-84,89-90,131-132,147-148`
- **Description:** `UpdateColumn("like_count", ...)` and `UpdateColumn("comment_count", ...)` results are not checked for errors.
- **Impact:** Counts can silently become inconsistent if the update fails.
- **Fix:** Check and return the error from `UpdateColumn`.

### 2.8 Fragile Error Matching via String Comparison

- **Files:** `services/molandolanService/features/gallery/handler/commentDeleteHandler.go:37-38` (`err.Error() == "FORBIDDEN"`), `services/lottoDefenseService/features/lottoDefense/handler/round_handler.go:57-58,97-98`
- **Description:** Errors are compared by their string message rather than using sentinel errors or custom types.
- **Impact:** A change in the error message silently breaks the comparison.
- **Fix:** Define sentinel errors (`var ErrForbidden = errors.New("forbidden")`) and use `errors.Is()`.

### 2.9 Hardcoded getPlayerNumber -- tdService

- **File:** `services/tdService/features/td/handler/websocketHandler.go:92`
- **Description:** `getPlayerNumber` always returns `1`.
- **Impact:** Co-op player ordering is incorrect; all players see themselves as player 1.
- **Fix:** Derive from join order or database state.

### 2.10 Email Loss When SES Channel Full

- **File:** `shared/aws/ses.go:127-134`
- **Description:** When the SES request channel is full, the oldest email is dequeued (`<-sesMailReqChan`) and the new one is enqueued, silently dropping the old email.
- **Impact:** Emails can be lost without any logging or visibility.
- **Fix:** Log the dropped email, or block/retry instead of dropping.

---

## 3. Medium-Priority Issues

### 3.1 Missing Input Validation -- molandolanService

- **Files:** `services/molandolanService/features/news/handler/updateHandler.go:32`, `services/molandolanService/features/product/handler/updateHandler.go:32`
- **Description:** After `c.Bind(req)`, `c.Validate(req)` is not called.
- **Impact:** Invalid data can reach the usecase layer.

### 3.2 No Max Cap on `limit` -- molandolanService List Usecases

- **Files:** All list usecases in `features/news/`, `features/product/`, `features/gallery/`, `features/ranking/`
- **Description:** `limit` parameter is not capped; arbitrary large values are accepted.
- **Impact:** High memory usage and slow DB queries.

### 3.3 Internal Errors Exposed to Clients

- **File:** `shared/errors/handler.go:48-49`
- **Description:** `echoErr.Internal.Error()` can include raw internal error messages in API responses.
- **Impact:** Information disclosure (DB errors, stack traces, internal paths).

### 3.4 Missing Database Indexes

- **Files:** `services/molandolanService/features/gallery/model/entity/comment.go:11-12` (GalleryID, AuthorID), `services/molandolanService/features/news/model/entity/news.go:15` (Category)
- **Description:** Columns used in `WHERE` clauses lack indexes.
- **Impact:** Slow queries as data grows.

### 3.5 `context.TODO()` in AWS Calls

- **Files:** `shared/aws/init.go:43`, `shared/aws/ssm.go:11,23`, `shared/aws/ses.go:147`
- **Description:** AWS SDK calls use `context.TODO()` instead of a real context with timeout.
- **Impact:** Calls cannot be cancelled or timed out; stuck calls block goroutines.

### 3.6 SES Worker Goroutine Never Exits

- **File:** `shared/aws/ses.go:140-174`
- **Description:** The SES worker runs an infinite loop with no shutdown signal.
- **Impact:** Graceful shutdown may hang or be delayed.

### 3.7 Migrate Instance Never Closed

- **File:** `shared/migrate/migrate.go:65-68,119-136,143-154`
- **Description:** `newMigrateInstance` returns a `*migrate.Migrate` that is never closed.
- **Impact:** Possible file descriptor leak.

### 3.8 Goroutine Leak in Timeout Middleware

- **File:** `shared/middleware/middleware.go:129-131`
- **Description:** When the timeout fires, `next(c)` continues running in a goroutine that is never cleaned up.
- **Impact:** Goroutine accumulation under timeouts.

### 3.9 CORS Wildcard Default in Docker Compose

- **File:** `docker-compose.yml:57,84,110`
- **Description:** `CORS_ALLOWED_ORIGINS` defaults to `*` if env var is not set.
- **Impact:** Overly permissive CORS in production if the variable is forgotten.

### 3.10 No Health Checks in Docker Compose

- **File:** `docker-compose.yml:43-67,70-94`
- **Description:** App services lack `healthcheck` definitions.
- **Impact:** Orchestrators cannot detect unhealthy containers.

### 3.11 Negative Offset Allowed -- lottoDefenseService

- **File:** `services/lottoDefenseService/features/towerDefense/handler/game_handler.go:53-54`
- **Description:** `Offset` from query params is not clamped; negative values pass through.
- **Impact:** Unexpected DB behavior.

### 3.12 Download Handler Maps All Errors to 404

- **File:** `services/cloudRepositoryService/features/cloudRepository/handler/downloadCloudHandler.go:49-52`
- **Description:** All usecase errors return HTTP 404, including internal server errors.
- **Impact:** Internal errors are misreported; clients cannot distinguish "not found" from "server error".

---

## 4. Low-Priority / Improvements

### 4.1 Hardcoded Configuration Values -- molandolanService

- **Files:** `cmd/main.go:60` (timeout), `cmd/main.go:169` (port `18083`), `features/auth/handler/oauthHandler.go:29` (frontendURL), `features/upload/usecase/uploadUseCase.go:16,50,77` (maxFileSize, bucket, region)
- **Recommendation:** Move to environment variables or config.

### 4.2 Deprecated `rand.Seed`

- **File:** `shared/aws/s3.go:186-187`
- **Description:** `rand.Seed(time.Now().UnixNano())` is deprecated in Go 1.20+.
- **Recommendation:** Use `rand.New(rand.NewSource(seed))` or remove (Go 1.20+ auto-seeds).

### 4.3 Ignored `time.Parse` Error

- **File:** `shared/db/mysql/mysql.go:96-99`
- **Description:** `date, _ := time.Parse(...)` discards the parse error; returns `0` for invalid input.
- **Recommendation:** Return the error.

### 4.4 Table Name vs Service Name Mismatch

- **File:** `services/molandolanService/features/auth/model/entity/user.go:23`
- **Description:** `TableName()` returns `"morandoran_users"` but the service is named `molandolan`.
- **Recommendation:** Rename via migration when ready, or document the intentional mismatch.

### 4.5 SES Retry Without Backoff

- **File:** `shared/aws/ses.go:164-168`
- **Description:** Failed emails are immediately requeued (up to 3 times) without any delay.
- **Recommendation:** Add exponential backoff between retries.

### 4.6 Default Credentials in tdService

- **File:** `services/tdService/cmd/main.go:24-27`
- **Description:** Defaults like `rootpassword` and `test-secret-key` if env vars are not set.
- **Recommendation:** Fail startup when critical secrets are missing in production mode.

### 4.7 LIKE Pattern Abuse in Search

- **Files:** `services/cloudRepositoryService/features/cloudRepository/repository/listCloudRepository.go:56`, `favoriteRepository.go:58,65,70`
- **Description:** User input in `LIKE` patterns can include `%` and `_` wildcards.
- **Recommendation:** Escape special LIKE characters in user input.

### 4.8 Missing `docker-compose.test.yml`

- **File:** `Makefile:91,100`
- **Description:** `make docker-up` and `make docker-down` reference `docker-compose.test.yml` which does not exist.
- **Recommendation:** Create the file or update the Makefile targets.

---

## 5. Verified Safe Areas

The following areas were reviewed and found to be correctly implemented:

- **CORS middleware** (`shared/middleware/cors.go`): Production mode forbids wildcard and enforces explicit origins.
- **Body size limit** (`shared/middleware/bodysizelimit.go`): DoS protection enabled.
- **Rate limiting** (`shared/middleware/ratelimit.go`): IP-based rate limiting with cleanup goroutine.
- **Password hashing** (authService): bcrypt with `DefaultCost`; passwords never stored in plaintext.
- **SQL injection**: All services use GORM parameterized queries; no raw string concatenation in SQL.
- **Request ID** (`shared/utils/requestid.go`): Uses `crypto/rand`.
- **S3 file handles**: Properly closed with `defer src.Close()`.
- **Recovery middleware**: Panics are recovered and logged.
- **Migrations**: All 13 migrations (000001-000013) have matching up/down files.
