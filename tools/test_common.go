package tools

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/tracing"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/abhirockzz/cosmosdb-go-sdk-helper/auth"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testOperationDBName        = "testDatabase"
	testOperationContainerName = "testContainer"
	testPartitionKey           = "/id"
	emulatorImage              = "mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:vnext-preview"
	//emulatorImage    = "mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:latest"
	emulatorPort = "8081"
	healthPort   = "8080"

	//emulatorEndpoint = "http://localhost:8081"
	emulatorKey = "C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw=="
)

var emulatorEndpoint string

func setupCosmosEmulator(ctx context.Context) (testcontainers.Container, error) {

	req := testcontainers.ContainerRequest{
		Image:        emulatorImage,
		ExposedPorts: []string{emulatorPort, healthPort},
		WaitingFor:   wait.ForListeningPort(healthPort),
		Env: map[string]string{
			"ENABLE_EXPLORER": "false",
			"PROTOCOL":        "https",
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return container, nil
}

// setupDatabaseAndContainer ensures the test database and container exist
func setupDatabaseAndContainer(ctx context.Context, client *azcosmos.Client) error {
	// Try to create the test database
	databaseProps := azcosmos.DatabaseProperties{ID: testOperationDBName}
	_, err := client.CreateDatabase(ctx, databaseProps, nil)
	if err != nil && !isResourceExistsError(err) {
		return fmt.Errorf("failed to create test database: %w", err)
	}

	// Create container if it doesn't exist
	database, err := client.NewDatabase(testOperationDBName)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	containerProps := azcosmos.ContainerProperties{
		ID: testOperationContainerName,
		PartitionKeyDefinition: azcosmos.PartitionKeyDefinition{
			Paths: []string{testPartitionKey},
		},
		DefaultTimeToLive: to.Ptr[int32](60), // Short TTL for test data (60 seconds)
	}

	_, err = database.CreateContainer(ctx, containerProps, nil)
	if err != nil && !isResourceExistsError(err) {
		return fmt.Errorf("failed to create test container: %w", err)
	}

	return nil
}

// isResourceExistsError checks if error is because resource already exists (status code 409)
func isResourceExistsError(err error) bool {
	var responseErr *azcore.ResponseError
	if errors.As(err, &responseErr) {
		return responseErr.StatusCode == 409
	}
	return false
}

// emulatorTransport is a custom http.RoundTripper that intercepts requests to the Cosmos DB emulator.
// The emulator advertises its internal port (8081) during endpoint discovery, which causes the SDK
// to try connecting to localhost:8081 instead of the mapped Testcontainers port.
// This transport rewrites the destination port to the mapped port to ensure connectivity.
type emulatorTransport struct {
	transport  http.RoundTripper
	mappedPort string
}

func (t *emulatorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Port() == emulatorPort {
		req.URL.Host = fmt.Sprintf("localhost:%s", t.mappedPort)
	}
	return t.transport.RoundTrip(req)
}

func getEmulatorClient(emulator testcontainers.Container) (*azcosmos.Client, error) {
	mappedPort, err := emulator.MappedPort(context.Background(), emulatorPort)
	if err != nil {
		fmt.Printf("Failed to mapped port: %v\n", err)
		os.Exit(1)
	}

	baseTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Wrap the base transport with our custom emulatorTransport to handle port rewriting
	rewritingTransport := &emulatorTransport{
		transport:  baseTransport,
		mappedPort: mappedPort.Port(),
	}

	options := &azcosmos.ClientOptions{ClientOptions: azcore.ClientOptions{
		Transport: &http.Client{Transport: rewritingTransport},
	}}

	emulatorEndpoint = fmt.Sprintf("https://localhost:%s", mappedPort.Port())
	fmt.Printf("Emulator endpoint: %s\n", emulatorEndpoint)

	// Set up the CosmosDB client
	client, err := auth.GetCosmosDBClient(emulatorEndpoint, true, options)

	if err != nil {
		fmt.Printf("Failed to set up CosmosDB client: %v\n", err)
		os.Exit(1)
	}

	return client, nil
}

// deprecated
func _getEmulatorClient() (*azcosmos.Client, error) {

	transport := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	options := &azcosmos.ClientOptions{ClientOptions: azcore.ClientOptions{
		TracingProvider: tracing.Provider{},
		Transport:       transport,
	}}

	// Create credential with the emulator key
	cred, err := azcosmos.NewKeyCredential(emulatorKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create key credential: %w", err)
	}

	// Create the client
	client, err := azcosmos.NewClientWithKey(emulatorEndpoint, cred, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create cosmos client: %w", err)
	}

	return client, nil
}
