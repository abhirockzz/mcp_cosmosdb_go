package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/abhirockzz/mcp_cosmosdb_go/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {

	server := newServer()

	// choose stdio or http server based on env variable

	if os.Getenv("COSMOSDB_MCP_SERVER_MODE") == "http" {
		// --- HTTP ---
		handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
			return server
		}, nil)

		port := os.Getenv("SERVER_PORT")
		if port == "" {
			port = "9090"
		}

		log.Printf("Starting HTTP server on port %s\n", port)
		log.Fatal(http.ListenAndServe(":"+port, handler))
	} else {
		// --- STDIO ---
		log.Printf("Starting STDIO server")

		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatalf("stdio server failed: %v", err)
		}
	}

}

func newServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:       "mcp_azure_cosmosdb_go",
		Title:      "Go based MCP server for Azure Cosmos DB using the Azure SDK for Go and the MCP Go SDK",
		Version:    "0.0.1",
		WebsiteURL: "https://github.com/abhirockzz/mcp_cosmosdb_go",
	}, nil)

	mcp.AddTool(server, tools.ListDatabases(), tools.ListDatabasesToolHandler)
	mcp.AddTool(server, tools.CreateDatabase(), tools.CreateDatabaseToolHandler)
	mcp.AddTool(server, tools.ListContainers(), tools.ListContainersToolHandler)
	mcp.AddTool(server, tools.ReadContainerMetadata(), tools.ReadContainerMetadataToolHandler)
	mcp.AddTool(server, tools.CreateContainer(), tools.CreateContainerToolHandler)
	mcp.AddTool(server, tools.AddItemToContainer(), tools.AddItemToContainerToolHandler)
	mcp.AddTool(server, tools.ReadItem(), tools.ReadItemToolHandler)
	mcp.AddTool(server, tools.ExecuteQuery(), tools.ExecuteQueryToolHandler)
	mcp.AddTool(server, tools.BatchCreateItems(), tools.BatchCreateItemsToolHandler)

	return server
}
