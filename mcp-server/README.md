# Dgraph MCP Server

A Model Completion Protocol (MCP) server for interacting with a Dgraph database. This server provides tools and resources for managing Dgraph schemas, executing queries, and performing mutations using the MCP protocol.

## Overview

This MCP server provides a set of tools that allow Claude and other AI assistants to interact directly with a Dgraph database. It leverages the Python `pydgraph` client to communicate with Dgraph and exposes several useful operations as MCP tools.

## Features

- Get the current Dgraph schema
- Update the Dgraph schema
- Execute DQL (Dgraph Query Language) queries
- Perform DQL mutations (using RDF N-Quads or JSON)
- Configuration via environment variables

## Prerequisites

- Python 3.7+
- Dgraph running and accessible (default: localhost:9080)
- `pydgraph` package installed
- `mcp` package for the MCP server

## Configuration

The server can be configured using environment variables:

- `DGRAPH_ALPHA_HOST`: Hostname for Dgraph Alpha (default: "localhost")
- `DGRAPH_ALPHA_PORT`: Port for Dgraph Alpha (default: "9080")

## Available Tools

### get_dgraph_schema()

Retrieves the current Dgraph schema.

**Returns:**
- JSON string containing the schema information

**Example usage in Claude:**
```
Get the current Dgraph schema.
```

### update_dgraph_schema(schema: str)

Updates the Dgraph schema with the provided schema definition.

**Parameters:**
- `schema`: The schema definition in Dgraph Schema format

**Returns:**
- Result of the schema update operation

**Example usage in Claude:**
```
Update the Dgraph schema with the following definition:

type Person {
  name: string
  age: int
  friend: [Person]
}

name: string @index(exact) .
age: int @index(int) .
friend: [uid] .
```

### run_dql_query(query: str, variables: Optional[Dict] = None)

Executes a DQL query against the Dgraph database.

**Parameters:**
- `query`: The DQL query to execute
- `variables`: Optional variables for the query

**Returns:**
- JSON string containing the query results

**Example usage in Claude:**
```
Run the following Dgraph query:

query {
  people(func: type(Person)) {
    name
    age
    friend {
      name
    }
  }
}
```

### run_dql_mutation(mutation: str, set_json: Optional[Dict] = None, delete_json: Optional[Dict] = None, commit_now: bool = True)

Performs a DQL mutation on the Dgraph database.

**Parameters:**
- `mutation`: The DQL mutation in RDF N-Quads format
- `set_json`: Optional JSON data for set mutation
- `delete_json`: Optional JSON data for delete mutation
- `commit_now`: Whether to commit the transaction immediately (default: True)

**Returns:**
- JSON string containing the mutation results, including assigned UIDs

**Example usages in Claude:**

*RDF N-Quads mutation:*
```
Run the following Dgraph mutation:

_:person1 <name> "John Doe" .
_:person1 <age> "30" .
_:person1 <dgraph.type> "Person" .
```

*JSON mutation:*
```
Run a Dgraph mutation with the following JSON data:

{
  "set": [
    {
      "name": "Jane Smith",
      "age": 28,
      "dgraph.type": "Person"
    }
  ]
}
```

## Using in Claude Desktop

To use this MCP server with Claude Desktop, follow these steps:

1. **Start the MCP server:**

   ```bash
   # Navigate to the mcp-server directory
   cd /path/to/mcp-server
   
   # Start the MCP server with the Python module
   python -m main
   ```

2. **Configure Claude Desktop:**

   - Open Claude Desktop
   - Go to Settings > MCP Servers
   - Click "Add MCP Server"
   - Enter the server URL (usually http://localhost:8000 by default)
   - Click "Connect"

3. **Verify Connection:**

   - Claude should show that it's connected to the "Dgraph MCP Server"
   - You can ask Claude to list available tools to verify

4. **Using the Tools:**

   You can now ask Claude to perform any of the Dgraph operations listed above. For example:
   
   - "Get the current Dgraph schema"
   - "Run a query to find all Person nodes"
   - "Update the schema to add a new type"

## Example Workflows

### Creating a New Schema

```
1. Update the Dgraph schema with the following definition:

type Article {
  title: string
  content: string
  author: Person
  published: datetime
  tags: [string]
}

title: string @index(fulltext) .
content: string .
author: [uid] .
published: datetime @index(hour) .
tags: [string] @index(exact) .
```

### Querying Data

```
Run the following Dgraph query to find articles with "AI" in the title:

query {
  articles(func: type(Article)) @filter(anyoftext(title, "AI")) {
    title
    published
    author {
      name
    }
    tags
  }
}
```

### Adding Data

```
Run a Dgraph mutation with the following JSON data:

{
  "set": [
    {
      "dgraph.type": "Article",
      "title": "Advanced AI Techniques in 2025",
      "content": "This article explores the latest AI advancements...",
      "published": "2025-01-15T10:30:00Z",
      "tags": ["AI", "Machine Learning", "Technology"]
    }
  ]
}
```

## Troubleshooting

- **Connection Issues**: Ensure Dgraph is running and accessible at the configured host/port
- **Authentication Errors**: If using Dgraph with authentication, make sure to configure appropriate credentials
- **Schema Errors**: Verify your schema syntax follows Dgraph's schema format
- **Query Errors**: Check your DQL syntax for errors

## License

This project is licensed under the MIT License - see the LICENSE file for details.