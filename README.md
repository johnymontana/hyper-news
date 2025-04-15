# HyperNews

Knowledge graph + AI Agent analysis of news using multi modal GraphRAG approach with Dgraph and Modus.

![hyper-news graph data model](img/graph-model.png)

## Data

This project uses data from the [New York Times developer API.](https://developer.nytimes.com/docs/most-popular-product/1/overview)

Register for a free NYTimes developer account, create an API key, then fetch article data:

```bash
wget --directory-prefix=data/articles/nyt/article_versions -O data/articles/nyt/article_versions/data.json https://api.nytimes.com/svc/mostpopular/v2/viewed/1.json?api-key=yourkey
```

Then use the included Python script to convert article JSON data to RDF N-Quad format to import into Dgraph

First create a venv in the root of the project:

```bash
python3 -m venv venv
```

Then you can enter the venv and install the requirements:

Enter venv:

```bash
source venv/bin/activate
```

Then run the script to convert the JSON data to RDF:

```bash
cd ./data/articles && python3 article_json_to_rdf.py
```

This will output `nyt_articles_versions.rdf` in the `data/articles/` directory.

## Dgraph

Launch a local Dgraph cluster using Docker or create a free hosted [Hypermode Graph.](https://hypermode.com)

```bash
cd dgraph
docker-compose up
```

Then to open a shell in the Dgraph Alpha container:

```bash
docker exec -it news_graph_alpha /bin/bash
```

To load the RDF generated in the previous step:

```bash
dgraph live -f /data/articles/nyt_articles_versions.rdf --zero zero:5080
```

Open [ratel.hypermode.com](https://ratel.hypermode.com), connect to `http://localhost:8080` then query your graph:

```dql
{
  articles(func:type(Article),first:100) {
	Article.title
    Article.uri
    Article.url
    Article.published
    Article.abstract
    Article.topic {
      Topic.name
    }
    Article.org {
      Organization.name
    }
    Article.geo {
      Geo.name
    }
  }
}
```

![querying the graph using Ratel](img/ratel-query.png)

### Update Dgraph schema

```bash
curl -X POST localhost:8080/alter --data-binary '@schema.dql'
```

## Modus

Install Modus CLI

```bash
npm i -g @hypermode/modus-cli
```

```bash
modus dev
```

### Embeddings

![chunking the articles](img/chunks.png)

## Vector Similarity Search

```dql
 query vector_search($embedding: string, $limit: int) {
          articles(func: similar_to(Article.embedding, $limit, $embedding)) {
            uid
            Article.title
            Article.abstract
            score
          }
        }
```
