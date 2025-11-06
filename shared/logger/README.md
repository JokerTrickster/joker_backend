# Logger Package

Structured logging implementation using Uber's Zap for high-performance logging.

## Features

- Structured JSON logging
- Configurable log levels (debug, info, warn, error)
- ISO8601 timestamps
- Caller information (file:line)
- Multiple output streams (stdout/stderr)
- Zero-allocation logging in production
- Development-friendly output

## Configuration

Initialize the logger with a log level:

```go
logger.Init("info") // Options: debug, info, warn, error
defer logger.Sync() // Flush buffered logs
```

The logger automatically detects the environment:
- Development mode: When `ENV != "production"`
- Production mode: When `ENV == "production"`

## Usage

### Basic Logging

```go
// Simple messages
logger.Debug("Debug message")
logger.Info("Server started")
logger.Warn("Deprecated API used")
logger.Error("Failed to connect")
logger.Fatal("Critical error") // Logs and exits
```

### Structured Logging with Fields

```go
import "go.uber.org/zap"

// Log with structured fields
logger.Info("User created",
    zap.String("user_id", "123"),
    zap.String("email", "user@example.com"),
    zap.Int("age", 25),
)

logger.Error("Database error",
    zap.String("operation", "SELECT"),
    zap.String("table", "users"),
    zap.Error(err),
    zap.Duration("elapsed", time.Since(start)),
)
```

### Common Field Types

```go
zap.String("key", "value")
zap.Int("count", 42)
zap.Int64("id", 1234567890)
zap.Float64("price", 19.99)
zap.Bool("active", true)
zap.Error(err)
zap.Duration("elapsed", duration)
zap.Time("timestamp", time.Now())
zap.Any("data", complexObject) // Use sparingly
```

### Child Loggers with Persistent Fields

```go
// Create a child logger with persistent fields
userLogger := logger.With(
    zap.String("user_id", "123"),
    zap.String("session_id", "abc"),
)

// All logs from userLogger will include these fields
userLogger.Info("Login successful")
userLogger.Warn("Password changed")
```

### Advanced Usage

```go
// Get the underlying zap logger for advanced features
zapLogger := logger.GetLogger()

// Sync buffered logs (important before program exit)
logger.Sync()
```

## Log Output Format

### JSON Format (Production)

```json
{
  "level": "info",
  "timestamp": "2025-11-04T16:30:00.000+0900",
  "caller": "handler/user_handler.go:35",
  "message": "User created",
  "user_id": "123",
  "email": "user@example.com",
  "age": 25
}
```

### Development Format

More human-readable output in development mode with colored levels and simplified timestamps.

## Log Levels

| Level | When to Use |
|-------|-------------|
| **Debug** | Detailed information for debugging |
| **Info** | General informational messages |
| **Warn** | Warning messages for potentially harmful situations |
| **Error** | Error messages for failures that don't stop the application |
| **Fatal** | Critical errors that require application shutdown |

## Best Practices

### DO ✅

```go
// Use structured fields
logger.Info("Order processed",
    zap.String("order_id", orderID),
    zap.Float64("amount", amount),
    zap.Duration("processing_time", elapsed),
)

// Log errors with context
logger.Error("Payment failed",
    zap.String("payment_id", paymentID),
    zap.String("gateway", "stripe"),
    zap.Error(err),
)
```

### DON'T ❌

```go
// Don't use string concatenation
logger.Info("Order " + orderID + " processed") // ❌

// Don't log sensitive data
logger.Info("Login",
    zap.String("password", password), // ❌
)

// Don't use fmt.Sprintf unnecessarily
logger.Info(fmt.Sprintf("Count: %d", count)) // ❌
// Use: logger.Info("Count", zap.Int("count", count))
```

## Performance

Zap is designed for high performance:
- Zero allocation in production
- Structured logging faster than fmt.Printf
- Efficient JSON encoding
- Minimal overhead for disabled log levels

Benchmark comparison:
```
Zap:         3-10x faster than other loggers
Zero alloc:  No garbage collection pressure
Structured:  Better than string concatenation
```

## Environment Configuration

Set log level via environment variable:

```bash
# .env file
LOG_LEVEL=debug  # For development
LOG_LEVEL=info   # For production
ENV=production   # Enables production optimizations
```

## Integration with Middleware

The logger integrates seamlessly with custom middleware:

```go
// In middleware
logger.Info("Request completed",
    zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
    zap.String("method", req.Method),
    zap.String("uri", req.RequestURI),
    zap.Int("status", res.Status),
    zap.Duration("latency", duration),
)
```

## Example: Complete Request Logging

```go
func (h *UserHandler) CreateUser(c echo.Context) error {
    start := time.Now()

    var req CreateUserRequest
    if err := c.Bind(&req); err != nil {
        logger.Warn("Invalid request",
            zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
            zap.Error(err),
        )
        return customErrors.BadRequest("Invalid request data")
    }

    user, err := h.service.CreateUser(&req)
    if err != nil {
        logger.Error("Failed to create user",
            zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
            zap.String("email", req.Email),
            zap.Error(err),
            zap.Duration("elapsed", time.Since(start)),
        )
        return customErrors.InternalServerError("Failed to create user")
    }

    logger.Info("User created successfully",
        zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
        zap.String("user_id", user.ID),
        zap.String("email", user.Email),
        zap.Duration("elapsed", time.Since(start)),
    )

    return c.JSON(http.StatusCreated, response.Success(user, "User created"))
}
```
