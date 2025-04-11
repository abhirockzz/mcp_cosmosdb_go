package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dummy_account_does_not_matter = "dummy_account_does_not_matter"

func TestListDatabases(t *testing.T) {

	tool, handler := ListDatabases(CosmosDBEmulatorClientRetriever{})

	assert.Equal(t, tool.Name, LIST_DATABASES_TOOL_NAME)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "account")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"account"})

	tests := []struct {
		name           string
		arguments      map[string]interface{}
		expectError    bool
		expectedResult string
		expectedErrMsg string
	}{
		{
			name: "valid account name",
			arguments: map[string]interface{}{
				"account": dummy_account_does_not_matter,
			},
			expectError:    false,
			expectedResult: testOperationDBName,
		},
		{
			name: "empplty account name",
			arguments: map[string]interface{}{
				"account": "",
			},
			expectError:    true,
			expectedErrMsg: "cosmos db account name missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Arguments: test.arguments,
				},
			}

			result, err := handler(context.Background(), req)
			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			textResult := getTextFromToolResult(t, result)

			var response ListDatabasesResponse
			err = json.Unmarshal([]byte(textResult), &response)
			require.NoError(t, err)
			assert.Equal(t, 1, len(response.Databases))
			assert.Equal(t, test.expectedResult, response.Databases[0])
		})
	}

}

func TestListContainers(t *testing.T) {

	tool, handler := ListContainers(CosmosDBEmulatorClientRetriever{})

	assert.Equal(t, tool.Name, LIST_CONTAINERS_TOOL_NAME)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "account")
	assert.Contains(t, tool.InputSchema.Properties, "database")

	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"account", "database"})

	tests := []struct {
		name           string
		arguments      map[string]interface{}
		expectError    bool
		expectedResult string
		expectedErrMsg string
	}{
		{
			name: "valid arguments",
			arguments: map[string]interface{}{
				"account":  dummy_account_does_not_matter,
				"database": testOperationDBName,
			},
			expectError:    false,
			expectedResult: testOperationContainerName,
		},
		{
			name: "empty account name",
			arguments: map[string]interface{}{
				"account":  "",
				"database": testOperationDBName,
			},
			expectError:    true,
			expectedErrMsg: "cosmos db account name missing",
		},
		{
			name: "empty database name",
			arguments: map[string]interface{}{
				"account":  dummy_account_does_not_matter,
				"database": "",
			},
			expectError:    true,
			expectedErrMsg: "database name missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Arguments: test.arguments,
				},
			}

			result, err := handler(context.Background(), req)
			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			textResult := getTextFromToolResult(t, result)

			var response ListContainersResponse
			err = json.Unmarshal([]byte(textResult), &response)
			require.NoError(t, err)

			assert.Equal(t, dummy_account_does_not_matter, response.Account)
			assert.Equal(t, testOperationDBName, response.Database)
			assert.Equal(t, 1, len(response.Containers))
			assert.Equal(t, test.expectedResult, response.Containers[0])
		})
	}
}

func TestReadContainerMetadata(t *testing.T) {
	tool, handler := ReadContainerMetadata(CosmosDBEmulatorClientRetriever{})

	assert.Equal(t, tool.Name, READ_CONTAINER_METADATA_TOOL_NAME)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "account")
	assert.Contains(t, tool.InputSchema.Properties, "database")
	assert.Contains(t, tool.InputSchema.Properties, "container")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"account", "database", "container"})

	tests := []struct {
		name           string
		arguments      map[string]interface{}
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "valid arguments",
			arguments: map[string]interface{}{
				"account":   dummy_account_does_not_matter,
				"database":  testOperationDBName,
				"container": testOperationContainerName,
			},
			expectError: false,
		},
		{
			name: "empty account name",
			arguments: map[string]interface{}{
				"account":   "",
				"database":  testOperationDBName,
				"container": testOperationContainerName,
			},
			expectError:    true,
			expectedErrMsg: "cosmos db account name missing",
		},
		{
			name: "empty database name",
			arguments: map[string]interface{}{
				"account":   dummy_account_does_not_matter,
				"database":  "",
				"container": testOperationContainerName,
			},
			expectError:    true,
			expectedErrMsg: "database name missing",
		},
		{
			name: "empty container name",
			arguments: map[string]interface{}{
				"account":   dummy_account_does_not_matter,
				"database":  testOperationDBName,
				"container": "",
			},
			expectError:    true,
			expectedErrMsg: "container name missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Arguments: test.arguments,
				},
			}

			result, err := handler(context.Background(), req)
			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			textResult := getTextFromToolResult(t, result)

			var metadata map[string]interface{}
			err = json.Unmarshal([]byte(textResult), &metadata)
			require.NoError(t, err)

			assert.Contains(t, metadata, "container_id")
			assert.Contains(t, metadata, "default_ttl")
			assert.Contains(t, metadata, "indexing_policy")
			assert.Contains(t, metadata, "partition_key_definition")
			assert.Contains(t, metadata, "conflict_resolution_policy")
		})
	}
}

func TestCreateContainer(t *testing.T) {
	tool, handler := CreateContainer(CosmosDBEmulatorClientRetriever{})

	assert.Equal(t, tool.Name, CREATE_CONTAINER_TOOL_NAME)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "account")
	assert.Contains(t, tool.InputSchema.Properties, "database")
	assert.Contains(t, tool.InputSchema.Properties, "container")
	assert.Contains(t, tool.InputSchema.Properties, "partitionKeyPath")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"account", "database", "container", "partitionKeyPath"})

	tests := []struct {
		name           string
		arguments      map[string]interface{}
		expectedResult string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "valid arguments",
			arguments: map[string]interface{}{
				"account":          dummy_account_does_not_matter,
				"database":         testOperationDBName,
				"container":        "testContainer_new_1",
				"partitionKeyPath": "/id",
			},
			expectedResult: fmt.Sprintf("Container '%s' created successfully in database '%s'", "testContainer_new_1", testOperationDBName),
			expectError:    false,
		},
		{
			name: "valid arguments with throughput",
			arguments: map[string]interface{}{
				"account":          dummy_account_does_not_matter,
				"database":         testOperationDBName,
				"container":        "testContainer_new_2",
				"partitionKeyPath": "/id",
				"throughput":       1000,
			},
			expectedResult: fmt.Sprintf("Container '%s' created successfully in database '%s'", "testContainer_new_2", testOperationDBName),
			expectError:    false,
		},
		{
			name: "empty account name",
			arguments: map[string]interface{}{
				"account":          "",
				"database":         testOperationDBName,
				"container":        "testContainer",
				"partitionKeyPath": "/id",
			},
			expectError:    true,
			expectedErrMsg: "cosmos db account name missing",
		},
		{
			name: "empty database name",
			arguments: map[string]interface{}{
				"account":          dummy_account_does_not_matter,
				"database":         "",
				"container":        "testContainer",
				"partitionKeyPath": "/id",
			},
			expectError:    true,
			expectedErrMsg: "database name missing",
		},
		{
			name: "empty container name",
			arguments: map[string]interface{}{
				"account":          dummy_account_does_not_matter,
				"database":         testOperationDBName,
				"container":        "",
				"partitionKeyPath": "/id",
			},
			expectError:    true,
			expectedErrMsg: "container name missing",
		},
		{
			name: "empty partition key path",
			arguments: map[string]interface{}{
				"account":          dummy_account_does_not_matter,
				"database":         testOperationDBName,
				"container":        "testContainer",
				"partitionKeyPath": "",
			},
			expectError:    true,
			expectedErrMsg: "partition key path missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Arguments: test.arguments,
				},
			}

			result, err := handler(context.Background(), req)
			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			textResult := getTextFromToolResult(t, result)
			assert.Equal(t, test.expectedResult, textResult)
			// assert.Contains(t, textResult, "Container '")
			// assert.Contains(t, textResult, "created successfully in database '")
		})
	}
}

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
	client, err := CosmosDBEmulatorClientRetriever{}.Get(dummy_account_does_not_matter)
	if err != nil {
		fmt.Printf("Failed to set up CosmosDB client: %v\n", err)
		os.Exit(1)
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
