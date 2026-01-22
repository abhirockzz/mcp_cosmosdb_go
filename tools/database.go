package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func ListDatabases() *mcp.Tool {

	return &mcp.Tool{
		Name:        "list_databases",
		Description: "List all databases in the specified Azure Cosmos DB account",
	}
}

type ListDatabasesToolInput struct {
	Account string `json:"account" jsonschema:"Azure Cosmos DB account name"`
}

type ListDatabasesToolResult struct {
	Account   string   `json:"account"`
	Databases []string `json:"databases" jsonschema:"list of databases in the account"`
}

func ListDatabasesToolHandler(ctx context.Context, request *mcp.CallToolRequest, input ListDatabasesToolInput) (*mcp.CallToolResult, ListDatabasesToolResult, error) {

	if input.Account == "" {
		return nil, ListDatabasesToolResult{}, errors.New("cosmos db account name missing")
	}

	databaseNames := []string{}

	client, err := GetCosmosClientFunc(input.Account)
	if err != nil {
		return nil, ListDatabasesToolResult{}, err
	}

	queryPager := client.NewQueryDatabasesPager("select * from dbs d", nil)

	for queryPager.More() {
		queryResponse, err := queryPager.NextPage(context.Background())
		if err != nil {
			return nil, ListDatabasesToolResult{}, err
		}

		for _, db := range queryResponse.Databases {
			databaseNames = append(databaseNames, db.ID)
		}
	}

	return nil, ListDatabasesToolResult{Account: input.Account, Databases: databaseNames}, nil
}

func CreateDatabase() *mcp.Tool {
	return &mcp.Tool{
		Name:        "create_database",
		Description: "Create a new database in the specified Azure Cosmos DB account",
	}
}

type CreateDatabaseToolInput struct {
	Account  string `json:"account" jsonschema:"Azure Cosmos DB account name"`
	Database string `json:"database" jsonschema:"Name of the database to create"`
}

type CreateDatabaseToolResult struct {
	Account  string `json:"account"`
	Database string `json:"database"`
	Message  string `json:"message"`
}

func CreateDatabaseToolHandler(ctx context.Context, request *mcp.CallToolRequest, input CreateDatabaseToolInput) (*mcp.CallToolResult, CreateDatabaseToolResult, error) {

	if input.Account == "" {
		return nil, CreateDatabaseToolResult{}, errors.New("cosmos db account name missing")
	}

	if input.Database == "" {
		return nil, CreateDatabaseToolResult{}, errors.New("database name missing")
	}

	client, err := GetCosmosClientFunc(input.Account)
	if err != nil {
		return nil, CreateDatabaseToolResult{}, err
	}

	databaseProps := azcosmos.DatabaseProperties{ID: input.Database}
	_, err = client.CreateDatabase(ctx, databaseProps, nil)
	if err != nil {
		return nil, CreateDatabaseToolResult{}, fmt.Errorf("error creating database: %w", err)
	}

	return nil, CreateDatabaseToolResult{
		Account:  input.Account,
		Database: input.Database,
		Message:  fmt.Sprintf("Database '%s' created successfully", input.Database),
	}, nil
}
