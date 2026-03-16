package namespace

import (
	"testing"
)

// TestValidate tests the Validate function with various inputs
func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantValid   bool
		wantReason  string
	}{
		// Valid cases
		{
			name:      "valid lowercase with hyphens",
			input:     "my-namespace",
			wantValid: true,
		},
		{
			name:      "valid single word",
			input:     "production",
			wantValid: true,
		},
		{
			name:      "valid with numbers",
			input:     "app-123",
			wantValid: true,
		},
		{
			name:      "valid max length (63 chars)",
			input:     "this-is-exactly-sixty-three-characters-long-namespace-name-x",
			wantValid: true,
		},
		
		// Invalid cases
		{
			name:       "empty string",
			input:      "",
			wantValid:  false,
			wantReason: "namespace name cannot be empty",
		},
		{
			name:       "uppercase letters",
			input:      "My-Namespace",
			wantValid:  false,
			wantReason: "namespace name must be lowercase",
		},
		{
			name:       "starts with hyphen",
			input:      "-invalid",
			wantValid:  false,
			wantReason: "namespace name cannot start with a hyphen",
		},
		{
			name:       "ends with hyphen",
			input:      "invalid-",
			wantValid:  false,
			wantReason: "namespace name cannot end with a hyphen",
		},
		{
			name:       "contains underscore",
			input:      "my_namespace",
			wantValid:  false,
			wantReason: "invalid character '_' at position 2",
		},
		{
			name:       "contains special character",
			input:      "my@namespace",
			wantValid:  false,
			wantReason: "invalid character '@' at position 2",
		},
		{
			name:       "too long (64 chars)",
			input:      "this-namespace-name-is-too-long-and-exceeds-sixty-three-chars-x",
			wantValid:  false,
			wantReason: "namespace name too long (64 characters, max 63)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid, gotReason := Validate(tt.input)
			
			if gotValid != tt.wantValid {
				t.Errorf("Validate(%q) valid = %v, want %v", tt.input, gotValid, tt.wantValid)
			}
			
			// Only check reason for invalid cases
			if !tt.wantValid && gotReason != tt.wantReason {
				t.Errorf("Validate(%q) reason = %q, want %q", tt.input, gotReason, tt.wantReason)
			}
		})
	}
}

// TestIsValidChar tests the isValidChar helper function
func TestIsValidChar(t *testing.T) {
	tests := []struct {
		char rune
		want bool
	}{
		// Valid characters
		{'a', true},
		{'z', true},
		{'m', true},
		{'0', true},
		{'9', true},
		{'5', true},
		{'-', true},
		
		// Invalid characters
		{'A', false},
		{'Z', false},
		{'_', false},
		{'@', false},
		{'!', false},
		{' ', false},
		{'.', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			got := isValidChar(tt.char)
			if got != tt.want {
				t.Errorf("isValidChar(%q) = %v, want %v", tt.char, got, tt.want)
			}
		})
	}
}

// BenchmarkValidate benchmarks the Validate function
func BenchmarkValidate(b *testing.B) {
	testName := "production-api-service-v2"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(testName)
	}
}
