package mysql

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTimeStringToEpoch(t *testing.T) {
	// 2024-01-15 10:30:45 UTC = 1705314645
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{
			name:    "valid UTC time",
			input:   "2024-01-15 10:30:45 +0000 UTC",
			want:    1705314645,
			wantErr: false,
		},
		{
			name:    "valid with timezone",
			input:   "2024-01-15 02:30:45 -0800 PST",
			want:    1705314645,
			wantErr: false,
		},
		{
			name:    "invalid format - no timezone",
			input:   "2024-01-15 10:30:45",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid format - garbage",
			input:   "not-a-date",
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TimeStringToEpoch(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TimeStringToEpoch(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("TimeStringToEpoch(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestEpochToTime(t *testing.T) {
	// EpochToTime uses time.Unix(t, t%1000*1000000) - nanosecond part from last 3 digits
	epoch := int64(1705311045)
	got := EpochToTime(epoch)
	wantSec := int64(1705311045)
	if got.Unix() != wantSec {
		t.Errorf("EpochToTime(%d).Unix() = %d, want %d", epoch, got.Unix(), wantSec)
	}
}

func TestTimeToEpoch(t *testing.T) {
	utcTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	got := TimeToEpoch(utcTime)
	want := int64(1705314645)
	if got != want {
		t.Errorf("TimeToEpoch(%v) = %d, want %d", utcTime, got, want)
	}
}

func TestPKIDGenerate(t *testing.T) {
	id := PKIDGenerate()
	if id == "" {
		t.Error("PKIDGenerate() returned empty string")
	}
	_, err := uuid.Parse(id)
	if err != nil {
		t.Errorf("PKIDGenerate() returned invalid UUID %q: %v", id, err)
	}
}

func TestPKIDGenerate_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := PKIDGenerate()
		if seen[id] {
			t.Errorf("PKIDGenerate() produced duplicate: %s", id)
		}
		seen[id] = true
	}
}

func TestNowDateGenerate(t *testing.T) {
	got := NowDateGenerate()
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", got, time.Local)
	if err != nil {
		t.Errorf("NowDateGenerate() returned invalid format %q: %v", got, err)
		return
	}

	now := time.Now()
	diff := now.Sub(parsed)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("NowDateGenerate() = %q (parsed %v) not within 1s of now %v", got, parsed, now)
	}
}
