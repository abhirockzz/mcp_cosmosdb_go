package main

import (
	"fmt"

	"github.com/abhirockzz/mcp_cosmosdb_go/tools"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"Azure Cosmos DB MCP server 🚀",
		"0.0.5",
		server.WithLogging(),
	)

	s.AddTool(tools.ListDatabases(tools.CosmosDBServiceClientRetriever{}))
	s.AddTool(tools.ListContainers())
	s.AddTool(tools.ReadContainerMetadata())
	s.AddTool(tools.CreateContainer())
	s.AddTool(tools.AddItemToContainer())
	s.AddTool(tools.ReadItem())
	s.AddTool(tools.ExecuteQuery())

	//fmt.Println("starting mcp go server for cosmosdb")

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
