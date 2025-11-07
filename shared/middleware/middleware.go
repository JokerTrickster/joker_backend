package middleware

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/labstack/echo/v4"
	"main/shared/logger"
	"go.uber.org/zap"
)

// RequestLogger logs incoming requests with detailed information
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			res := c.Response()

			// Generate request ID if not exists
			reqID := c.Request().Header.Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = fmt.Sprintf("%d", time.Now().UnixNano())
				c.Request().Header.Set(echo.HeaderXRequestID, reqID)
			}
			c.Response().Header().Set(echo.HeaderXRequestID, reqID)

			// Process request
			err := next(c)

			// Calculate duration
			duration := time.Since(start)

			// Log request details
			fields := []zap.Field{
				zap.String("request_id", reqID),
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.String("remote_ip", c.RealIP()),
				zap.Int("status", res.Status),
				zap.Int64("bytes_out", res.Size),
				zap.Duration("latency", duration),
				zap.String("user_agent", req.UserAgent()),
			}

			if err != nil {
				fields = append(fields, zap.Error(err))
				logger.Error("Request failed", fields...)
			} else {
				// Log different levels based on status code
				status := res.Status
				if status >= 500 {
					logger.Error("Server error", fields...)
				} else if status >= 400 {
					logger.Warn("Client error", fields...)
				} else {
					logger.Info("Request completed", fields...)
				}
			}

			return err
		}
	}
}

// Recovery recovers from panics and logs them
func Recovery() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					stack := string(debug.Stack())

					logger.Error("Panic recovered",
						zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
						zap.String("method", c.Request().Method),
						zap.String("uri", c.Request().RequestURI),
						zap.Error(err),
						zap.String("stack", stack),
					)

					c.Error(echo.NewHTTPError(500, "Internal server error"))
				}
			}()
			return next(c)
		}
	}
}

// RequestID adds a unique request ID to each request
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Request().Header.Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = fmt.Sprintf("%d", time.Now().UnixNano())
			}
			c.Request().Header.Set(echo.HeaderXRequestID, reqID)
			c.Response().Header().Set(echo.HeaderXRequestID, reqID)
			return next(c)
		}
	}
}

// Timeout sets a timeout for request processing
func Timeout(duration time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			timeoutCtx, cancel := context.WithTimeout(ctx, duration)
			defer cancel()

			c.SetRequest(c.Request().WithContext(timeoutCtx))

			done := make(chan error, 1)
			go func() {
				done <- next(c)
			}()

			select {
			case err := <-done:
				return err
			case <-timeoutCtx.Done():
				logger.Warn("Request timeout",
					zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
					zap.String("method", c.Request().Method),
					zap.String("uri", c.Request().RequestURI),
					zap.Duration("timeout", duration),
				)
				return echo.NewHTTPError(408, "Request timeout")
			}
		}
	}
}
