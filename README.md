# MCP server for Azure Cosmos DB using the Go SDK

This is a simple implementation of a MCP server for Cosmos DB built using its [Go SDK](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos). [mcp-go](https://github.com/mark3labs/mcp-go) project has been used as the MCP Go implementation.

Here is a demo (recommend watching at 2x speed 😉) using [VS Code Insiders in Agent mode](https://code.visualstudio.com/blogs/2025/02/24/introducing-copilot-agent-mode):

[![Demo: MCP server for Azure Cosmos DB using the Go SDK](https://img.youtube.com/vi/CsM-mspWJeM/0.jpg)](https://www.youtube.com/watch?v=CsM-mspWJeM)

## How to run

```bash
git clone https://github.com/abhirockzz/mcp_cosmosdb_go
cd mcp_cosmosdb_go

go build -o mcp_azure_cosmosdb main.go
```

Configure the MCP server:

```bash
mkdir -p .vscode

# Define the content for mcp.json
MCP_JSON_CONTENT=$(cat <<EOF
{
  "servers": {
    "CosmosDB Golang MCP": {
      "type": "stdio",
      "command": "$(pwd)/mcp_azure_cosmosdb"
    }
  }
}
EOF
)

# Write the content to mcp.json
echo "$MCP_JSON_CONTENT" > .vscode/mcp.json
```

## Azure Cosmos DB permissions and auth

- The user principal you will be using should have permissions to execute CRUD operations on database, container, and items.
- Authentication

  - You can either use Microsoft Entra ID (recommended) - login locally using `az cli` and the MCP server will use those credentials automatically.
  - Or, you can set the `COSMOSDB_ACCOUNT_KEY` environment variable in the MCP server configuration:

  ```json
  {
    "servers": {
      "CosmosDB Golang MCP": {
        "type": "stdio",
        "command": "/Users/demo/mcp_azure_cosmosdb",
        "env": {
          "COSMOSDB_ACCOUNT_KEY": "enter "
        }
      }
    }
  }
  ```

You are good to go! Now spin up VS Code Insiders in Agent Mode, or any other MCP tool (like Claude Desktop) and try this out!

## Local dev/testing

Start with [MCP inspector](https://modelcontextprotocol.io/docs/tools/inspector) - `npx @modelcontextprotocol/inspector ./mcp_azure_cosmosdb`

![](images/mcp_inspector.png)
