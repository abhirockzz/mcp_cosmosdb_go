package tools

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
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
			var responseErr *azcore.ResponseError
			errors.As(err, &responseErr)
			return nil, ListDatabasesToolResult{}, err
		}

		for _, db := range queryResponse.Databases {
			databaseNames = append(databaseNames, db.ID)
		}
	}

	return nil, ListDatabasesToolResult{Account: input.Account, Databases: databaseNames}, nil
}
