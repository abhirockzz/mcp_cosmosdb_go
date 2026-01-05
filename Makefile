# Simple Makefile for MCP Azure CosmosDB Server

BINARY = mcp_azure_cosmosdb_go
SRC = main.go

.PHONY: build run clean

build:
	go build -o $(BINARY) $(SRC)

run_server: build
	export COSMOSDB_MCP_SERVER_MODE=http && ./$(BINARY)

test: 
	go test -v ./...

mcp_inspector: build
	npx @modelcontextprotocol/inspector ./$(BINARY)

clean:
	rm -f $(BINARY)
