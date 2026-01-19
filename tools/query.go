package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func ReadItem() *mcp.Tool {

	return &mcp.Tool{
		Name:        "read_item",
		Description: "Read a specific item from a container in an Azure Cosmos DB database using the item ID and partition key",
	}
}

type ReadItemToolInput struct {
	Account      string `json:"account" jsonschema:"Azure Cosmos DB account name"`
	Database     string `json:"database" jsonschema:"Name of the database"`
	Container    string `json:"container" jsonschema:"Name of the container to read data from"`
	ItemID       string `json:"itemID" jsonschema:"ID of the item to read"`
	PartitionKey string `json:"partitionKey" jsonschema:"Partition key value of the item"`
}

type ReadItemToolResult struct {
	Item string `json:"item" jsonschema:"The item data as JSON string"`
}

func ReadItemToolHandler(ctx context.Context, _ *mcp.CallToolRequest, input ReadItemToolInput) (*mcp.CallToolResult, ReadItemToolResult, error) {

	if input.Account == "" {
		return nil, ReadItemToolResult{}, errors.New("cosmos db account name missing")
	}

	if input.Database == "" {
		return nil, ReadItemToolResult{}, errors.New("database name missing")
	}

	if input.Container == "" {
		return nil, ReadItemToolResult{}, errors.New("container name missing")
	}

	if input.ItemID == "" {
		return nil, ReadItemToolResult{}, errors.New("item ID missing")
	}

	if input.PartitionKey == "" {
		return nil, ReadItemToolResult{}, errors.New("partition key missing")
	}

	client, err := GetCosmosClientFunc(input.Account)
	if err != nil {
		return nil, ReadItemToolResult{}, err
	}

	databaseClient, err := client.NewDatabase(input.Database)
	if err != nil {
		return nil, ReadItemToolResult{}, fmt.Errorf("error creating database client: %v", err)
	}

	containerClient, err := databaseClient.NewContainer(input.Container)
	if err != nil {
		return nil, ReadItemToolResult{}, fmt.Errorf("error creating container client: %v", err)
	}

	partitionKey := azcosmos.NewPartitionKeyString(input.PartitionKey)

	itemResponse, err := containerClient.ReadItem(ctx, partitionKey, input.ItemID, nil)
	if err != nil {
		return nil, ReadItemToolResult{}, fmt.Errorf("error reading item: %v", err)
	}

	return nil, ReadItemToolResult{Item: string(itemResponse.Value)}, nil
}

func ExecuteQuery() *mcp.Tool {

	return &mcp.Tool{
		Name: "execute_query",
		Description: `Execute a SQL query on a Cosmos DB container. Ensure that the query string is valid and adheres to Cosmos DB SQL syntax. To use a partition key in the query directly, add it in the WHERE clause. Example: SELECT * FROM c WHERE c.department='HR'.

IMPORTANT LIMITATION: The Azure Cosmos DB Gateway API (used by the Go SDK) only supports simple projections and filtering for cross-partition queries.

UNSUPPORTED cross-partition operations: TOP, ORDER BY, OFFSET LIMIT, Aggregates (COUNT, SUM, AVG, MIN, MAX), DISTINCT, GROUP BY.

WORKAROUNDS:
1. Provide a partition key value to scope the query to a single partition - this enables all query features.
2. For cross-partition queries, use only SELECT and WHERE clauses, then sort/limit/aggregate the results. Be transparent about these limitations and let the user know (brief note) when you do so.

For details, refer to https://learn.microsoft.com/en-us/rest/api/cosmos-db/querying-cosmosdb-resources-using-the-rest-api#queries-that-cannot-be-served-by-gateway`,
	}
}

type ExecuteQueryToolInput struct {
	Account      string `json:"account" jsonschema:"Azure Cosmos DB account name"`
	Database     string `json:"database" jsonschema:"Name of the database"`
	Container    string `json:"container" jsonschema:"Name of the container to query"`
	Query        string `json:"query" jsonschema:"The SQL query string to execute"`
	PartitionKey string `json:"partitionKey,omitempty" jsonschema:"The partition key value for the query. If provided, the query will be scoped to this partition."`
}

type ExecuteQueryToolResult struct {
	//QueryResults []json.RawMessage `json:"results" jsonschema:"Query results as JSON objects"`
	QueryResults []string `json:"results" jsonschema:"Query results as JSON strings"`
	//QueryMetrics []string `json:"metrics" jsonschema:"Query execution metrics"`
}

func ExecuteQueryToolHandler(ctx context.Context, _ *mcp.CallToolRequest, input ExecuteQueryToolInput) (*mcp.CallToolResult, ExecuteQueryToolResult, error) {

	if input.Account == "" {
		return nil, ExecuteQueryToolResult{}, errors.New("cosmos db account name missing")
	}

	if input.Database == "" {
		return nil, ExecuteQueryToolResult{}, errors.New("database name missing")
	}

	if input.Container == "" {
		return nil, ExecuteQueryToolResult{}, errors.New("container name missing")
	}

	if input.Query == "" {
		return nil, ExecuteQueryToolResult{}, errors.New("query string missing")
	}

	client, err := GetCosmosClientFunc(input.Account)
	if err != nil {
		return nil, ExecuteQueryToolResult{}, err
	}

	databaseClient, err := client.NewDatabase(input.Database)
	if err != nil {
		return nil, ExecuteQueryToolResult{}, fmt.Errorf("error creating database client: %v", err)
	}

	containerClient, err := databaseClient.NewContainer(input.Container)
	if err != nil {
		return nil, ExecuteQueryToolResult{}, fmt.Errorf("error creating container client: %v", err)
	}

	var partitionKey azcosmos.PartitionKey
	if input.PartitionKey != "" {
		partitionKey = azcosmos.NewPartitionKeyString(input.PartitionKey)
	} else {
		partitionKey = azcosmos.PartitionKey{} // Empty partition key for cross-partition queries
	}

	queryPager := containerClient.NewQueryItemsPager(input.Query, partitionKey, nil)

	var response ExecuteQueryToolResult

	for queryPager.More() {
		queryResponse, err := queryPager.NextPage(ctx)
		if err != nil {
			return nil, ExecuteQueryToolResult{}, fmt.Errorf("query page error: %v", err)
		}

		for _, item := range queryResponse.Items {
			response.QueryResults = append(response.QueryResults, string(item))
		}

		// Append query metrics if available
		// if queryResponse.QueryMetrics != nil {
		// 	response.QueryMetrics = append(response.QueryMetrics, *queryResponse.QueryMetrics)
		// }
		//response.QueryMetrics = append(response.QueryMetrics, *queryResponse.QueryMetrics)
	}

	return nil, response, nil
}
