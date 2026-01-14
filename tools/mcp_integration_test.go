package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPIntegration_ListDatabases tests the list_databases tool through the full MCP stack
func TestMCPIntegration_ListDatabases(t *testing.T) {
	ctx := context.Background()

	// Create MCP server and register tools
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-cosmosdb-server",
		Version: "0.0.1",
	}, nil)

	mcp.AddTool(server, ListDatabases(), ListDatabasesToolHandler)

	// Create in-memory transports for testing
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	// Connect server
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	// Connect client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Call the list_databases tool via MCP protocol
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_databases",
		Arguments: map[string]any{
			"account": "dummy_account_does_not_matter",
		},
	})

	// Verify the call succeeded
	require.NoError(t, err, "CallTool should not return an error")
	require.NotNil(t, result, "Result should not be nil")
	require.False(t, result.IsError, "Result should not be an error")
	require.NotEmpty(t, result.Content, "Result content should not be empty")

	// Parse the response content
	// The content should be a TextContent with JSON
	require.Len(t, result.Content, 1, "Should have exactly one content item")

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Content should be TextContent")

	// Parse the JSON response
	var response ListDatabasesToolResult
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify the response contains the expected database
	assert.NotEmpty(t, response.Databases, "Should have at least one database")
	assert.Contains(t, response.Databases, testOperationDBName, "Should contain the test database")
}

// TestMCPIntegration_CreateDatabase tests the create_database tool through the full MCP stack
func TestMCPIntegration_CreateDatabase(t *testing.T) {
	ctx := context.Background()

	// Create MCP server and register tools
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-cosmosdb-server",
		Version: "0.0.1",
	}, nil)

	mcp.AddTool(server, CreateDatabase(), CreateDatabaseToolHandler)

	// Create in-memory transports for testing
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	// Connect server
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	// Connect client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Call the create_database tool via MCP protocol
	databaseName := "mcp_integration_test_db"
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "create_database",
		Arguments: map[string]any{
			"account":  "dummy_account_does_not_matter",
			"database": databaseName,
		},
	})

	// Verify the call succeeded
	require.NoError(t, err, "CallTool should not return an error")
	require.NotNil(t, result, "Result should not be nil")
	require.False(t, result.IsError, "Result should not be an error")
	require.NotEmpty(t, result.Content, "Result content should not be empty")

	// Parse the response content
	// The content should be a TextContent with JSON
	require.Len(t, result.Content, 1, "Should have exactly one content item")

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Content should be TextContent")

	// Parse the JSON response
	var response CreateDatabaseToolResult
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify the response contains expected data
	assert.Equal(t, "dummy_account_does_not_matter", response.Account, "Account should match")
	assert.Equal(t, databaseName, response.Database, "Database name should match")
	assert.Contains(t, response.Message, "created successfully", "Message should indicate success")
}

// TestMCPIntegration_ListContainers tests the list_containers tool through the full MCP stack
// to be investigted: in vNext emulator, this returns 400 error with message "id is required in the request body"
// skipping for now
func TestMCPIntegration_ListContainers(t *testing.T) {
	t.Skip("Skipping due to vNext emulator issue: returns 400 error with message 'id is required in the request body'")
	ctx := context.Background()

	// Create MCP server and register tools
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-cosmosdb-server",
		Version: "0.0.1",
	}, nil)

	mcp.AddTool(server, ListContainers(), ListContainersToolHandler)

	// Create in-memory transports for testing
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	// Connect server
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	// Connect client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Call the list_containers tool via MCP protocol
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_containers",
		Arguments: map[string]any{
			"account":  "dummy_account_does_not_matter",
			"database": testOperationDBName,
		},
	})

	// Verify the call succeeded
	require.NoError(t, err, "CallTool should not return an error")
	require.NotNil(t, result, "Result should not be nil")
	require.False(t, result.IsError, "Result should not be an error")
	require.NotEmpty(t, result.Content, "Result content should not be empty")

	// Parse the response content
	// The content should be a TextContent with JSON
	require.Len(t, result.Content, 1, "Should have exactly one content item")

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Content should be TextContent")

	// Parse the JSON response
	var response ListContainersToolResult
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify the response contains the expected container
	assert.Equal(t, "dummy_account_does_not_matter", response.Account, "Account should match")
	assert.Equal(t, testOperationDBName, response.Database, "Database should match")
	assert.NotEmpty(t, response.Containers, "Should have at least one container")
	assert.Contains(t, response.Containers, testOperationContainerName, "Should contain the test container")
}

// TestMCPIntegration_ReadItem tests the read_item tool through the full MCP stack
func TestMCPIntegration_ReadItem(t *testing.T) {
	ctx := context.Background()

	// First, add an item to the container to be read later
	id := "mcp_test_user1"
	partitionKeyValue := "mcp_test_user1"

	_, _, err := AddItemToContainerToolHandler(ctx, nil, AddItemToContainerToolInput{
		Account:      "dummy_account_does_not_matter",
		Database:     testOperationDBName,
		Container:    testOperationContainerName,
		PartitionKey: partitionKeyValue,
		Item:         `{"id": "mcp_test_user1", "email": "mcp_test_user1@example.com", "name": "MCP Test User"}`,
	})
	require.NoError(t, err, "Failed to add test item")

	// Create MCP server and register tools
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-cosmosdb-server",
		Version: "0.0.1",
	}, nil)

	mcp.AddTool(server, ReadItem(), ReadItemToolHandler)

	// Create in-memory transports for testing
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	// Connect server
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	// Connect client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Call the read_item tool via MCP protocol
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "read_item",
		Arguments: map[string]any{
			"account":      "dummy_account_does_not_matter",
			"database":     testOperationDBName,
			"container":    testOperationContainerName,
			"itemID":       id,
			"partitionKey": partitionKeyValue,
		},
	})

	// Verify the call succeeded
	require.NoError(t, err, "CallTool should not return an error")
	require.NotNil(t, result, "Result should not be nil")
	require.False(t, result.IsError, "Result should not be an error")
	require.NotEmpty(t, result.Content, "Result content should not be empty")

	// Parse the response content
	// The content should be a TextContent with JSON
	require.Len(t, result.Content, 1, "Should have exactly one content item")

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Content should be TextContent")

	// Parse the JSON response
	var response ReadItemToolResult
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify the item data
	assert.NotEmpty(t, response.Item, "Item should not be empty")

	// Parse the item itself
	var item map[string]any
	err = json.Unmarshal([]byte(response.Item), &item)
	require.NoError(t, err, "Item should be valid JSON")

	assert.Equal(t, id, item["id"], "Item ID should match")
	assert.Equal(t, "mcp_test_user1@example.com", item["email"], "Email should match")
	assert.Equal(t, "MCP Test User", item["name"], "Name should match")
}

// TestMCPIntegration_ReadContainerMetadata tests the read_container_metadata tool through the full MCP stack
func TestMCPIntegration_ReadContainerMetadata(t *testing.T) {
	ctx := context.Background()

	// Create MCP server and register tools
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-cosmosdb-server",
		Version: "0.0.1",
	}, nil)

	mcp.AddTool(server, ReadContainerMetadata(), ReadContainerMetadataToolHandler)

	// Create in-memory transports for testing
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	// Connect server
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	// Connect client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Call the read_container_metadata tool via MCP protocol
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "read_container_metadata",
		Arguments: map[string]any{
			"account":   "dummy_account_does_not_matter",
			"database":  testOperationDBName,
			"container": testOperationContainerName,
		},
	})

	// Verify the call succeeded
	require.NoError(t, err, "CallTool should not return an error")
	require.NotNil(t, result, "Result should not be nil")
	require.False(t, result.IsError, "Result should not be an error")
	require.NotEmpty(t, result.Content, "Result content should not be empty")

	// Parse the response content
	// The content should be a TextContent with JSON
	require.Len(t, result.Content, 1, "Should have exactly one content item")

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Content should be TextContent")

	// Parse the JSON response
	var response map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify the response contains expected metadata
	assert.Equal(t, testOperationContainerName, response["container_id"], "Container ID should match")
	assert.NotNil(t, response["partition_key_definition"], "Should have partition key definition")
	assert.NotNil(t, response["indexing_policy"], "Should have indexing policy")
}

// TestMCPIntegration_CreateContainer tests the create_container tool through the full MCP stack
func TestMCPIntegration_CreateContainer(t *testing.T) {
	ctx := context.Background()

	// Create MCP server and register tools
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-cosmosdb-server",
		Version: "0.0.1",
	}, nil)

	mcp.AddTool(server, CreateContainer(), CreateContainerToolHandler)

	// Create in-memory transports for testing
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	// Connect server
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	// Connect client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Call the create_container tool via MCP protocol
	containerName := "mcp_test_container_new"
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "create_container",
		Arguments: map[string]any{
			"account":          "dummy_account_does_not_matter",
			"database":         testOperationDBName,
			"container":        containerName,
			"partitionKeyPath": "/id",
		},
	})

	// Verify the call succeeded
	require.NoError(t, err, "CallTool should not return an error")
	require.NotNil(t, result, "Result should not be nil")
	require.False(t, result.IsError, "Result should not be an error")
	require.NotEmpty(t, result.Content, "Result content should not be empty")

	// Parse the response content
	// The content should be a TextContent with JSON
	require.Len(t, result.Content, 1, "Should have exactly one content item")

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Content should be TextContent")

	// Parse the JSON response
	var response CreateContainerToolResult
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify the response contains expected data
	assert.Equal(t, "dummy_account_does_not_matter", response.Account, "Account should match")
	assert.Equal(t, testOperationDBName, response.Database, "Database should match")
	assert.Equal(t, containerName, response.Container, "Container name should match")
	assert.Contains(t, response.Message, "created successfully", "Message should indicate success")
}

// TestMCPIntegration_AddItemToContainer tests the add_item_to_container tool through the full MCP stack
func TestMCPIntegration_AddItemToContainer(t *testing.T) {
	ctx := context.Background()

	// Create MCP server and register tools
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-cosmosdb-server",
		Version: "0.0.1",
	}, nil)

	mcp.AddTool(server, AddItemToContainer(), AddItemToContainerToolHandler)

	// Create in-memory transports for testing
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	// Connect server
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	// Connect client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Call the add_item_to_container tool via MCP protocol
	itemID := "mcp_add_test_item"
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "add_item_to_container",
		Arguments: map[string]any{
			"account":      "dummy_account_does_not_matter",
			"database":     testOperationDBName,
			"container":    testOperationContainerName,
			"partitionKey": itemID,
			"item":         `{"id": "mcp_add_test_item", "name": "MCP Add Test", "value": "test@example.com"}`,
		},
	})

	// Verify the call succeeded
	require.NoError(t, err, "CallTool should not return an error")
	require.NotNil(t, result, "Result should not be nil")
	require.False(t, result.IsError, "Result should not be an error")
	require.NotEmpty(t, result.Content, "Result content should not be empty")

	// Parse the response content
	// The content should be a TextContent with JSON
	require.Len(t, result.Content, 1, "Should have exactly one content item")

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Content should be TextContent")

	// Parse the JSON response
	var response AddItemToContainerToolResult
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify the response contains expected data
	assert.Equal(t, "dummy_account_does_not_matter", response.Account, "Account should match")
	assert.Equal(t, testOperationDBName, response.Database, "Database should match")
	assert.Equal(t, testOperationContainerName, response.Container, "Container should match")
	assert.Contains(t, response.Message, "added successfully", "Message should indicate success")
}

// TestMCPIntegration_ExecuteQuery tests the execute_query tool through the full MCP stack
func TestMCPIntegration_ExecuteQuery(t *testing.T) {
	ctx := context.Background()

	// First, add an item to the container to query later
	itemID := "mcp_query_test_item"
	_, _, err := AddItemToContainerToolHandler(ctx, nil, AddItemToContainerToolInput{
		Account:      "dummy_account_does_not_matter",
		Database:     testOperationDBName,
		Container:    testOperationContainerName,
		PartitionKey: itemID,
		Item:         `{"id": "mcp_query_test_item", "department": "Engineering", "email": "query_test@example.com"}`,
	})
	require.NoError(t, err, "Failed to add test item")

	// Create MCP server and register tools
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-cosmosdb-server",
		Version: "0.0.1",
	}, nil)

	mcp.AddTool(server, ExecuteQuery(), ExecuteQueryToolHandler)

	// Create in-memory transports for testing
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	// Connect server
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	// Connect client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Call the execute_query tool via MCP protocol with partition key
	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "execute_query",
		Arguments: map[string]any{
			"account":      "dummy_account_does_not_matter",
			"database":     testOperationDBName,
			"container":    testOperationContainerName,
			"query":        "SELECT * FROM c",
			"partitionKey": itemID,
		},
	})

	// Verify the call succeeded
	require.NoError(t, err, "CallTool should not return an error")
	require.NotNil(t, result, "Result should not be nil")
	require.False(t, result.IsError, "Result should not be an error")
	require.NotEmpty(t, result.Content, "Result content should not be empty")

	// Parse the response content
	// The content should be a TextContent with JSON
	require.Len(t, result.Content, 1, "Should have exactly one content item")

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Content should be TextContent")

	// Parse the JSON response
	var response ExecuteQueryToolResult
	err = json.Unmarshal([]byte(textContent.Text), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify the response contains query results
	assert.NotEmpty(t, response.QueryResults, "Should have query results")
	assert.GreaterOrEqual(t, len(response.QueryResults), 1, "Should have at least one result")

	// Parse the first result to verify it contains our test item
	var firstItem map[string]any
	err = json.Unmarshal([]byte(response.QueryResults[0]), &firstItem)
	require.NoError(t, err, "First result should be valid JSON")

	assert.Equal(t, itemID, firstItem["id"], "Item ID should match")
	assert.Equal(t, "Engineering", firstItem["department"], "Department should match")
	assert.Equal(t, "query_test@example.com", firstItem["email"], "Email should match")
}
