.PHONY: help upload-data parse-env start-mcp clean

# Default target
help:
	@echo "Available targets:"
	@echo "  make help         - Show this help message"
	@echo "  make upload-data  - Upload article data to Dgraph"
	@echo "  make parse-env    - Parse Dgraph connection string from .env file"
	@echo "  make start-mcp    - Start the MCP server"
	@echo "  make clean        - Clean up generated files"

# Upload article data to Dgraph
upload-data:
	@echo "Parsing connection string from .env file..."
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found"; \
		exit 1; \
	fi
	@DGRAPH_CONNECTION_STRING=$$(grep DGRAPH_CONNECTION_STRING .env | cut -d '=' -f2- | tr -d '"' | tr -d "'"); \
	if [ -z "$$DGRAPH_CONNECTION_STRING" ]; then \
		echo "Error: DGRAPH_CONNECTION_STRING not found in .env file"; \
		exit 1; \
	fi; \
	HOST=$$(echo "$$DGRAPH_CONNECTION_STRING" | sed 's|^dgraph://||' | cut -d':' -f1); \
	TOKEN=$$(echo "$$DGRAPH_CONNECTION_STRING" | grep -o 'bearertoken=[^&]*' | sed 's/bearertoken=//'); \
	echo "Using host: $$HOST"; \
	echo "Using token: $$TOKEN"; \
	echo "\n1. Updating schema..."; \
	curl -X POST "https://$$HOST/dgraph/alter" \
		--header "Authorization: Bearer $$TOKEN" \
		--header "Content-Type: application/dql" \
		--data-binary "@dgraph/schema.dql"; \
	echo "\n\n2. Uploading RDF data to Dgraph..."; \
	(echo '{ set {'; cat data/articles/nyt_articles_versions.rdf; echo '}}') > /tmp/wrapped_data.rdf; \
	curl -v -X POST "https://$$HOST/dgraph/mutate?commitNow=true&timeout=90s" \
		--header "Authorization: Bearer $$TOKEN" \
		--header "Content-Type: application/rdf" \
		--data-binary "@/tmp/wrapped_data.rdf";
	@echo "\n\nAll operations complete!"
	@rm /tmp/wrapped_data.rdf


# Parse environment variables from .env file
parse-env:
	@echo "Parsing connection string from .env file..."
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found"; \
		exit 1; \
	fi
	@DGRAPH_CONNECTION_STRING=$$(grep DGRAPH_CONNECTION_STRING .env | cut -d '=' -f2- | tr -d '"' | tr -d "'"); \
	if [ -z "$$DGRAPH_CONNECTION_STRING" ]; then \
		echo "Error: DGRAPH_CONNECTION_STRING not found in .env file"; \
		exit 1; \
	fi; \
	HOST=$$(echo "$$DGRAPH_CONNECTION_STRING" | sed 's|^dgraph://||' | cut -d':' -f1); \
	TOKEN=$$(echo "$$DGRAPH_CONNECTION_STRING" | grep -o 'bearertoken=[^&]*' | sed 's/bearertoken=//'); \
	echo "Host: $$HOST"; \
	echo "Bearer Token: $$TOKEN"; \
	echo "DGRAPH_HOST=$$HOST" > .env.parsed; \
	echo "DGRAPH_BEARER_TOKEN=$$TOKEN" >> .env.parsed; \
	echo "DGRAPH_ALPHA_HOST=$$HOST" >> .env.parsed; \
	echo "DGRAPH_ALPHA_PORT=443" >> .env.parsed; \
	echo "Environment variables written to .env.parsed"; \
	echo "To use these variables directly in your shell, run: source .env.parsed"

# Start the MCP server
start-mcp:
	@echo "Starting Dgraph MCP server..."
	cd mcp-server && modus dev

# Clean up generated files
clean:
	@echo "Cleaning up..."
	rm -f .env.parsed
	find . -type d -name "__pycache__" -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
