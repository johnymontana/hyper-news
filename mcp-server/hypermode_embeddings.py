"""
Embedding module that uses models.hypermode.host to generate embeddings for text inputs.
This script can be invoked directly or imported into another Python script.
"""

import json
import requests
import sys
import os
from typing import List, Dict, Any, Union, Optional


def get_embeddings(
    text: str, 
    api_url: str = "https://models.hypermode.host/v1/embeddings", 
    token: Optional[str] = None,
    model: str = "nomic-ai/nomic-embed-text-v1.5"
) -> List[float]:
    """
    Generate embeddings for text input using models.hypermode.host API.
    
    Args:
        text: The text to generate embeddings for
        api_url: The API endpoint URL (default: https://models.hypermode.host/embeddings)
        token: Bearer token for authentication (required)
        model: The embedding model to use (default: nomic-embed-text)
        
    Returns:
        A list of floats representing the embedding vector
        
    Raises:
        ValueError: If token is not provided or if there's an error in the API call
    """
    if not token:
        # Check for token in environment variable
        token = os.environ.get("HYPERMODE_TOKEN")
        if not token:
            raise ValueError(
                "API token is required. Provide it as a parameter or set HYPERMODE_TOKEN environment variable."
            )

    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    
    payload = {
        "model": model,
        "input": text
    }
    
    try:
        response = requests.post(api_url, headers=headers, json=payload)
        response.raise_for_status()  # Raise exception for HTTP errors
        
        result = response.json()
        
        # Extract the embedding vector from the response
        # Adjust this based on the actual response structure from the API
        if "data" in result and len(result["data"]) > 0 and "embedding" in result["data"][0]:
            return result["data"][0]["embedding"]
        else:
            raise ValueError(f"Unexpected response format: {result}")
            
    except requests.RequestException as e:
        raise ValueError(f"API request failed: {str(e)}")


def get_embeddings_batch(
    texts: List[str],
    api_url: str = "https://models.hypermode.host/v1/embeddings",
    token: Optional[str] = None,
    model: str = "nomic-ai/nomic-embed-text-v1.5"
) -> List[List[float]]:
    """
    Generate embeddings for a batch of text inputs.
    
    Args:
        texts: List of texts to generate embeddings for
        api_url: The API endpoint URL
        token: Bearer token for authentication (required)
        model: The embedding model to use
        
    Returns:
        A list of embedding vectors (each is a list of floats)
    """
    embeddings = []
    for text in texts:
        embedding = get_embeddings(text, api_url, token, model)
        embeddings.append(embedding)
    return embeddings


def main():
    """Command line interface for generating embeddings."""
    import argparse
    
    parser = argparse.ArgumentParser(description="Generate embeddings for text using hypermode.host")
    parser.add_argument("text", help="Text to generate embeddings for")
    parser.add_argument(
        "--token", 
        help="Bearer token for API authentication (can also use HYPERMODE_TOKEN env var)"
    )
    parser.add_argument(
        "--model", 
        default="nomic-ai/nomic-embed-text-v1.5",
        help="Embedding model to use (default: nomic-ai/nomic-embed-text-v1.5)"
    )
    parser.add_argument(
        "--url", 
        default="https://models.hypermode.host/v1/embeddings", 
        help="API endpoint URL"
    )
    parser.add_argument(
        "--output", 
        choices=["json", "plain"], 
        default="json", 
        help="Output format (default: json)"
    )
    
    args = parser.parse_args()
    
    try:
        embedding = get_embeddings(args.text, args.url, args.token, args.model)
        
        if args.output == "json":
            print(json.dumps({"embedding": embedding}))
        else:
            print(embedding)
            
    except ValueError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
