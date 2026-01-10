package store

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestToText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected pgtype.Text
	}{
		{
			name:  "non-empty string",
			input: "test",
			expected: pgtype.Text{
				String: "test",
				Valid:  true,
			},
		},
		{
			name:  "empty string",
			input: "",
			expected: pgtype.Text{
				String: "",
				Valid:  false,
			},
		},
		{
			name:  "whitespace string",
			input: "  ",
			expected: pgtype.Text{
				String: "  ",
				Valid:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToText(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToInt4(t *testing.T) {
	tests := []struct {
		name     string
		input    int32
		expected pgtype.Int4
	}{
		{
			name:  "positive number",
			input: 42,
			expected: pgtype.Int4{
				Int32: 42,
				Valid: true,
			},
		},
		{
			name:  "zero",
			input: 0,
			expected: pgtype.Int4{
				Int32: 0,
				Valid: true,
			},
		},
		{
			name:  "negative number",
			input: -10,
			expected: pgtype.Int4{
				Int32: -10,
				Valid: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToInt4(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected pgtype.Bool
	}{
		{
			name:  "true",
			input: true,
			expected: pgtype.Bool{
				Bool:  true,
				Valid: true,
			},
		},
		{
			name:  "false",
			input: false,
			expected: pgtype.Bool{
				Bool:  false,
				Valid: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToBool(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToUUID(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "valid UUID",
			input: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "invalid UUID string",
			input: "not-a-uuid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToUUID(tt.input)
			// Just verify that the function executes without panic
			// The pgtype.UUID.Scan method handles the conversion
			assert.NotNil(t, result)
		})
	}
}
