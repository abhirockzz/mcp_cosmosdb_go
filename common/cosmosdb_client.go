package common

import (
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

// GetCosmosDBClient initializes and returns a Cosmos DB client
func GetCosmosDBClient(accountName string) (*azcosmos.Client, error) {

	endpoint := fmt.Sprintf("https://%s.documents.azure.com:443/", accountName)

	accountKey := os.Getenv("COSMOSDB_ACCOUNT_KEY")
	// if only account name is provided, use managed identity
	if accountKey == "" {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("error creating credential: %v", err)
		}

		client, err := azcosmos.NewClient(endpoint, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating Cosmos client: %v", err)
		}

		return client, nil
	} else {
		// if both account name and key are provided, use the key
		cred, err := azcosmos.NewKeyCredential(accountKey)
		if err != nil {
			return nil, fmt.Errorf("error creating key credential: %v", err)
		}

		client, err := azcosmos.NewClientWithKey(endpoint, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating Cosmos client: %v", err)
		}

		return client, nil
	}

}

// GetDatabaseClient initializes and returns a Cosmos DB database client
func GetDatabaseClient(account, database string) (*azcosmos.DatabaseClient, error) {
	if database == "" {
		return nil, fmt.Errorf("database name is required")
	}
	client, err := GetCosmosDBClient(account)
	if err != nil {
		return nil, fmt.Errorf("error getting Cosmos client: %v", err)
	}
	databaseClient, err := client.NewDatabase(database)
	if err != nil {
		return nil, fmt.Errorf("error creating database client: %v", err)
	}

	return databaseClient, nil
}

// GetContainerClient initializes and returns a Cosmos DB container client
func GetContainerClient(account, database, container string) (*azcosmos.ContainerClient, error) {
	if container == "" {
		return nil, fmt.Errorf("container name is required")
	}
	databaseClient, err := GetDatabaseClient(account, database)
	if err != nil {
		return nil, fmt.Errorf("error creating container client: %v", err)
	}

	containerClient, err := databaseClient.NewContainer(container)
	if err != nil {
		return nil, fmt.Errorf("error creating container client: %v", err)
	}

	return containerClient, nil
}
