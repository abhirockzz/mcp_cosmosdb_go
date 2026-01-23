package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

var (
	emulator testcontainers.Container
	client   *azcosmos.Client
)

func TestMain(m *testing.M) {
	// Set up the CosmosDB emulator container
	ctx := context.Background()
	var err error
	emulator, err = setupCosmosEmulator(ctx)
	if err != nil {
		fmt.Printf("Failed to set up CosmosDB emulator: %v\n", err)
		os.Exit(1)
	}

	// Set up the CosmosDB client
	client, err = getEmulatorClient(emulator)
	if err != nil {
		fmt.Printf("Failed to set up CosmosDB client: %v\n", err)
		os.Exit(1)
	}

	// Override the client function for tests (new method)
	GetClientFunc = func(config ConnectionConfig) (*azcosmos.Client, error) {
		return client, nil // Always return the emulator client
	}

	// Also override legacy function for backward compatibility
	GetCosmosClientFunc = func(account string) (*azcosmos.Client, error) {
		return client, nil // Always return the emulator client
	}

	// Set up the database and container
	err = setupDatabaseAndContainer(ctx, client)
	if err != nil {
		fmt.Printf("Failed to set up database and container: %v\n", err)
		os.Exit(1)
	}

	// Run the tests
	code := m.Run()

	// Tear down the CosmosDB emulator container
	if emulator != nil {
		_ = emulator.Terminate(ctx)
	}

	os.Exit(code)
}

func TestListDatabases(t *testing.T) {

	tests := []struct {
		name           string
		input          ListDatabasesToolInput
		expectError    bool
		expectedResult string
		expectedErrMsg string
	}{
		{
			name: "valid account name",
			input: ListDatabasesToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
			},
			expectError:    false,
			expectedResult: testOperationDBName,
		},
		{
			name: "empty account name",
			input: ListDatabasesToolInput{
				ConnectionConfig: ConnectionConfig{Account: ""},
			},
			expectError:    true,
			expectedErrMsg: "account name is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			_, response, err := ListDatabasesToolHandler(context.Background(), nil, ListDatabasesToolInput{
				ConnectionConfig: ConnectionConfig{Account: test.input.Account},
			})

			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(response.Databases), 1, "Should have at least one database")
			assert.Contains(t, response.Databases, test.expectedResult, "Should contain the test database")
		})
	}

}

func TestCreateDatabase(t *testing.T) {

	tests := []struct {
		name           string
		input          CreateDatabaseToolInput
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "valid arguments",
			input: CreateDatabaseToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database: "newTestDatabase1",
			},
			expectError: false,
		},
		{
			name: "empty account name",
			input: CreateDatabaseToolInput{
				ConnectionConfig: ConnectionConfig{Account: ""},
				Database: "newTestDatabase",
			},
			expectError:    true,
			expectedErrMsg: "account name is required",
		},
		{
			name: "empty database name",
			input: CreateDatabaseToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database: "",
			},
			expectError:    true,
			expectedErrMsg: "database name missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			_, response, err := CreateDatabaseToolHandler(context.Background(), nil, test.input)

			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "dummy_account_does_not_matter", response.Account)
			assert.Equal(t, test.input.Database, response.Database)
			assert.Contains(t, response.Message, "created successfully")
		})
	}
}

func TestListContainers(t *testing.T) {

	tests := []struct {
		name           string
		input          ListContainersToolInput
		expectError    bool
		expectedResult string
		expectedErrMsg string
	}{
		// to be investigted: in vNext emulator, this returns 400 error with message "id is required in the request body"
		// commenting out temporarily
		// {
		// 	name: "valid arguments",
		// 	input: ListContainersToolInput{
		// 		ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
		// 		Database: testOperationDBName,
		// 	},
		// 	expectError:    false,
		// 	expectedResult: testOperationContainerName,
		// },
		{
			name: "empty account name",
			input: ListContainersToolInput{
				ConnectionConfig: ConnectionConfig{Account: ""},
				Database: testOperationDBName,
			},
			expectError:    true,
			expectedErrMsg: "account name is required",
		},
		{
			name: "empty database name",
			input: ListContainersToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database: "",
			},
			expectError:    true,
			expectedErrMsg: "cosmos db database name missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			_, response, err := ListContainersToolHandler(context.Background(), nil, test.input)

			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "dummy_account_does_not_matter", response.Account)
			assert.Equal(t, testOperationDBName, response.Database)
			assert.Equal(t, 1, len(response.Containers))
			assert.Equal(t, test.expectedResult, response.Containers[0])
		})
	}
}

func TestReadContainerMetadata(t *testing.T) {

	tests := []struct {
		name           string
		input          ReadContainerMetadataToolInput
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "valid arguments",
			input: ReadContainerMetadataToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:  testOperationDBName,
				Container: testOperationContainerName,
			},
			expectError: false,
		},
		{
			name: "empty account name",
			input: ReadContainerMetadataToolInput{
				ConnectionConfig: ConnectionConfig{Account: ""},
				Database:  testOperationDBName,
				Container: testOperationContainerName,
			},
			expectError:    true,
			expectedErrMsg: "account name is required",
		},
		{
			name: "empty database name",
			input: ReadContainerMetadataToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:  "",
				Container: testOperationContainerName,
			},
			expectError:    true,
			expectedErrMsg: "cosmos db database name missing",
		},
		{
			name: "empty container name",
			input: ReadContainerMetadataToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:  testOperationDBName,
				Container: "",
			},
			expectError:    true,
			expectedErrMsg: "container name missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			toolResult, _, err := ReadContainerMetadataToolHandler(context.Background(), nil, test.input)

			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, toolResult)
			require.NotEmpty(t, toolResult.Content)

			textContent, ok := toolResult.Content[0].(*mcp.TextContent)
			require.True(t, ok)

			var metadata map[string]any
			err = json.Unmarshal([]byte(textContent.Text), &metadata)
			require.NoError(t, err)

			assert.Contains(t, metadata, "container_id")
			assert.Equal(t, testOperationContainerName, metadata["container_id"])
			assert.Contains(t, metadata, "indexing_policy")
			assert.Contains(t, metadata, "partition_key_definition")
			assert.Contains(t, metadata, "conflict_resolution_policy")
			assert.Contains(t, metadata, "unique_key_policy")
			assert.Contains(t, metadata, "throughput")

			// Verify throughput structure
			throughput, ok := metadata["throughput"].(map[string]any)
			require.True(t, ok, "throughput should be a map")
			assert.Contains(t, throughput, "type")
			throughputType := throughput["type"].(string)
			// On emulator: "unknown" (400), On real Azure: "manual", "autoscale", or "shared"
			assert.Contains(t, []string{"manual", "autoscale", "shared", "unknown", "error"}, throughputType)
		})
	}
}

func TestReadContainerMetadata_ThroughputScenarios(t *testing.T) {
	ctx := context.Background()

	// Create a container with manual throughput
	manualContainerName := "testContainer_manual_throughput_diag"
	_, _, err := CreateContainerToolHandler(ctx, nil, CreateContainerToolInput{
		ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
		Database:         testOperationDBName,
		Container:        manualContainerName,
		PartitionKeyPath: "/id",
		Throughput:       func() *int32 { v := int32(400); return &v }(),
	})
	require.NoError(t, err, "Failed to create container with manual throughput")

	t.Run("manual throughput container", func(t *testing.T) {
		// Note: vNext emulator returns 400 for /offers endpoint (not implemented)
		// so this test validates the "unknown" throughput type on emulator
		toolResult, _, err := ReadContainerMetadataToolHandler(ctx, nil, ReadContainerMetadataToolInput{
			ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
			Database:  testOperationDBName,
			Container: manualContainerName,
		})

		require.NoError(t, err)
		require.NotNil(t, toolResult)

		textContent, ok := toolResult.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		var metadata map[string]any
		err = json.Unmarshal([]byte(textContent.Text), &metadata)
		require.NoError(t, err)

		throughput, ok := metadata["throughput"].(map[string]any)
		require.True(t, ok, "throughput should be a map")

		// On emulator, we expect "unknown" due to /offers endpoint not being implemented
		// On real Azure, we'd expect "manual" with ru_per_second
		throughputType := throughput["type"].(string)
		assert.Contains(t, []string{"manual", "unknown"}, throughputType, "Should be manual (Azure) or unknown (emulator)")

		if throughputType == "manual" {
			assert.Contains(t, throughput, "ru_per_second", "Should have ru_per_second field")
			ruPerSecond, ok := throughput["ru_per_second"].(float64)
			require.True(t, ok, "ru_per_second should be a number")
			assert.Equal(t, float64(400), ruPerSecond, "Should have 400 RU/s")
		} else {
			assert.Contains(t, throughput, "message", "Should have message explaining limitation")
		}
	})

	t.Run("shared throughput container", func(t *testing.T) {
		// Note: vNext emulator returns 400 for /offers endpoint
		// On real Azure, a container without dedicated throughput returns 404 → "shared"
		toolResult, _, err := ReadContainerMetadataToolHandler(ctx, nil, ReadContainerMetadataToolInput{
			ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
			Database:  testOperationDBName,
			Container: testOperationContainerName,
		})

		require.NoError(t, err)
		require.NotNil(t, toolResult)

		textContent, ok := toolResult.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		var metadata map[string]any
		err = json.Unmarshal([]byte(textContent.Text), &metadata)
		require.NoError(t, err)

		throughput, ok := metadata["throughput"].(map[string]any)
		require.True(t, ok, "throughput should be a map")

		// On emulator: "unknown" (400 error)
		// On real Azure: "shared" (404 error) for container without dedicated throughput
		throughputType := throughput["type"].(string)
		assert.Contains(t, []string{"shared", "unknown"}, throughputType, "Should be shared (Azure) or unknown (emulator)")
		assert.Contains(t, throughput, "message", "Should have message explaining throughput status")
	})
}

func TestCreateContainer(t *testing.T) {

	tests := []struct {
		name           string
		input          CreateContainerToolInput
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "valid arguments",
			input: CreateContainerToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:         testOperationDBName,
				Container:        "testContainer_new_1",
				PartitionKeyPath: "/id",
			},
			expectError: false,
		},
		{
			name: "valid arguments with throughput",
			input: CreateContainerToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:         testOperationDBName,
				Container:        "testContainer_new_2",
				PartitionKeyPath: "/id",
				Throughput:       func() *int32 { v := int32(400); return &v }(),
			},
			expectError: false,
		},
		{
			name: "empty account name",
			input: CreateContainerToolInput{
				ConnectionConfig: ConnectionConfig{Account: ""},
				Database:         testOperationDBName,
				Container:        "testContainer",
				PartitionKeyPath: "/id",
			},
			expectError:    true,
			expectedErrMsg: "account name is required",
		},
		{
			name: "empty database name",
			input: CreateContainerToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:         "",
				Container:        "testContainer",
				PartitionKeyPath: "/id",
			},
			expectError:    true,
			expectedErrMsg: "cosmos db database name missing",
		},
		{
			name: "empty container name",
			input: CreateContainerToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:         testOperationDBName,
				Container:        "",
				PartitionKeyPath: "/id",
			},
			expectError:    true,
			expectedErrMsg: "container name missing",
		},
		{
			name: "empty partition key path",
			input: CreateContainerToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:         testOperationDBName,
				Container:        "testContainer",
				PartitionKeyPath: "",
			},
			expectError:    true,
			expectedErrMsg: "partition key path missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			_, response, err := CreateContainerToolHandler(context.Background(), nil, test.input)

			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "dummy_account_does_not_matter", response.Account)
			assert.Equal(t, testOperationDBName, response.Database)
			assert.Equal(t, test.input.Container, response.Container)
			assert.Contains(t, response.Message, "created successfully")
		})
	}
}

func TestAddItemToContainer(t *testing.T) {

	tests := []struct {
		name           string
		input          AddItemToContainerToolInput
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "valid arguments",
			input: AddItemToContainerToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				PartitionKey: "user1",
				Item:         `{"id": "user1", "value": "user1@foo.com"}`,
			},
			expectError: false,
		},
		// {
		// 	name: "invalid partition key",
		// 	input: AddItemToContainerToolInput{
		// 		ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
		// 		Database:     testOperationDBName,
		// 		Container:    testOperationContainerName,
		// 		PartitionKey: "1",
		// 		Item:         `{"id": "testItem", "value": "testValue"}`,
		// 	},
		// 	expectError:    true,
		// 	expectedErrMsg: "error adding item to container",
		// },
		{
			name: "missing id attribute",
			input: AddItemToContainerToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				PartitionKey: "1",
				Item:         `{"value": "testValue"}`,
			},
			expectError:    true,
			expectedErrMsg: "error adding item to container",
		},
		{
			name: "empty account name",
			input: AddItemToContainerToolInput{
				ConnectionConfig: ConnectionConfig{Account: ""},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				PartitionKey: "testPartitionKey",
				Item:         `{"id": "testItem", "value": "testValue"}`,
			},
			expectError:    true,
			expectedErrMsg: "account name is required",
		},
		{
			name: "empty database name",
			input: AddItemToContainerToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     "",
				Container:    testOperationContainerName,
				PartitionKey: "testPartitionKey",
				Item:         `{"id": "testItem", "value": "testValue"}`,
			},
			expectError:    true,
			expectedErrMsg: "cosmos db database name missing",
		},
		{
			name: "empty container name",
			input: AddItemToContainerToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    "",
				PartitionKey: "testPartitionKey",
				Item:         `{"id": "testItem", "value": "testValue"}`,
			},
			expectError:    true,
			expectedErrMsg: "container name missing",
		},
		{
			name: "empty partition key",
			input: AddItemToContainerToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				PartitionKey: "",
				Item:         `{"id": "testItem", "value": "testValue"}`,
			},
			expectError:    true,
			expectedErrMsg: "value for partition key missing",
		},
		{
			name: "empty item JSON",
			input: AddItemToContainerToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				PartitionKey: "testPartitionKey",
				Item:         "",
			},
			expectError:    true,
			expectedErrMsg: "item JSON missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			_, response, err := AddItemToContainerToolHandler(context.Background(), nil, test.input)

			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "dummy_account_does_not_matter", response.Account)
			assert.Equal(t, testOperationDBName, response.Database)
			assert.Equal(t, testOperationContainerName, response.Container)
			assert.Contains(t, response.Message, "added successfully")
		})
	}
}

func TestReadItem(t *testing.T) {

	id := "user2"
	partitionKeyValue := "user2"

	// First, add an item to the container to be read later
	_, _, err := AddItemToContainerToolHandler(context.Background(), nil, AddItemToContainerToolInput{
		ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
		Database:     testOperationDBName,
		Container:    testOperationContainerName,
		PartitionKey: partitionKeyValue,
		Item:         `{"id": "user2", "value": "user2@foo.com"}`,
	})

	require.NoError(t, err)

	tests := []struct {
		name           string
		input          ReadItemToolInput
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "valid arguments",
			input: ReadItemToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				ItemID:       id,
				PartitionKey: partitionKeyValue,
			},
			expectError: false,
		},
		{
			name: "empty account name",
			input: ReadItemToolInput{
				ConnectionConfig: ConnectionConfig{Account: ""},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				ItemID:       "testItem",
				PartitionKey: "testPartitionKey",
			},
			expectError:    true,
			expectedErrMsg: "account name is required",
		},
		{
			name: "empty database name",
			input: ReadItemToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     "",
				Container:    testOperationContainerName,
				ItemID:       "testItem",
				PartitionKey: "testPartitionKey",
			},
			expectError:    true,
			expectedErrMsg: "database name missing",
		},
		{
			name: "empty container name",
			input: ReadItemToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    "",
				ItemID:       "testItem",
				PartitionKey: "testPartitionKey",
			},
			expectError:    true,
			expectedErrMsg: "container name missing",
		},
		{
			name: "empty item ID",
			input: ReadItemToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				ItemID:       "",
				PartitionKey: "testPartitionKey",
			},
			expectError:    true,
			expectedErrMsg: "item ID missing",
		},
		{
			name: "empty partition key",
			input: ReadItemToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				ItemID:       "testItem",
				PartitionKey: "",
			},
			expectError:    true,
			expectedErrMsg: "partition key missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			_, response, err := ReadItemToolHandler(context.Background(), nil, test.input)

			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, response.Item)

			var item map[string]any
			err = json.Unmarshal([]byte(response.Item), &item)

			require.NoError(t, err)
			assert.Equal(t, id, item["id"].(string))
			assert.Equal(t, "user2@foo.com", item["value"].(string))
		})
	}
}

func TestExecuteQuery(t *testing.T) {

	partitionKeyValue := "user3"

	// First, add an item to the container to query later
	_, _, err := AddItemToContainerToolHandler(context.Background(), nil, AddItemToContainerToolInput{
		ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
		Database:     testOperationDBName,
		Container:    testOperationContainerName,
		PartitionKey: partitionKeyValue,
		Item:         `{"id": "user3", "value": "user3@foo.com"}`,
	})

	require.NoError(t, err)

	tests := []struct {
		name           string
		input          ExecuteQueryToolInput
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "valid arguments with partition key",
			input: ExecuteQueryToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				Query:        "SELECT * FROM c",
				PartitionKey: partitionKeyValue,
			},
			expectError: false,
		},
		{
			name: "valid arguments - no partition key",
			input: ExecuteQueryToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:  testOperationDBName,
				Container: testOperationContainerName,
				Query:     "SELECT * FROM c",
			},
			expectError: false,
		},
		{
			name: "empty account name",
			input: ExecuteQueryToolInput{
				ConnectionConfig: ConnectionConfig{Account: ""},
				Database:  testOperationDBName,
				Container: testOperationContainerName,
				Query:     "SELECT * FROM c",
			},
			expectError:    true,
			expectedErrMsg: "account name is required",
		},
		{
			name: "empty database name",
			input: ExecuteQueryToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:  "",
				Container: testOperationContainerName,
				Query:     "SELECT * FROM c",
			},
			expectError:    true,
			expectedErrMsg: "database name missing",
		},
		{
			name: "empty container name",
			input: ExecuteQueryToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:  testOperationDBName,
				Container: "",
				Query:     "SELECT * FROM c",
			},
			expectError:    true,
			expectedErrMsg: "container name missing",
		},
		{
			name: "empty query string",
			input: ExecuteQueryToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:  testOperationDBName,
				Container: testOperationContainerName,
				Query:     "",
			},
			expectError:    true,
			expectedErrMsg: "query string missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			_, response, err := ExecuteQueryToolHandler(context.Background(), nil, test.input)

			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, response.QueryResults)
			// assert.NotEmpty(t, response.QueryMetrics)
		})
	}
}

func TestBatchCreateItems(t *testing.T) {

	tests := []struct {
		name           string
		input          BatchCreateItemsToolInput
		expectError    bool
		expectedErrMsg string
		expectedCount  int
	}{
		{
			name: "valid batch with multiple items",
			input: BatchCreateItemsToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				PartitionKey: "batch_pk_1",
				Items: []string{
					`{"id": "batch_item_1", "category": "batch_pk_1", "value": "item1@foo.com"}`,
					`{"id": "batch_item_2", "category": "batch_pk_1", "value": "item2@foo.com"}`,
					`{"id": "batch_item_3", "category": "batch_pk_1", "value": "item3@foo.com"}`,
				},
			},
			expectError:   false,
			expectedCount: 3,
		},
		{
			name: "valid batch with single item",
			input: BatchCreateItemsToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				PartitionKey: "batch_pk_2",
				Items: []string{
					`{"id": "batch_single_item", "category": "batch_pk_2", "value": "single@foo.com"}`,
				},
			},
			expectError:   false,
			expectedCount: 1,
		},
		{
			name: "empty account name",
			input: BatchCreateItemsToolInput{
				ConnectionConfig: ConnectionConfig{Account: ""},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				PartitionKey: "batch_pk",
				Items:        []string{`{"id": "item1", "value": "test"}`},
			},
			expectError:    true,
			expectedErrMsg: "account name is required",
		},
		{
			name: "empty database name",
			input: BatchCreateItemsToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     "",
				Container:    testOperationContainerName,
				PartitionKey: "batch_pk",
				Items:        []string{`{"id": "item1", "value": "test"}`},
			},
			expectError:    true,
			expectedErrMsg: "cosmos db database name missing",
		},
		{
			name: "empty container name",
			input: BatchCreateItemsToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    "",
				PartitionKey: "batch_pk",
				Items:        []string{`{"id": "item1", "value": "test"}`},
			},
			expectError:    true,
			expectedErrMsg: "container name missing",
		},
		{
			name: "empty partition key",
			input: BatchCreateItemsToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				PartitionKey: "",
				Items:        []string{`{"id": "item1", "value": "test"}`},
			},
			expectError:    true,
			expectedErrMsg: "partition key value missing",
		},
		{
			name: "empty items array",
			input: BatchCreateItemsToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				PartitionKey: "batch_pk",
				Items:        []string{},
			},
			expectError:    true,
			expectedErrMsg: "items array is empty",
		},
		{
			name: "exceeds 100 items limit",
			input: BatchCreateItemsToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				PartitionKey: "batch_pk",
				Items:        generateItemsArray(101),
			},
			expectError:    true,
			expectedErrMsg: "batch exceeds maximum of 100 items",
		},
		{
			name: "missing id in item",
			input: BatchCreateItemsToolInput{
				ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
				Database:     testOperationDBName,
				Container:    testOperationContainerName,
				PartitionKey: "batch_pk_3",
				Items: []string{
					`{"id": "valid_item", "category": "batch_pk_3", "value": "valid"}`,
					`{"category": "batch_pk_3", "value": "missing_id"}`,
				},
			},
			expectError:    true,
			expectedErrMsg: "batch failed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			_, response, err := BatchCreateItemsToolHandler(context.Background(), nil, test.input)

			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "dummy_account_does_not_matter", response.Account)
			assert.Equal(t, testOperationDBName, response.Database)
			assert.Equal(t, testOperationContainerName, response.Container)
			assert.Equal(t, test.expectedCount, response.ItemsCreated)
			assert.Contains(t, response.Message, "Successfully created")
		})
	}
}

func TestBatchCreateItems_DuplicateId(t *testing.T) {
	// Test that batch fails atomically when duplicate id is encountered
	partitionKey := "batch_dup_pk"

	// First, create an item
	_, _, err := AddItemToContainerToolHandler(context.Background(), nil, AddItemToContainerToolInput{
		ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
		Database:     testOperationDBName,
		Container:    testOperationContainerName,
		PartitionKey: partitionKey,
		Item:         `{"id": "existing_item", "category": "batch_dup_pk", "value": "existing"}`,
	})
	require.NoError(t, err)

	// Now try to batch create items including the duplicate
	_, _, err = BatchCreateItemsToolHandler(context.Background(), nil, BatchCreateItemsToolInput{
		ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
		Database:     testOperationDBName,
		Container:    testOperationContainerName,
		PartitionKey: partitionKey,
		Items: []string{
			`{"id": "new_item_1", "category": "batch_dup_pk", "value": "new1"}`,
			`{"id": "existing_item", "category": "batch_dup_pk", "value": "duplicate"}`,
			`{"id": "new_item_2", "category": "batch_dup_pk", "value": "new2"}`,
		},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "batch failed")
}

func TestBatchCreateItems_MaxLimit(t *testing.T) {
	// Test that exactly 100 items works (boundary test)
	partitionKey := "batch_max_pk"

	items := generateItemsArrayWithPrefix(100, partitionKey)

	_, response, err := BatchCreateItemsToolHandler(context.Background(), nil, BatchCreateItemsToolInput{
		ConnectionConfig: ConnectionConfig{Account: "dummy_account_does_not_matter"},
		Database:     testOperationDBName,
		Container:    testOperationContainerName,
		PartitionKey: partitionKey,
		Items:        items,
	})

	require.NoError(t, err)
	assert.Equal(t, 100, response.ItemsCreated)
}

// Helper function to generate an array of items for testing
func generateItemsArray(count int) []string {
	items := make([]string, count)
	for i := 0; i < count; i++ {
		items[i] = fmt.Sprintf(`{"id": "gen_item_%d", "value": "value_%d"}`, i, i)
	}
	return items
}

// Helper function to generate items with a specific partition key value
func generateItemsArrayWithPrefix(count int, partitionKey string) []string {
	items := make([]string, count)
	for i := range count {
		items[i] = fmt.Sprintf(`{"id": "max_item_%d", "category": "%s", "value": "value_%d"}`, i, partitionKey, i)
	}
	return items
}
