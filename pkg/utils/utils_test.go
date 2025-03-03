package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDeviceName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:     "Valid device name",
			input:    "sde",
			expected: nil,
		},
		{
			name:     "Invalid device name",
			input:    "I",
			expected: errors.New("error invalid device name"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ValidateDeviceName(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestValidateCommandInput(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:     "Valid command input",
			input:    "ls",
			expected: nil,
		},
		{
			name:     "Invalid command input",
			input:    "ls | grep",
			expected: errors.New("error invalid command input"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ValidateCommandInput(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
