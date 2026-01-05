package tools

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

// GetCosmosClientFunc is a function variable that can be overridden for testing
var GetCosmosClientFunc = GetCosmosDBClient

func GetCosmosDBClient(accountName string) (*azcosmos.Client, error) {
	endpoint := fmt.Sprintf("https://%s.documents.azure.com:443/", accountName)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating credential: %v", err)
	}

	client, err := azcosmos.NewClient(endpoint, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating Cosmos client: %v", err)
	}

	if err != nil {
		fmt.Printf("Error creating Cosmos client: %v\n", err)
		return nil, err
	}

	return client, nil
}
