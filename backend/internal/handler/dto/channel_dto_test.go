package dto

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"gorm.io/datatypes"
)

func TestChannelConfig_Scan(t *testing.T) {
	testCases := []struct {
		name        string
		input       any
		expected    ChannelConfig
		expectError bool
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: nil,
		},
		{
			name:     "[]byte valid JSON",
			input:    []byte(`{"key":"value","num":42}`),
			expected: ChannelConfig{"key": "value", "num": float64(42)},
		},
		{
			name:     "string valid JSON",
			input:    `{"webhook_url":"https://example.com"}`,
			expected: ChannelConfig{"webhook_url": "https://example.com"},
		},
		{
			name:        "invalid type int",
			input:       123,
			expectError: true,
		},
		{
			name:        "invalid type float64",
			input:       3.14,
			expectError: true,
		},
		{
			name:        "invalid JSON bytes",
			input:       []byte(`{invalid json`),
			expectError: true,
		},
		{
			name:        "invalid JSON string",
			input:       `not a json`,
			expectError: true,
		},
		{
			name:     "[]byte empty JSON object",
			input:    []byte(`{}`),
			expected: ChannelConfig{},
		},
		{
			name:     "string empty JSON object",
			input:    `{}`,
			expected: ChannelConfig{},
		},
		{
			name:     "[]byte null JSON",
			input:    []byte(`null`),
			expected: nil,
		},
		{
			name:     "complex nested JSON",
			input:    []byte(`{"a":{"b":1},"c":[1,2,3]}`),
			expected: ChannelConfig{"a": map[string]any{"b": float64(1)}, "c": []any{float64(1), float64(2), float64(3)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var cfg ChannelConfig
			err := cfg.Scan(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(cfg, tc.expected) {
				t.Errorf("got %v, expected %v", cfg, tc.expected)
			}
		})
	}
}

func TestChannelConfig_Value(t *testing.T) {
	testCases := []struct {
		name        string
		input       ChannelConfig
		expected    driver.Value
		expectError bool
	}{
		{
			name:     "nil ChannelConfig",
			input:    nil,
			expected: nil,
		},
		{
			name:        "empty map",
			input:       ChannelConfig{},
			expected:    []byte(`{}`),
			expectError: false,
		},
		{
			name:        "map with nested data",
			input:       ChannelConfig{"headers": map[string]any{"Content-Type": "application/json"}, "method": "POST"},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.input.Value()

			if tc.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tc.input == nil {
				if result != nil {
					t.Errorf("expected nil but got %v", result)
				}
				return
			}

			bytes, ok := result.([]byte)
			if !ok {
				t.Errorf("expected []byte but got %T", result)
				return
			}

			if len(bytes) == 0 {
				t.Errorf("expected non-empty bytes")
				return
			}

			// Compare by re-parsing JSON since map key order is not guaranteed
			if tc.expected != nil {
				var got, want map[string]any
				if err := json.Unmarshal(bytes, &got); err != nil {
					t.Errorf("failed to parse result JSON: %v", err)
					return
				}
				if err := json.Unmarshal(tc.expected.([]byte), &want); err != nil {
					t.Errorf("failed to parse expected JSON: %v", err)
					return
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("got %v, expected %v", got, want)
				}
			}
		})
	}

	// Test normal map data separately to verify JSON content
	t.Run("normal map data", func(t *testing.T) {
		input := ChannelConfig{"bot_token": "abc123", "chat_id": "-1001234567890"}
		result, err := input.Value()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		bytes, ok := result.([]byte)
		if !ok {
			t.Fatalf("expected []byte but got %T", result)
		}

		var parsed map[string]any
		if err := json.Unmarshal(bytes, &parsed); err != nil {
			t.Fatalf("failed to parse result JSON: %v", err)
		}

		if parsed["bot_token"] != "abc123" {
			t.Errorf("bot_token = %v, want abc123", parsed["bot_token"])
		}
		if parsed["chat_id"] != "-1001234567890" {
			t.Errorf("chat_id = %v, want -1001234567890", parsed["chat_id"])
		}
	})
}

func TestChannelConfig_ToJSON(t *testing.T) {
	testCases := []struct {
		name         string
		input        ChannelConfig
		expectedKeys []string
	}{
		{
			name:         "nil ChannelConfig",
			input:        nil,
			expectedKeys: nil, // special case: expect "{}"
		},
		{
			name:         "empty map",
			input:        ChannelConfig{},
			expectedKeys: nil, // special case: expect "{}"
		},
		{
			name:         "normal map data",
			input:        ChannelConfig{"key": "value"},
			expectedKeys: []string{"key"},
		},
		{
			name:         "map with multiple fields",
			input:        ChannelConfig{"smtp_host": "smtp.example.com", "smtp_port": float64(587)},
			expectedKeys: []string{"smtp_host", "smtp_port"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.input.ToJSON()

			if tc.expectedKeys == nil {
				if string(result) != "{}" {
					t.Errorf("got %s, expected {}", result)
				}
				return
			}

			if len(result) == 0 {
				t.Errorf("expected non-empty result")
				return
			}

			resultStr := string(result)
			for _, key := range tc.expectedKeys {
				if !strings.Contains(resultStr, `"`+key+`"`) {
					t.Errorf("result should contain key %q", key)
				}
			}
		})
	}
}

func TestFromJSON(t *testing.T) {
	testCases := []struct {
		name     string
		input    datatypes.JSON
		expected ChannelConfig
	}{
		{
			name:     "nil JSON",
			input:    nil,
			expected: nil,
		},
		{
			name:     "normal JSON data",
			input:    datatypes.JSON(`{"bot_token":"abc123","chat_id":"-1001"}`),
			expected: ChannelConfig{"bot_token": "abc123", "chat_id": "-1001"},
		},
		{
			name:     "empty JSON object",
			input:    datatypes.JSON(`{}`),
			expected: ChannelConfig{},
		},
		{
			name:     "invalid JSON (silent failure returns nil)",
			input:    datatypes.JSON(`{invalid`),
			expected: nil,
		},
		{
			name:     "JSON with nested objects",
			input:    datatypes.JSON(`{"headers":{"Content-Type":"application/json"}}`),
			expected: ChannelConfig{"headers": map[string]any{"Content-Type": "application/json"}},
		},
		{
			name:     "JSON with arrays",
			input:    datatypes.JSON(`{"tags":["a","b","c"]}`),
			expected: ChannelConfig{"tags": []any{"a", "b", "c"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FromJSON(tc.input)

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("got %v, expected %v", result, tc.expected)
			}
		})
	}
}
