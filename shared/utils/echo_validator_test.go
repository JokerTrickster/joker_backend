package utils

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
)

type validStruct struct {
	Name  string `validate:"required,min=2"`
	Email string `validate:"required,email"`
}

type invalidStruct struct {
	Name  string `validate:"required,min=5"`
	Email string `validate:"required,email"`
}

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("NewValidator() returned nil")
	}
	if v.validator == nil {
		t.Error("NewValidator() has nil validator")
	}
}

func TestCustomValidator_Validate_Valid(t *testing.T) {
	v := NewValidator()
	input := validStruct{Name: "John Doe", Email: "john@example.com"}

	err := v.Validate(&input)
	if err != nil {
		t.Errorf("Validate() unexpected error for valid struct: %v", err)
	}
}

func TestCustomValidator_Validate_Invalid(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name   string
		input  interface{}
		wantOk bool
	}{
		{
			name:   "missing required field",
			input:  &validStruct{Name: "John", Email: ""},
			wantOk: false,
		},
		{
			name:   "invalid email",
			input:  &validStruct{Name: "John Doe", Email: "not-an-email"},
			wantOk: false,
		},
		{
			name:   "min length violation",
			input:  &validStruct{Name: "J", Email: "john@example.com"},
			wantOk: false,
		},
		{
			name:   "min length on Name field",
			input:  &invalidStruct{Name: "Jo", Email: "valid@example.com"},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.input)
			if (err == nil) == !tt.wantOk {
				t.Errorf("Validate() err = %v, wantErr %v", err, !tt.wantOk)
			}
			if err != nil {
				httpErr, ok := err.(*echo.HTTPError)
				if !ok {
					t.Errorf("Expected *echo.HTTPError, got %T", err)
					return
				}
				if httpErr.Code != http.StatusBadRequest {
					t.Errorf("Expected status 400, got %d", httpErr.Code)
				}
			}
		})
	}
}

