package response

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSuccessResponse(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		message string
	}{
		{
			name:    "with data and message",
			data:    map[string]string{"id": "123", "name": "test"},
			message: "Operation completed",
		},
		{
			name:    "with nil data",
			data:    nil,
			message: "Success",
		},
		{
			name:    "with slice data",
			data:    []int{1, 2, 3},
			message: "Items retrieved",
		},
		{
			name:    "empty message",
			data:    "result",
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := SuccessResponse(tt.data, tt.message)

			if !res.Success {
				t.Errorf("SuccessResponse().Success = false, want true")
			}
			if res.Message != tt.message {
				t.Errorf("SuccessResponse().Message = %q, want %q", res.Message, tt.message)
			}
			if tt.data != nil && !reflect.DeepEqual(res.Data, tt.data) {
				t.Errorf("SuccessResponse().Data = %v, want %v", res.Data, tt.data)
			}
			if res.Error != nil {
				t.Errorf("SuccessResponse().Error = %v, want nil", res.Error)
			}

			// Marshal to JSON and verify structure
			b, err := json.Marshal(res)
			if err != nil {
				t.Fatalf("json.Marshal(SuccessResponse) = %v", err)
			}
			var parsed map[string]interface{}
			if err := json.Unmarshal(b, &parsed); err != nil {
				t.Fatalf("json.Unmarshal = %v", err)
			}
			if success, ok := parsed["success"].(bool); !ok || !success {
				t.Errorf("JSON success field = %v, want true", parsed["success"])
			}
		})
	}
}

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
		details string
	}{
		{
			name:    "full error response",
			code:    "VALIDATION_ERROR",
			message: "Invalid input",
			details: "Field 'email' is required",
		},
		{
			name:    "error without details",
			code:    "NOT_FOUND",
			message: "Resource not found",
			details: "",
		},
		{
			name:    "internal error",
			code:    "INTERNAL_SERVER_ERROR",
			message: "Something went wrong",
			details: "stack trace here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ErrorResponse(tt.code, tt.message, tt.details)

			if res.Success {
				t.Errorf("ErrorResponse().Success = true, want false")
			}
			if res.Error == nil {
				t.Fatal("ErrorResponse().Error = nil, want non-nil")
			}
			if res.Error.Code != tt.code {
				t.Errorf("ErrorResponse().Error.Code = %q, want %q", res.Error.Code, tt.code)
			}
			if res.Error.Message != tt.message {
				t.Errorf("ErrorResponse().Error.Message = %q, want %q", res.Error.Message, tt.message)
			}
			if res.Error.Details != tt.details {
				t.Errorf("ErrorResponse().Error.Details = %q, want %q", res.Error.Details, tt.details)
			}

			// Marshal to JSON and verify structure
			b, err := json.Marshal(res)
			if err != nil {
				t.Fatalf("json.Marshal(ErrorResponse) = %v", err)
			}
			var parsed map[string]interface{}
			if err := json.Unmarshal(b, &parsed); err != nil {
				t.Fatalf("json.Unmarshal = %v", err)
			}
			if success, ok := parsed["success"].(bool); !ok || success {
				t.Errorf("JSON success field = %v, want false", parsed["success"])
			}
			errObj, ok := parsed["error"].(map[string]interface{})
			if !ok {
				t.Fatal("JSON error field missing or wrong type")
			}
			if errObj["code"] != tt.code {
				t.Errorf("JSON error.code = %v, want %q", errObj["code"], tt.code)
			}
			if errObj["message"] != tt.message {
				t.Errorf("JSON error.message = %v, want %q", errObj["message"], tt.message)
			}
		})
	}
}
