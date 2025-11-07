package utils

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name:    "email with only whitespace",
			email:   "   ",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "invalid email no @",
			email:   "notanemail",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "invalid email multiple @",
			email:   "user@@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "invalid email no domain",
			email:   "user@",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "invalid email no local part",
			email:   "@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "too long email",
			email:   strings.Repeat("a", 310) + "@example.com", // 321+ characters
			wantErr: true,
			errMsg:  "email too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateEmail() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid name",
			input:   "John Doe",
			wantErr: false,
		},
		{
			name:    "valid name with unicode",
			input:   "김철수",
			wantErr: false,
		},
		{
			name:    "valid name at min length",
			input:   "Jo",
			wantErr: false,
		},
		{
			name:    "valid name at max length",
			input:   strings.Repeat("a", 255),
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name:    "name with only whitespace",
			input:   "   ",
			wantErr: true,
			errMsg:  "name too short",
		},
		{
			name:    "name too short (1 char)",
			input:   "J",
			wantErr: true,
			errMsg:  "name too short",
		},
		{
			name:    "name too long",
			input:   strings.Repeat("a", 256),
			wantErr: true,
			errMsg:  "name too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateName() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateStringLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		min       int
		max       int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid string within range",
			value:     "hello",
			fieldName: "message",
			min:       1,
			max:       10,
			wantErr:   false,
		},
		{
			name:      "string at minimum length",
			value:     "a",
			fieldName: "message",
			min:       1,
			max:       10,
			wantErr:   false,
		},
		{
			name:      "string at maximum length",
			value:     "1234567890",
			fieldName: "message",
			min:       1,
			max:       10,
			wantErr:   false,
		},
		{
			name:      "string too short",
			value:     "ab",
			fieldName: "password",
			min:       8,
			max:       100,
			wantErr:   true,
			errMsg:    "password too short (min 8 characters)",
		},
		{
			name:      "string too long",
			value:     "12345678901",
			fieldName: "username",
			min:       1,
			max:       10,
			wantErr:   true,
			errMsg:    "username too long (max 10 characters)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringLength(tt.value, tt.fieldName, tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStringLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidateStringLength() error = %v, want %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "non-empty value",
			value:     "test",
			fieldName: "field",
			wantErr:   false,
		},
		{
			name:      "empty value",
			value:     "",
			fieldName: "field",
			wantErr:   true,
			errMsg:    "field is required",
		},
		{
			name:      "whitespace only",
			value:     "   ",
			fieldName: "username",
			wantErr:   true,
			errMsg:    "username is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequired() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidateRequired() error = %v, want %v", err, tt.errMsg)
			}
		})
	}
}
