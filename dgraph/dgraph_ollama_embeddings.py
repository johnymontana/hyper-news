#!/usr/bin/env python3
"""
Script to query articles from Dgraph, generate embeddings for abstracts using Ollama,
and write the embeddings back to Dgraph.
"""

import json
import sys
import numpy as np
import requests
from tqdm import tqdm
import pydgraph

class DgraphOllamaEmbeddings:
    def __init__(self, dgraph_host="localhost", dgraph_port=9080, ollama_host="localhost", ollama_port=11434, model="nomic-embed-text"):
        """
        Initialize the DgraphOllamaEmbeddings class.
        
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
        client_stub = pydgraph.DgraphClientStub(f'{host}:{port}')
        return pydgraph.DgraphClient(client_stub)
    
    def query_articles(self, limit=1000):
        """
        Query articles from Dgraph.
        
        Args:
            limit (int): Maximum number of articles to retrieve
            
        Returns:
            list: List of article data with UIDs and abstracts
        """
        query = """
        query articles($limit: int) {
            articles(func: type(Article), first: $limit) @filter(NOT has(Article.embedding)) {
                uid
                Article.title
                Article.abstract
            }
        }
        """
        variables = {"$limit": str(limit)}
        
        try:
            res = self.dgraph_client.txn(read_only=True).query(query, variables=variables)
            articles = json.loads(res.json)['articles']
            print(f"Retrieved {len(articles)} articles from Dgraph")
            return articles
        except Exception as e:
            print(f"Error querying Dgraph: {e}")
            return []
    
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
    
    def update_article_with_embedding(self, uid, embedding):
        """
        Update article in Dgraph with embedding.
        
        Args:
            uid (str): UID of the article
            embedding (list): Embedding vector
            
        Returns:
            bool: True if successful, False otherwise
        """
        if not embedding:
            return False
            
        # Convert embedding to JSON string to store in Dgraph
        embedding_json = json.dumps(embedding)
        
        mutation = {
            'set': [
                {
                    'uid': uid,
                    'Article.embedding': embedding_json
                }
            ]
        }
        
        try:
            txn = self.dgraph_client.txn()
            response = txn.mutate(set_obj=mutation['set'][0])
            txn.commit()
            return True
        except Exception as e:
            print(f"Error updating Dgraph: {e}")
            return False
        
    def process_articles(self, limit=100):
        """
        Main function to process articles, generate embeddings, and update Dgraph.
        
        Args:
            limit (int): Maximum number of articles to process
        """
        # Query articles
        articles = self.query_articles(limit)
        if not articles:
            print("No articles found or error querying Dgraph")
            return
        
        # Process each article
        success_count = 0
        for article in tqdm(articles, desc="Processing articles"):
            uid = article['uid']
            abstract = article.get('Article.abstract', '')
            
            if not abstract:
                print(f"Article {uid} has no abstract, skipping")
                continue
                
            # Generate embedding
            embedding = self.generate_embedding(abstract)
            if not embedding:
                print(f"Failed to generate embedding for article {uid}")
                continue
                
            # Update article with embedding
            success = self.update_article_with_embedding(uid, embedding)
            if success:
                success_count += 1
            
        print(f"Successfully updated {success_count} out of {len(articles)} articles with embeddings")

def main():
    """Main entry point for the script."""
    # Parse command line arguments
    import argparse
    parser = argparse.ArgumentParser(description='Generate embeddings for article abstracts in Dgraph')
    parser.add_argument('--dgraph-host', default='localhost', help='Dgraph host')
    parser.add_argument('--dgraph-port', type=int, default=9080, help='Dgraph port')
    parser.add_argument('--ollama-host', default='localhost', help='Ollama host')
    parser.add_argument('--ollama-port', type=int, default=11434, help='Ollama port')
    parser.add_argument('--model', default='nomic-embed-text', help='Ollama embedding model')
    parser.add_argument('--limit', type=int, default=10000, help='Maximum number of articles to process')
    
    args = parser.parse_args()
    
    # Create processor and run
    processor = DgraphOllamaEmbeddings(
        dgraph_host=args.dgraph_host,
        dgraph_port=args.dgraph_port,
        ollama_host=args.ollama_host,
        ollama_port=args.ollama_port,
        model=args.model
    )
    
    processor.process_articles(args.limit)

if __name__ == "__main__":
    main()
