package tools

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

// DefaultEmulatorEndpoint is the default endpoint for the Cosmos DB emulator
const DefaultEmulatorEndpoint = "http://localhost:8081"

// EmulatorKey is the well-known key for the Cosmos DB emulator
const EmulatorKey = "C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw=="

// ConnectionConfig holds connection settings for Azure Cosmos DB.
// It can be embedded in tool input structs to provide consistent connection options.
type ConnectionConfig struct {
	Account          string `json:"account,omitempty" jsonschema:"Azure Cosmos DB account name (required when not using emulator)"`
	UseEmulator      bool   `json:"useEmulator,omitempty" jsonschema:"Set to true to use local Cosmos DB emulator instead of Azure service"`
	EmulatorEndpoint string `json:"emulatorEndpoint,omitempty" jsonschema:"Emulator endpoint URL (default: http://localhost:8081)"`
}

// Validate checks if the connection config is valid
func (c ConnectionConfig) Validate() error {
	if !c.UseEmulator && c.Account == "" {
		return errors.New("account name is required when not using emulator")
	}
	return nil
}

// GetEndpoint returns the appropriate endpoint based on the connection mode
func (c ConnectionConfig) GetEndpoint() string {
	if c.UseEmulator {
		if c.EmulatorEndpoint != "" {
			return c.EmulatorEndpoint
		}
		return DefaultEmulatorEndpoint
	}
	return fmt.Sprintf("https://%s.documents.azure.com:443/", c.Account)
}

// GetClientFunc is a function variable that can be overridden for testing
// It takes a ConnectionConfig and returns a Cosmos DB client
var GetClientFunc func(config ConnectionConfig) (*azcosmos.Client, error)

// GetClient returns a Cosmos DB client based on the connection config
func (c ConnectionConfig) GetClient() (*azcosmos.Client, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	// If a test override is set, use it
	if GetClientFunc != nil {
		return GetClientFunc(c)
	}

	if c.UseEmulator {
		return c.getEmulatorClient()
	}
	return c.getServiceClient()
}

// getServiceClient creates a client for Azure Cosmos DB service using DefaultAzureCredential
func (c ConnectionConfig) getServiceClient() (*azcosmos.Client, error) {
	endpoint := c.GetEndpoint()

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating credential: %v", err)
	}

	client, err := azcosmos.NewClient(endpoint, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating Cosmos client: %v", err)
	}

	return client, nil
}

// getEmulatorClient creates a client for the local Cosmos DB emulator
func (c ConnectionConfig) getEmulatorClient() (*azcosmos.Client, error) {
	endpoint := c.GetEndpoint()

	// Create transport that skips TLS verification (emulator uses self-signed cert)
	transport := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	options := &azcosmos.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Transport: transport,
		},
	}

	// Create credential with the well-known emulator key
	cred, err := azcosmos.NewKeyCredential(EmulatorKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create emulator key credential: %w", err)
	}

	client, err := azcosmos.NewClientWithKey(endpoint, cred, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create emulator client: %w", err)
	}

	return client, nil
}

// GetCosmosClientFunc is a function variable that can be overridden for testing
// Deprecated: Use ConnectionConfig.GetClient() instead
var GetCosmosClientFunc = GetCosmosDBClient

// GetCosmosDBClient creates a Cosmos DB client for Azure service (legacy function)
// Deprecated: Use ConnectionConfig.GetClient() instead
func GetCosmosDBClient(accountName string) (*azcosmos.Client, error) {
	config := ConnectionConfig{Account: accountName, UseEmulator: false}
	return config.getServiceClient()
}
