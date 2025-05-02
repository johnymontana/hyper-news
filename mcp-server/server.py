import json
import os
from typing import Dict, Optional, Any

import pydgraph
from mcp.server.fastmcp import FastMCP

# Initialize Dgraph client
def get_dgraph_client():
    """Create and return a Dgraph client instance"""
    # Get Dgraph Alpha host from environment or use default
    dgraph_host = os.environ.get("DGRAPH_ALPHA_HOST", "localhost")
    dgraph_port = os.environ.get("DGRAPH_ALPHA_PORT", "9080")
    
    # Create client stub
    client_stub = pydgraph.DgraphClientStub(f"{dgraph_host}:{dgraph_port}")
    
    # Create client
    return pydgraph.DgraphClient(client_stub)

# Initialize MCP server
mcp = FastMCP("Dgraph MCP Server")

@mcp.tool()
def get_dgraph_schema() -> str:
    """
    Get the current Dgraph schema
    
    Returns:
        The current Dgraph schema as a string
    """
    client = get_dgraph_client()
    query = """
    schema {}
    """
    
    try:
        response = client.txn(read_only=True).query(query)
        return response.json.decode('utf-8')
    except Exception as e:
        return f"Error retrieving schema: {str(e)}"

@mcp.tool()
def update_dgraph_schema(schema: str) -> str:
    """
    Update the Dgraph schema
    
    Args:
        schema: The schema definition to apply
    
    Returns:
        Result of the schema update operation
    """
    client = get_dgraph_client()
    
    try:
        return client.alter(pydgraph.Operation(schema=schema))
    except Exception as e:
        return f"Error updating schema: {str(e)}"

@mcp.tool()
def run_dql_query(query: str, variables: Optional[Dict[str, Any]] = None) -> str:
    """
    Run a DQL query against Dgraph
    
    Args:
        query: The DQL query to execute
        variables: Optional variables for the query
    
    Returns:
        Query results as a JSON string
    """
    client = get_dgraph_client()
    
    try:
        txn = client.txn(read_only=True)
        if variables:
            response = txn.query(query, variables=variables)
        else:
            response = txn.query(query)
        
        return response.json.decode('utf-8')
    except Exception as e:
        return f"Error executing query: {str(e)}"
    finally:
        if 'txn' in locals():
            txn.discard()

@mcp.tool()
def run_dql_mutation(mutation: str, set_json: Optional[Dict] = None, delete_json: Optional[Dict] = None, 
                     commit_now: bool = True) -> str:
    """
    Run a DQL mutation against Dgraph
    
    Args:
        mutation: The DQL mutation to execute
        set_json: Optional JSON data for set mutation
        delete_json: Optional JSON data for delete mutation
        commit_now: Whether to commit the transaction immediately
    
    Returns:
        Mutation results as a JSON string
    """
    client = get_dgraph_client()
    
    try:
        txn = client.txn()
        
        # Handle different types of mutations
        if mutation:
            mu = pydgraph.Mutation(set_nquads=mutation)
            response = txn.mutate(mutation=mu, commit_now=commit_now)
        elif set_json:
            mu = pydgraph.Mutation(set_json=json.dumps(set_json).encode('utf-8'))
            response = txn.mutate(mutation=mu, commit_now=commit_now)
        elif delete_json:
            mu = pydgraph.Mutation(delete_json=json.dumps(delete_json).encode('utf-8'))
            response = txn.mutate(mutation=mu, commit_now=commit_now)
        else:
            return "Error: No mutation data provided"
        
        if not commit_now:
            txn.commit()
            
        return json.dumps({
            "uids": response.uids,
            "message": "Mutation successful"
        })
    except Exception as e:
        if 'txn' in locals():
            txn.discard()
        return f"Error executing mutation: {str(e)}"

@mcp.resource("dgraph://info")
def dgraph_info() -> Dict[str, str]:
    """Return information about the Dgraph MCP server"""
    return {
        "name": "Dgraph MCP Server",
        "version": "1.0.0",
        "description": "MCP server for interacting with Dgraph database",
        "endpoints": [
            "get_dgraph_schema",
            "update_dgraph_schema",
            "run_dql_query",
            "run_dql_mutation"
        ]
    }