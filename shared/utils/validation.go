package utils

import (
	"fmt"
	"net/mail"
	"strings"
)

// ValidateEmail validates email format and length
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Trim whitespace
	email = strings.TrimSpace(email)

	// Check length (RFC 5321 specifies max 320 characters)
	if len(email) > 320 {
		return fmt.Errorf("email too long (max 320 characters)")
	}

	// Validate format using standard library
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidateName validates name format and length
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}

	// Trim whitespace
	name = strings.TrimSpace(name)

	// Check minimum length
	if len(name) < 2 {
		return fmt.Errorf("name too short (min 2 characters)")
	}

	// Check maximum length
	if len(name) > 255 {
		return fmt.Errorf("name too long (max 255 characters)")
	}

	return nil
}

// ValidateStringLength validates string length with custom limits
func ValidateStringLength(value, fieldName string, min, max int) error {
	value = strings.TrimSpace(value)
	length := len(value)

	if length < min {
		return fmt.Errorf("%s too short (min %d characters)", fieldName, min)
	}

	if length > max {
		return fmt.Errorf("%s too long (max %d characters)", fieldName, max)
	}

	return nil
}

// ValidateRequired checks if a string value is non-empty after trimming
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}
