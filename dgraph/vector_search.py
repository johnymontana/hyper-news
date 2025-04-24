#!/usr/bin/env python3
"""
Script to perform vector similarity search in Dgraph using Ollama embeddings.
Takes a string input, generates an embedding, and finds similar articles in Dgraph.
"""

import json
import argparse
import requests
import pydgraph

class DgraphVectorSearch:
    def __init__(self, dgraph_host="localhost", dgraph_port=9080, ollama_host="localhost", ollama_port=11434, model="nomic-embed-text"):
        """
        Initialize the DgraphVectorSearch class.
        
        Args:
            dgraph_host (str): Dgraph host
            dgraph_port (int): Dgraph port
            ollama_host (str): Ollama host
            ollama_port (int): Ollama port
            model (str): Ollama embedding model to use
        """
        self.dgraph_client = self._create_dgraph_client(dgraph_host, dgraph_port)
        self.ollama_url = f"http://{ollama_host}:{ollama_port}/api/embeddings"
        self.model = model
        
    def _create_dgraph_client(self, host, port):
        """Create and return a Dgraph client."""

        channel_options = [
            ('grpc.max_send_message_length', 50 * 1024 * 1024),  # 50MB
            ('grpc.max_receive_message_length', 50 * 1024 * 1024),  # 50MB
            ('grpc.max_metadata_size', 32 * 1024)  # 32KB (increase from default 8KB)
        ]

        client_stub = pydgraph.DgraphClientStub(f'{host}:{port}', options=channel_options)
        return pydgraph.DgraphClient(client_stub)
    
    def generate_embedding(self, text):
        """
        Generate embeddings for text using Ollama.
        
        Args:
            text (str): Text to generate embedding for
            
        Returns:
            list: Embedding vector or None if failed
        """
        if not text:
            return None
            
        payload = {
            "model": self.model,
            "prompt": text
        }
        
        try:
            response = requests.post(self.ollama_url, json=payload)
            if response.status_code == 200:
                return response.json()['embedding']
            else:
                print(f"Error from Ollama API: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            print(f"Error generating embedding: {e}")
            return None
    
    def vector_similarity_search(self, embedding, limit=10):
        """
        Perform vector similarity search in Dgraph.
        
        Args:
            embedding (list): Query embedding vector
            limit (int): Maximum number of results to return
            
        Returns:
            list: List of similar articles
        """
        if not embedding:
            return []
            
        # Convert embedding to JSON string for the query
        embedding_json = json.dumps(embedding)
        
        # DQL query for vector similarity search using similar_to function
        query = """
        query vector_search($embedding: string, $limit: int) {
          articles(func: similar_to(Article.embedding, 10, $embedding)) {
            uid
            Article.title
            Article.abstract
            score
          }
        }
        """
        
        variables = {
            "$embedding": embedding_json,
            "$limit": str(limit)
        }
        
        try:
            res = self.dgraph_client.txn(read_only=True).query(query, variables=variables)
            results = json.loads(res.json)['articles']
            
            # Results are already sorted by similarity score from the similar_to function
            return results
        except Exception as e:
            print(f"Error performing vector search: {e}")
            return []
    
    def search(self, query_text, limit=10):
        """
        Search for similar articles based on query text.
        
        Args:
            query_text (str): Query text
            limit (int): Maximum number of results to return
            
        Returns:
            list: List of similar articles
        """
        # Generate embedding for query text
        embedding = self.generate_embedding(query_text)
        if not embedding:
            print("Failed to generate embedding for query text")
            return []
        
        # Perform vector similarity search
        results = self.vector_similarity_search(embedding, limit)
        return results

def main():
    """Main entry point for the script."""
    # Parse command line arguments
    parser = argparse.ArgumentParser(description='Vector similarity search in Dgraph using Ollama embeddings')
    parser.add_argument('query', help='Query text to search for similar articles')
    parser.add_argument('--dgraph-host', default='localhost', help='Dgraph host')
    parser.add_argument('--dgraph-port', type=int, default=9080, help='Dgraph port')
    parser.add_argument('--ollama-host', default='localhost', help='Ollama host')
    parser.add_argument('--ollama-port', type=int, default=11434, help='Ollama port')
    parser.add_argument('--model', default='nomic-embed-text', help='Ollama embedding model')
    parser.add_argument('--limit', type=int, default=10, help='Maximum number of results to return')
    
    args = parser.parse_args()
    
    # Create searcher and perform search
    searcher = DgraphVectorSearch(
        dgraph_host=args.dgraph_host,
        dgraph_port=args.dgraph_port,
        ollama_host=args.ollama_host,
        ollama_port=args.ollama_port,
        model=args.model
    )
    
    results = searcher.search(args.query, args.limit)
    
    # Display results
    if not results:
        print("No similar articles found")
        return
    
    print(f"Found {len(results)} similar articles:")
    for i, article in enumerate(results):
        print(f"\n{i+1}. {article.get('Article.title', 'No title')} (Score: {article.get('score', 0):.4f})")
        abstract = article.get('Article.abstract', 'No abstract')
        # Truncate abstract if too long
        if len(abstract) > 200:
            abstract = abstract[:200] + "..."
        print(f"   {abstract}")

if __name__ == "__main__":
    main()
