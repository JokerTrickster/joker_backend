# Errors Package

Centralized error handling system for the Joker Backend with custom error types and HTTP error handler.

## Features

- Custom `AppError` type with HTTP status codes
- Predefined error constructors for common scenarios
- Error wrapping with context
- Centralized HTTP error handler
- Consistent JSON error responses
- Automatic logging of errors

## AppError Structure

```go
type AppError struct {
    Code       string // Error code (e.g., "BAD_REQUEST")
    Message    string // Human-readable message
    HTTPStatus int    // HTTP status code
    Err        error  // Wrapped error (optional)
}
```

## Common Error Codes

- `BAD_REQUEST` - Invalid request
- `UNAUTHORIZED` - Authentication required
- `FORBIDDEN` - Access denied
- `NOT_FOUND` - Resource not found
- `CONFLICT` - Resource conflict
- `VALIDATION_ERROR` - Validation failed
- `INTERNAL_SERVER_ERROR` - Server error
- `DATABASE_ERROR` - Database operation failed
- `INVALID_INPUT` - Invalid input data
- `RESOURCE_EXISTS` - Resource already exists
- `RESOURCE_NOT_FOUND` - Resource not found

## Usage

### Creating Custom Errors

```go
// Create a new error
err := errors.New("CUSTOM_CODE", "Custom message", http.StatusBadRequest)

// Wrap an existing error
err := errors.Wrap(dbErr, "DATABASE_ERROR", "Failed to query", http.StatusInternalServerError)
```

### Using Predefined Errors

```go
// Bad request
return errors.BadRequest("Invalid email format")

// Unauthorized
return errors.Unauthorized("Invalid credentials")

// Not found
return errors.NotFound("User not found")

// Resource-specific errors
return errors.ResourceNotFound("User")
return errors.ResourceExists("User")

// Validation error
return errors.ValidationError("Email and name are required")

// Database error (wraps original error)
return errors.DatabaseError(err)

// Internal server error
return errors.InternalServerError("Something went wrong")
```

### In Handlers

```go
func (h *UserHandler) GetUser(c echo.Context) error {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        return customErrors.InvalidInput("Invalid user ID format")
    }

    user, err := h.service.GetUserByID(id)
    if err != nil {
        logger.Error("Failed to get user", zap.Int64("user_id", id), zap.Error(err))
        return customErrors.InternalServerError("Failed to retrieve user")
    }

    if user == nil {
        return customErrors.ResourceNotFound("User")
    }

    return c.JSON(http.StatusOK, response.Success(user, "User retrieved successfully"))
}
```

## Custom Error Handler

Set the custom error handler on your Echo instance:

```go
e.HTTPErrorHandler = customErrors.CustomErrorHandler
```

The handler:
- Automatically converts errors to JSON responses
- Logs errors with appropriate levels (error for 5xx, warn for 4xx)
- Includes request ID in logs for tracing
- Handles both custom AppError and Echo's HTTPError
- Returns consistent error response format

## Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "User not found"
  }
}
```

## Error Logging

Errors are automatically logged with structured fields:
- Request ID
- Error code
- Message
- HTTP status
- Original error (if wrapped)

Example log entry:
```json
{
  "level": "error",
  "timestamp": "2025-11-04T16:30:00.000+0900",
  "caller": "errors/handler.go:35",
  "message": "Application error",
  "request_id": "1730704200000",
  "error_code": "DATABASE_ERROR",
  "message": "Failed to query database",
  "error": "connection refused"
}
```

## Testing

Run error tests:

```bash
go test ./pkg/errors -v
```

Run with coverage:

```bash
go test ./pkg/errors -cover
```
