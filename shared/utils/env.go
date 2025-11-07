package utils

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
)

type CtxValues struct {
	Method    string
	Url       string
	UserID    uint
	StartTime time.Time
	RequestID string
	Email     string
}

func TimeToEpochMillis(time time.Time) int64 {
	nanos := time.UnixNano()
	millis := nanos / 1000000
	return millis
}

func EpochToTime(date int64) time.Time {
	return time.Unix(date, 0)
}

func EpochToTimeMillis(t int64) time.Time {
	return time.Unix(t/1000, t%1000*1000000)
}

func CtxGenerate(c echo.Context) (context.Context, uint, string) {
	userID, _ := c.Get("uID").(uint)
	requestID, _ := c.Get("rID").(string)
	startTime, _ := c.Get("startTime").(time.Time)
	email, _ := c.Get("email").(string)
	req := c.Request()
	ctx := context.WithValue(req.Context(), "key", &CtxValues{
		Method:    req.Method,
		Url:       req.URL.Path,
		UserID:    userID,
		RequestID: requestID,
		StartTime: startTime,
		Email:     email,
	})
	return ctx, userID, email

}
