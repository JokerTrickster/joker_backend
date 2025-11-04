# Middleware Package

Custom middleware implementations for the Joker Backend Echo server.

## Available Middleware

### RequestLogger
Logs all incoming HTTP requests with detailed information including:
- Request ID
- HTTP method and URI
- Remote IP address
- Response status code
- Response size
- Request latency
- User agent

```go
e.Use(customMiddleware.RequestLogger())
```

### Recovery
Recovers from panics in handlers and logs them with stack traces.

```go
e.Use(customMiddleware.Recovery())
```

### RequestID
Adds a unique request ID to each request for tracing purposes. The request ID is:
- Generated automatically if not provided
- Propagated in response headers
- Available via `X-Request-ID` header

```go
e.Use(customMiddleware.RequestID())
```

### CORS
Configures Cross-Origin Resource Sharing (CORS) with sensible defaults:
- Allows all origins (configure based on environment)
- Supports common HTTP methods
- Includes authorization headers
- 24-hour max age

```go
e.Use(customMiddleware.CORS())
```

### RateLimiter
IP-based rate limiting to protect against abuse:

```go
// 10 requests per second, burst of 20
rateLimiter := customMiddleware.NewRateLimiter(10, 20)
e.Use(rateLimiter.Middleware())
```

Features:
- Per-IP rate limiting
- Configurable rate and burst size
- Automatic cleanup of old visitors
- Responds with HTTP 429 when limit exceeded

### Timeout
Sets a timeout for request processing:

```go
// 30-second timeout
e.Use(customMiddleware.Timeout(30 * time.Second))
```

Returns HTTP 408 when timeout is exceeded.

## Middleware Order

The order of middleware matters! Recommended order:

```go
e.Use(customMiddleware.RequestID())     // 1. Generate request ID
e.Use(customMiddleware.Recovery())      // 2. Catch panics
e.Use(customMiddleware.RequestLogger()) // 3. Log requests
e.Use(customMiddleware.CORS())          // 4. CORS headers
rateLimiter := customMiddleware.NewRateLimiter(10, 20)
e.Use(rateLimiter.Middleware())         // 5. Rate limiting
e.Use(customMiddleware.Timeout(30 * time.Second)) // 6. Request timeout
```

## Testing

Run middleware tests:

```bash
go test ./pkg/middleware -v
```

Run with coverage:

```bash
go test ./pkg/middleware -cover
```
