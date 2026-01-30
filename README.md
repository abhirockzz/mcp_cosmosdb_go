# Go based implementation of an MCP server for Azure Cosmos DB

This is a Go based implementation of an MCP server for Azure Cosmos DB using the [Azure SDK for Go](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos) and the official [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk).

It works with the Azure Cosmos DB service and the [vNext emulator](https://learn.microsoft.com/en-us/azure/cosmos-db/emulator-linux), and exposes the following tools for interacting with Azure Cosmos DB:

1. **List Databases**: Retrieve a list of all databases in a Cosmos DB account.
2. **Create Database**: Create a new database in the Cosmos DB account.
3. **List Containers**: Retrieve a list of all containers in a specific database.
4. **Read Container Metadata**: Fetch metadata or configuration details of a specific container.
5. **Create Container**: Create a new container in a specified database with a defined partition key.
6. **Add Item to Container**: Add a new item to a specified container in a database.
7. **Read Item**: Read a specific item from a container using its ID and partition key.
8. **Execute Query**: Execute a SQL query on a Cosmos DB container with optional partition key scoping.
9. **Batch Create Items**: Add multiple items to a container using Transactional Batch operation.

⚠️ This project is not intended to replace the [Azure MCP Server](https://github.com/azure/azure-mcp) or [Azure Cosmos DB MCP Toolkit](https://github.com/AzureCosmosDB/MCPToolKit). Rather, it serves as an experimental **learning tool** that demonstrates how to combine the Azure Go SDK and MCP Go SDK to build AI tooling for Azure Cosmos DB.

▶️ Here is a demo using GitHub Copilot CLI, but same would work with [Agent Mode in Visual Studio Code](https://code.visualstudio.com/docs/copilot/chat/chat-agent-mode), or any other MCP compatible tool (Claude Code, etc.):

[![MCP server demo](https://img.youtube.com/vi/l6gSYNd1Txs/hqdefault.jpg)](https://www.youtube.com/watch?v=l6gSYNd1Txs)

## 🚀 How to Run

To start with, clone the GitHub repo and build the binary:

```bash
git clone https://github.com/abhirockzz/mcp_cosmosdb_go
cd mcp_cosmosdb_go

go build -o mcp_azure_cosmosdb_go main.go
```

**Note**: The MCP server uses the [DefaultAzureCredential](https://learn.microsoft.com/en-us/azure/developer/go/sdk/authentication/credential-chains#defaultazurecredential-overview) implementation from the Azure SDK for Go to authenticate with Azure Cosmos DB. This means that you can authenticate using various methods, including environment variables, managed identity, or Azure CLI login, among others. 

> This ^ is not applicable to the local emulator since it uses a well known key-based authentication.

This MCP server supports both Streamable HTTP and Stdio transports. You can run the MCP server in two modes:

- Locally on your machine as an HTTP server, or `stdio` process
- Deployed to a remote endpoint (like Azure App Service, Azure Container Apps, etc.) as an HTTP(s) server

### 💻 Local mode

Thanks to Streamable HTTP support, you can easily run this MCP server as an HTTP server locally on your machine.

Login using Azure CLI ([az login](https://learn.microsoft.com/en-us/cli/azure/authenticate-azure-cli)), or the Azure Developer CLI ([azd auth login](https://learn.microsoft.com/en-us/azure/developer/azure-developer-cli/authenticate-azure-developer-cli)). Since the MCP server uses `DefaultAzureCredential`, it will authenticate as the identity logged in to the Azure CLI or the Azure Developer CLI.

The user principal (identity) you are logged in with should have permissions ([control](https://learn.microsoft.com/en-us/azure/cosmos-db/nosql/security/how-to-grant-control-plane-role-based-access?tabs=built-in-definition%2Ccsharp&pivots=azure-interface-cli) and [data plane](https://learn.microsoft.com/en-us/azure/cosmos-db/nosql/security/how-to-grant-data-plane-role-based-access?tabs=built-in-definition%2Ccsharp&pivots=azure-interface-cli)) to execute CRUD operations on database, container, and items.

**🌐 HTTP server**

Start the server:

```bash
export COSMOSDB_MCP_SERVER_MODE=http
./mcp_azure_cosmosdb_go
```

This will start the server on port `9090` by default. You can change the port by setting the `PORT` environment variable.

How you **configure** the MCP server will differ based on the MCP client/tool you use. For VS Code you can [follow these instructions](https://code.visualstudio.com/docs/copilot/chat/mcp-servers#_add-an-mcp-server) on how to configure this server using a `mcp.json` file.

Here is an example of the `mcp.json` configuration for the HTTP server:

```json
{
  "servers": {
    "mcp_azure_cosmosdb_go_http": {
      "type": "http",
      "url": "http://localhost:9090"
    }
  }
  //other MCP servers...
}
```

> Change the port if you have configured a different one.

**🖥️ Stdio server**

Here is an example of the `mcp.json` configuration for the `stdio` mode:

```json
{
  "servers": {
    "mcp_azure_cosmosdb_go_stdio": {
      "type": "stdio",
      "command": "./mcp_azure_cosmosdb_go"
    }
  }
  //other MCP servers...
}
```

Once you have configured the MCP server in your tool, you can start using it to interact with Azure Cosmos DB (just like in the demo shown above). For other tools like Claude Code, Claude Desktop, etc., refer to their respective documentation on how to configure an MCP HTTP/`stdio` server.

> Large Language Models (LLMs) are non-deterministic by nature and can make mistakes. **Always validate** the results and queries before making any decisions based on them.

For both the cases, if you want to use the vNext emulator, make sure its already running on your machine - `docker run -p 8081:8081 -p 1234:1234 mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:vnext-preview`

### ☁️ Remote endpoint

You can also deploy this MCP server to any cloud service (like Azure App Service, Azure Container Apps, etc.) and expose it as an HTTP(s) endpoint. The Azure service should support Managed Identity, and the MCP server will automatically pick up the credentials using the [DefaultAzureCredential](https://learn.microsoft.com/en-us/azure/developer/go/sdk/authentication/credential-chains#defaultazurecredential-overview) implementation.

⛔️ This execution mode is **not recommended**. Use this only for testing purposes. This is because, although MCP server can access Azure Cosmos DB securely using Managed Identity, it **does not** authenticate (or authorize) clients yet - anyone who can access the endpoint can execute operations on your Cosmos DB account.

## 🧪 Local dev and testing

Use [MCP inspector](https://modelcontextprotocol.io/docs/tools/inspector) - `make mcp_inspector`

![](images/mcp_inspector.png)
