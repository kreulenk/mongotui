package cmd

import (
	"testing"
)

func TestGetConnectionString(t *testing.T) {
	tests := []struct {
		name        string
		dbAddress   string
		hostFlag    string
		portFlag    int
		expected    string
		expectError bool
	}{
		{
			name:      "Valid address with default port",
			dbAddress: "localhost",
			expected:  "mongodb://localhost:27017",
		},
		{
			name:      "Valid address with custom port",
			dbAddress: "localhost",
			portFlag:  27018,
			expected:  "mongodb://localhost:27018",
		},
		{
			name:      "Valid address with host flag",
			dbAddress: "localhost",
			hostFlag:  "127.0.0.1",
			expected:  "mongodb://127.0.0.1:27017",
		},
		{
			name:      "Valid address with host and port flag",
			dbAddress: "localhost",
			hostFlag:  "127.0.0.1",
			portFlag:  27018,
			expected:  "mongodb://127.0.0.1:27018",
		},
		{
			name:        "No host provided",
			dbAddress:   "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getConnectionString(tt.dbAddress, tt.hostFlag, tt.portFlag)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}
			if result != tt.expected {
				t.Errorf("expected: %s, got: %s", tt.expected, result)
			}
		})
	}
}
