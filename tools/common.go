package tools

import (
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

const ACCOUNT_PARAMETER_DESCRIPTION = "Name of the Cosmos DB account. If not available, ask the user to provide the account name. Do not use a random account name of your choice."
const LIST_DATABASES_TOOL_NAME = "list_databases"

type CosmosDBClientRetriever interface {
	Get(accountName string) (*azcosmos.Client, error)
}

type CosmosDBServiceClientRetriever struct {
	//accountName string
}

func (retriever CosmosDBServiceClientRetriever) Get(accountName string) (*azcosmos.Client, error) {
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
