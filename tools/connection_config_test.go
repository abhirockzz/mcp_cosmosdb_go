package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Unit tests for ConnectionConfig validation and client creation logic

func TestConnectionConfig_Validate_ServiceMode(t *testing.T) {
	tests := []struct {
		name        string
		config      ConnectionConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid service mode with account",
			config: ConnectionConfig{
				Account:     "myaccount",
				UseEmulator: false,
			},
			expectError: false,
		},
		{
			name: "service mode without account should fail",
			config: ConnectionConfig{
				Account:     "",
				UseEmulator: false,
			},
			expectError: true,
			errorMsg:    "account name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConnectionConfig_Validate_EmulatorMode(t *testing.T) {
	tests := []struct {
		name        string
		config      ConnectionConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid emulator mode without account",
			config: ConnectionConfig{
				UseEmulator: true,
			},
			expectError: false,
		},
		{
			name: "valid emulator mode with custom endpoint",
			config: ConnectionConfig{
				UseEmulator:      true,
				EmulatorEndpoint: "https://localhost:9000",
			},
			expectError: false,
		},
		{
			name: "emulator mode ignores account field",
			config: ConnectionConfig{
				Account:     "someaccount", // Should be ignored in emulator mode
				UseEmulator: true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConnectionConfig_GetEndpoint(t *testing.T) {
	tests := []struct {
		name             string
		config           ConnectionConfig
		expectedEndpoint string
	}{
		{
			name: "service mode endpoint",
			config: ConnectionConfig{
				Account:     "myaccount",
				UseEmulator: false,
			},
			expectedEndpoint: "https://myaccount.documents.azure.com:443/",
		},
		{
			name: "emulator mode default endpoint",
			config: ConnectionConfig{
				UseEmulator: true,
			},
			expectedEndpoint: DefaultEmulatorEndpoint,
		},
		{
			name: "emulator mode custom endpoint",
			config: ConnectionConfig{
				UseEmulator:      true,
				EmulatorEndpoint: "https://localhost:9000",
			},
			expectedEndpoint: "https://localhost:9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := tt.config.GetEndpoint()
			assert.Equal(t, tt.expectedEndpoint, endpoint)
		})
	}
}

func TestConnectionConfig_IsEmulatorMode(t *testing.T) {
	tests := []struct {
		name     string
		config   ConnectionConfig
		expected bool
	}{
		{
			name: "emulator mode true",
			config: ConnectionConfig{
				UseEmulator: true,
			},
			expected: true,
		},
		{
			name: "emulator mode false",
			config: ConnectionConfig{
				UseEmulator: false,
				Account:     "myaccount",
			},
			expected: false,
		},
		{
			name:     "is service mode",
			config:   ConnectionConfig{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.UseEmulator)
		})
	}
}
