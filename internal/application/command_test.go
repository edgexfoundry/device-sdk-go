// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeToObject(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expected    map[string]any
		expectError bool
	}{
		{
			name:        "nil input",
			input:       nil,
			expected:    nil,
			expectError: false,
		},
		{
			name:        "string input (valid JSON)",
			input:       `{"key1":"abc","key2":123}`,
			expected:    map[string]any{"key1": "abc", "key2": 123},
			expectError: false,
		},
		{
			name:        "string input (invalid JSON)",
			input:       `not a json`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "map[string]any input",
			input:       map[string]any{"key1": "abc", "key2": 123},
			expected:    map[string]any{"key1": "abc", "key2": 123},
			expectError: false,
		},
		{
			name:        "unsupported type",
			input:       12345,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeToObject(tt.input)
			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				expectedJSON, err := json.Marshal(tt.expected)
				require.NoError(t, err)
				resultJSON, err := json.Marshal(result)
				require.NoError(t, err)
				assert.True(t, bytes.Equal(expectedJSON, resultJSON))
			}
		})
	}
}

func TestNormalizeToObjectArray(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expected    []map[string]any
		expectError bool
	}{
		{
			name:        "nil input",
			input:       nil,
			expected:    nil,
			expectError: false,
		},
		{
			name:        "string input (valid JSON)",
			input:       `[{"key1":"abc"},{"key2":123}]`,
			expected:    []map[string]any{{"key1": "abc"}, {"key2": 123}},
			expectError: false,
		},
		{
			name:        "string input (invalid JSON)",
			input:       `not a json`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "[]any input (all elements are map[string]any)",
			input:       []any{map[string]any{"x": 10}, map[string]any{"y": 20}},
			expected:    []map[string]any{{"x": 10}, {"y": 20}},
			expectError: false,
		},
		{
			name:        "[]any input (element is not map[string]any)",
			input:       []any{map[string]any{"x": 10}, "not a map"},
			expected:    nil,
			expectError: true,
		},
		{
			name:        "[]map[string]any input",
			input:       []map[string]any{{"foo": "bar"}, {"baz": 123}},
			expected:    []map[string]any{{"foo": "bar"}, {"baz": 123}},
			expectError: false,
		},
		{
			name:        "unsupported type",
			input:       12345,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeToObjectArray(tt.input)
			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				expectedJSON, err := json.Marshal(tt.expected)
				require.NoError(t, err)
				resultJSON, err := json.Marshal(result)
				require.NoError(t, err)
				assert.True(t, bytes.Equal(expectedJSON, resultJSON))
			}
		})
	}
}
