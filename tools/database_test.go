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

func TestListDatabases(t *testing.T) {

	// tool, handler := ListDatabases(CosmosDBServiceClientRetriever{accountName: accountName})
	tool, handler := ListDatabases(CosmosDBEmulatorClientRetriever{})
	//tool, handler := ListDatabases()

	assert.Equal(t, tool.Name, LIST_DATABASES_TOOL_NAME)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "account")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"account"})

	req := mcp.CallToolRequest{
		Params: struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Arguments: map[string]interface{}{
				"account": "dummy_account_does_not_matter",
			},
		},
	}

	result, err := handler(context.Background(), req)
	require.NoError(t, err)

	textResult := getTextFromToolResult(t, result)

	var response ListDatabasesResponse
	err = json.Unmarshal([]byte(textResult), &response)
	require.NoError(t, err)
	assert.Equal(t, len(response.Databases), 1)

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
	client, err = CosmosDBEmulatorClientRetriever{}.Get("dummy_account_does_not_matter")
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
