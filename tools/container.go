package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func ListContainers() *mcp.Tool {

	return &mcp.Tool{
		Name:        "list_containers",
		Description: "List all containers in the specified Azure Cosmos DB database",
	}
}

type ListContainersToolInput struct {
	Account  string `json:"account" jsonschema:"Azure Cosmos DB account name"`
	Database string `json:"database" jsonschema:"Azure Cosmos DB database name"`
}

type ListContainersToolResult struct {
	Account    string   `json:"account"`
	Database   string   `json:"database"`
	Containers []string `json:"containers"`
}

func ListContainersToolHandler(ctx context.Context, _ *mcp.CallToolRequest, input ListContainersToolInput) (*mcp.CallToolResult, ListContainersToolResult, error) {

	accountName := input.Account

	if accountName == "" {
		return nil, ListContainersToolResult{}, errors.New("cosmos db account name missing")
	}

	database := input.Database

	if database == "" {
		return nil, ListContainersToolResult{}, errors.New("cosmos db database name missing")
	}

	client, err := GetCosmosClientFunc(accountName)
	if err != nil {
		return nil, ListContainersToolResult{}, err
	}

	databaseClient, err := client.NewDatabase(database)
	if err != nil {
		return nil, ListContainersToolResult{}, fmt.Errorf("error creating database client: %v", err)
	}

	containerPager := databaseClient.NewQueryContainersPager("select * from c", nil)

	containerNames := []string{}

	for containerPager.More() {
		containerResponse, err := containerPager.NextPage(ctx)
		if err != nil {
			var responseErr *azcore.ResponseError
			errors.As(err, &responseErr)
			return nil, ListContainersToolResult{}, err
		}

		for _, container := range containerResponse.Containers {
			containerNames = append(containerNames, container.ID)
		}
	}

	return nil, ListContainersToolResult{
		Account:    accountName,
		Database:   database,
		Containers: containerNames,
	}, nil

}

func ReadContainerMetadata() *mcp.Tool {

	return &mcp.Tool{
		Name:        "read_container_metadata",
		Description: "Read metadata of the specified container in Azure Cosmos DB",
	}
}

type ReadContainerMetadataToolInput struct {
	Account   string `json:"account" jsonschema:"Azure Cosmos DB account name"`
	Database  string `json:"database" jsonschema:"Azure Cosmos DB database name"`
	Container string `json:"container" jsonschema:"Azure Cosmos DB container name"`
}

type ReadContainerMetadataToolResult struct {
	ContainerID              string `json:"container_id"`
	DefaultTTL               *int32 `json:"default_ttl,omitempty"`
	IndexingPolicy           any    `json:"indexing_policy"`
	PartitionKeyDefinition   any    `json:"partition_key_definition"`
	ConflictResolutionPolicy any    `json:"conflict_resolution_policy"`
}

// func ReadContainerMetadataToolHandler(ctx context.Context, _ *mcp.CallToolRequest, input ReadContainerMetadataToolInput) (*mcp.CallToolResult, ReadContainerMetadataToolResult, error) {
func ReadContainerMetadataToolHandler(ctx context.Context, _ *mcp.CallToolRequest, input ReadContainerMetadataToolInput) (*mcp.CallToolResult, any, error) {

	accountName := input.Account

	if accountName == "" {
		return nil, ReadContainerMetadataToolResult{}, errors.New("cosmos db account name missing")
	}

	database := input.Database

	if database == "" {
		return nil, ReadContainerMetadataToolResult{}, errors.New("cosmos db database name missing")
	}

	container := input.Container

	if container == "" {
		return nil, ReadContainerMetadataToolResult{}, errors.New("container name missing")
	}

	client, err := GetCosmosClientFunc(accountName)
	if err != nil {
		return nil, ReadContainerMetadataToolResult{}, err
	}

	databaseClient, err := client.NewDatabase(database)
	if err != nil {
		return nil, ReadContainerMetadataToolResult{}, fmt.Errorf("error creating database client: %v", err)
	}

	containerClient, err := databaseClient.NewContainer(container)
	if err != nil {
		return nil, ReadContainerMetadataToolResult{}, fmt.Errorf("error creating container client: %v", err)
	}

	response, err := containerClient.Read(ctx, nil)
	if err != nil {
		var responseErr *azcore.ResponseError
		errors.As(err, &responseErr)
		return nil, ReadContainerMetadataToolResult{}, err
	}

	// Build throughput info
	var throughputInfo map[string]any
	throughputResp, throughputErr := containerClient.ReadThroughput(ctx, nil)
	if throughputErr != nil {
		// Check the error type to distinguish between shared throughput and other errors
		var responseErr *azcore.ResponseError
		if errors.As(throughputErr, &responseErr) {
			switch responseErr.StatusCode {
			case 404:
				// 404 means container has no dedicated throughput - uses database-level (shared)
				throughputInfo = map[string]any{
					"type":    "shared",
					"message": "Throughput is provisioned at database level",
				}
			case 400:
				// 400 typically means emulator limitation (offers endpoint not implemented)
				throughputInfo = map[string]any{
					"type":    "unknown",
					"message": "Unable to read throughput (emulator limitation or unsupported operation)",
				}
			default:
				// Other errors (e.g., 403 permission denied)
				throughputInfo = map[string]any{
					"type":    "error",
					"message": fmt.Sprintf("Failed to read throughput: %s", responseErr.ErrorCode),
				}
			}
		} else {
			throughputInfo = map[string]any{
				"type":    "error",
				"message": fmt.Sprintf("Failed to read throughput: %v", throughputErr),
			}
		}
	} else {
		if manual, ok := throughputResp.ThroughputProperties.ManualThroughput(); ok {
			throughputInfo = map[string]any{
				"type":          "manual",
				"ru_per_second": manual,
			}
		} else if maxRU, ok := throughputResp.ThroughputProperties.AutoscaleMaxThroughput(); ok {
			throughputInfo = map[string]any{
				"type":              "autoscale",
				"max_ru_per_second": maxRU,
			}
		}
	}

	metadata := map[string]any{
		"container_id":               response.ContainerProperties.ID,
		"default_ttl":                response.ContainerProperties.DefaultTimeToLive,
		"indexing_policy":            response.ContainerProperties.IndexingPolicy,
		"partition_key_definition":   response.ContainerProperties.PartitionKeyDefinition,
		"conflict_resolution_policy": response.ContainerProperties.ConflictResolutionPolicy,
		"unique_key_policy":          response.ContainerProperties.UniqueKeyPolicy,
		"throughput":                 throughputInfo,
	}

	jsonResult, err := json.Marshal(metadata)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshalling result to JSON: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonResult)},
		},
	}, nil, nil

	// return nil, ReadContainerMetadataToolResult{
	// 	ContainerID:              response.ContainerProperties.ID,
	// 	DefaultTTL:               response.ContainerProperties.DefaultTimeToLive,
	// 	IndexingPolicy:           response.ContainerProperties.IndexingPolicy,
	// 	PartitionKeyDefinition:   response.ContainerProperties.PartitionKeyDefinition,
	// 	ConflictResolutionPolicy: response.ContainerProperties.ConflictResolutionPolicy,
	// }, nil

}

func CreateContainer() *mcp.Tool {
	return &mcp.Tool{
		Name:        "create_container",
		Description: "Create a new container in the specified Azure Cosmos DB database",
	}
}

type CreateContainerToolInput struct {
	Account          string `json:"account" jsonschema:"Azure Cosmos DB account name"`
	Database         string `json:"database" jsonschema:"Azure Cosmos DB database name"`
	Container        string `json:"container" jsonschema:"Name of the container to create"`
	PartitionKeyPath string `json:"partitionKeyPath" jsonschema:"Partition key path for the container, example /id, /tentant, /category etc."`
	Throughput       *int32 `json:"throughput,omitempty" jsonschema:"Provisioned throughput for the container (optional)"`
}

type CreateContainerToolResult struct {
	Account   string `json:"account"`
	Database  string `json:"database"`
	Container string `json:"container"`
	Message   string `json:"message"`
}

func CreateContainerToolHandler(ctx context.Context, _ *mcp.CallToolRequest, input CreateContainerToolInput) (*mcp.CallToolResult, CreateContainerToolResult, error) {
	accountName := input.Account

	if accountName == "" {
		return nil, CreateContainerToolResult{}, errors.New("cosmos db account name missing")
	}

	database := input.Database

	if database == "" {
		return nil, CreateContainerToolResult{}, errors.New("cosmos db database name missing")
	}

	container := input.Container

	if container == "" {
		return nil, CreateContainerToolResult{}, errors.New("container name missing")
	}

	partitionKeyPath := input.PartitionKeyPath

	if partitionKeyPath == "" {
		return nil, CreateContainerToolResult{}, errors.New("partition key path missing")
	}

	client, err := GetCosmosClientFunc(accountName)
	if err != nil {
		return nil, CreateContainerToolResult{}, err
	}

	databaseClient, err := client.NewDatabase(database)
	if err != nil {
		return nil, CreateContainerToolResult{}, fmt.Errorf("error creating database client: %v", err)
	}

	properties := azcosmos.ContainerProperties{
		ID: container,
		PartitionKeyDefinition: azcosmos.PartitionKeyDefinition{
			Paths: []string{partitionKeyPath},
		},
	}

	if input.Throughput != nil {
		throughputProps := azcosmos.NewManualThroughputProperties(*input.Throughput)
		_, err = databaseClient.CreateContainer(ctx, properties, &azcosmos.CreateContainerOptions{
			ThroughputProperties: &throughputProps,
		})
	} else {
		_, err = databaseClient.CreateContainer(ctx, properties, nil)
	}

	if err != nil {
		var responseErr *azcore.ResponseError
		errors.As(err, &responseErr)
		return nil, CreateContainerToolResult{}, fmt.Errorf("error creating container: %v", err)
	}

	message := fmt.Sprintf("Container '%s' created successfully in database '%s'", container, database)

	return nil, CreateContainerToolResult{
		Account:   accountName,
		Database:  database,
		Container: container,
		Message:   message,
	}, nil
}

func AddItemToContainer() *mcp.Tool {
	return &mcp.Tool{
		Name:        "add_item_to_container",
		Description: "Add an item to the specified container in Azure Cosmos DB",
	}
}

type AddItemToContainerToolInput struct {
	Account      string `json:"account" jsonschema:"Azure Cosmos DB account name"`
	Database     string `json:"database" jsonschema:"Azure Cosmos DB database name"`
	Container    string `json:"container" jsonschema:"Name of the container to add the item to"`
	PartitionKey string `json:"partitionKey" jsonschema:"Partition key value for the item"`
	Item         string `json:"item" jsonschema:"The JSON representation of the item to add. id field is mandatory"`
}

type AddItemToContainerToolResult struct {
	Account   string `json:"account"`
	Database  string `json:"database"`
	Container string `json:"container"`
	Message   string `json:"message"`
}

func AddItemToContainerToolHandler(ctx context.Context, _ *mcp.CallToolRequest, input AddItemToContainerToolInput) (*mcp.CallToolResult, AddItemToContainerToolResult, error) {
	accountName := input.Account

	if accountName == "" {
		return nil, AddItemToContainerToolResult{}, errors.New("cosmos db account name missing")
	}

	database := input.Database

	if database == "" {
		return nil, AddItemToContainerToolResult{}, errors.New("cosmos db database name missing")
	}

	container := input.Container

	if container == "" {
		return nil, AddItemToContainerToolResult{}, errors.New("container name missing")
	}

	partitionKeyValue := input.PartitionKey

	if partitionKeyValue == "" {
		return nil, AddItemToContainerToolResult{}, errors.New("value for partition key missing")
	}

	itemJSON := input.Item

	if itemJSON == "" {
		return nil, AddItemToContainerToolResult{}, errors.New("item JSON missing")
	}

	client, err := GetCosmosClientFunc(accountName)
	if err != nil {
		return nil, AddItemToContainerToolResult{}, err
	}

	databaseClient, err := client.NewDatabase(database)
	if err != nil {
		return nil, AddItemToContainerToolResult{}, fmt.Errorf("error creating database client: %v", err)
	}

	containerClient, err := databaseClient.NewContainer(container)
	if err != nil {
		return nil, AddItemToContainerToolResult{}, fmt.Errorf("error creating container client: %v", err)
	}

	partitionKey := azcosmos.NewPartitionKeyString(partitionKeyValue)

	_, err = containerClient.CreateItem(ctx, partitionKey, []byte(itemJSON), nil)
	if err != nil {
		var responseErr *azcore.ResponseError
		errors.As(err, &responseErr)
		return nil, AddItemToContainerToolResult{}, fmt.Errorf("error adding item to container: %v", err)
	}

	message := fmt.Sprintf("Item added successfully to container '%s' in database '%s'", container, database)

	return nil, AddItemToContainerToolResult{
		Account:   accountName,
		Database:  database,
		Container: container,
		Message:   message,
	}, nil
}
