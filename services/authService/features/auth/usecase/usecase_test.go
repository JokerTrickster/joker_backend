package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTokenDTO_CorrectFields(t *testing.T) {
	userID := uint(42)
	accessToken := "access-token-abc123"
	refreshToken := "refresh-token-xyz789"

	dto := createTokenDTO(userID, accessToken, refreshToken)

	require.NotNil(t, dto, "createTokenDTO should not return nil")
	assert.Equal(t, userID, dto.UserID, "UserID should be set correctly")
	assert.Equal(t, accessToken, dto.AccessToken, "AccessToken should be set correctly")
	assert.Equal(t, refreshToken, dto.RefreshToken, "RefreshToken should be set correctly")
	assert.NotZero(t, dto.RefreshExpiredAt, "RefreshExpiredAt should be set")

	t.Logf("createTokenDTO: UserID=%d, AccessToken=%s, RefreshToken=%s, RefreshExpiredAt=%d",
		dto.UserID, dto.AccessToken, dto.RefreshToken, dto.RefreshExpiredAt)
}

func TestCreateTokenDTO_RefreshExpiredAtApproxSevenDays(t *testing.T) {
	before := time.Now()
	dto := createTokenDTO(1, "a", "b")
	after := time.Now()

	expiredAt := time.Unix(dto.RefreshExpiredAt, 0)
	expectedMin := before.Add(7*24*time.Hour - 5*time.Second)
	expectedMax := after.Add(7*24*time.Hour + 5*time.Second)

	assert.True(t, expiredAt.After(expectedMin), "RefreshExpiredAt should be at least ~7 days from call start")
	assert.True(t, expiredAt.Before(expectedMax), "RefreshExpiredAt should be at most ~7 days from call end")

	diff := expiredAt.Sub(before)
	assert.InDelta(t, 7*24*time.Hour, diff, float64(2*time.Second), "RefreshExpiredAt should be approximately 7 days from now")

	t.Logf("RefreshExpiredAt=%d (unix), ~%v from now", dto.RefreshExpiredAt, diff)
}

func TestCreateTokenDTO_MultipleCallsProduceDifferentTimestamps(t *testing.T) {
	dto1 := createTokenDTO(1, "a", "b")
	time.Sleep(1100 * time.Millisecond) // Unix() is in seconds; need >1s to get different timestamp
	dto2 := createTokenDTO(1, "a", "b")

	assert.NotEqual(t, dto1.RefreshExpiredAt, dto2.RefreshExpiredAt,
		"Multiple calls should produce different RefreshExpiredAt timestamps")
	assert.True(t, dto2.RefreshExpiredAt > dto1.RefreshExpiredAt,
		"Later call should have later RefreshExpiredAt")

	t.Logf("dto1.RefreshExpiredAt=%d, dto2.RefreshExpiredAt=%d", dto1.RefreshExpiredAt, dto2.RefreshExpiredAt)
}
