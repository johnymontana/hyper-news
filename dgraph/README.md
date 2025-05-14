# Dgraph

To update the schema using the schema.dql file:

```bash
curl -X POST https://<YOUR_DGRAPH_URI>/dgraph/schema \
    --header "Authorization: Bearer {{TOKEN}}" \
    --header "Content-Type: application/dql" \
    --data @schema.dql
```

To delete all data in your Dgraph instance (but keep the schema):

```bash
curl -X POST https://<YOUR_DGRAPH_URI>/dgraph/alter \
    --header "Authorization: Bearer {{TOKEN}}" \
    --header "Content-Type: application/json"  \
    --data '{"drop_op": "DATA"}'
```

## Introduction to DQL (Dgraph Query Language)

DQL is Dgraph's query language that allows you to retrieve and manipulate data in a graph-like fashion. This guide introduces DQL concepts in order of increasing complexity.

### Basic Queries

The simplest DQL query retrieves nodes of a certain type:

```graphql
{
  articles(func: type(Article)) {
    Article.title
    Article.abstract
    Article.uri
  }
}
```

This query fetches all Article nodes and returns their title, abstract, and URI.

### Filtering and Ordering

You can filter results using the `@filter` directive:

```graphql
{
  articles(func: type(Article)) @filter(has(Article.abstract)) {
    Article.title
    Article.abstract
  }
}
```

This returns only articles that have an abstract.

To order results, use the `orderasc` or `orderdesc` parameter:

```graphql
{
  articles(func: type(Article), orderasc: Article.title) {
    Article.title
    Article.abstract
  }
}
```

**Schema Improvement:** Add an `@index` to `Article.title` to enable fast sorting:
```
<Article.title>: string @index(exact) .
```

### Date Filtering

Your schema includes `Article.published` as a datetime field. To filter by date:

```graphql
{
  recent_articles(func: type(Article)) @filter(ge(Article.published, "2025-01-01T00:00:00Z")) {
    Article.title
    Article.published
  }
}
```

**Schema Improvement:** Add a datetime index for faster date-based queries:
```
<Article.published>: datetime @index(hour) .
```

### Nested Traversals

Follow relationships between entities with nested queries:

```graphql
{
  topics(func: type(Topic)) {
    Topic.name
    ~Article.topic {  # Traverse reverse edge to articles
      Article.title
      Article.abstract
    }
  }
}
```

You can also query articles and include their related entities:

```graphql
{
  articles(func: type(Article)) {
    Article.title
    Article.topic {
      Topic.name
    }
    Article.org {
      Organization.name
    }
  }
}
```

### Full-Text Search

The schema has a full-text index on `Topic.name`, enabling text search:

```graphql
{
  topics(func: anyoftext(Topic.name, "technology AI")) {
    Topic.name
    ~Article.topic {
      Article.title
    }
  }
}
```

**Schema Improvement:** Add full-text search to Article titles and abstracts:
```
<Article.title>: string @index(fulltext) .
<Article.abstract>: string @index(fulltext, term) .
```

### Geospatial Queries

Your schema has `Geo.location` as a geo field, enabling location-based queries:

```graphql
{
  nearby_locations(func: near(Geo.location, [-74.0060, 40.7128], 50000)) {
    Geo.name
    Geo.location
    ~Article.geo {
      Article.title
    }
  }
}
```

This finds locations within 10km of New York City coordinates and their associated articles.

### Vector Similarity Search

The schema includes `Article.embedding` with an HNSW vector index, allowing semantic searches:

```graphql
query vector_search($embedding: string, $limit: int) {
          articles(func: similar_to(Article.embedding, $limit, $embedding)) {
            uid
            Article.title
            Article.abstract
            score
          }
        }
```

This finds the 5 articles with embeddings most similar to the given vector.

### Advanced Queries: Combining Multiple Filters

Combine multiple filters for complex queries:

```graphql
query vector_search($embedding: string, $limit: int) {
          articles(func: similar_to(Article.embedding, $limit, $embedding)) @filter(
            anyoftext(Article.abstract, "technology AI") AND
            ge(Article.published, "2025-01-01") AND
            has(Article.geo)
          ) {
            uid
            Article.title
            Article.abstract
            score
          }
        }
}
```

This finds the 5 articles with embeddings most similar to the given vector.

### Advanced Queries: Combining Multiple Filters

Combine multiple filters for complex queries:

```graphql
{
  tech_articles_2025(func: type(Article)) @filter(
    anyoftext(Article.abstract, "technology AI") AND
    ge(Article.published, "2025-01-01") AND
    has(Article.geo)
  ) {
    Article.title
    Article.abstract
    Article.published
    Article.geo {
      Geo.name
      Geo.location
    }
    Article.topic {
      Topic.name
    }
  }
}
```

### Additional Schema Improvements

To enable more advanced queries, consider these improvements:

1. Add indexes to organization and author names for searching:
   ```
   <Organization.name>: string @index(exact, term) .
   <Author.name>: string @index(exact, term) .
   ```

2. Add count indexing to quickly count relationships:
   ```
   <Article.topic>: [uid] @count @reverse .
   <Author.article>: [uid] @count @reverse .
   ```

3. Add unique ID constraints for article URIs:
   ```
   <Article.uri>: string @index(exact) @upsert .
   ```

4. Add date partitioning for more efficient date range queries:
   ```
   <Article.published>: datetime @index(year, month, day, hour) .
   ```

These enhancements will provide more query capabilities without requiring changes to your data model.

### Client Directives

DQL offers several client-side directives that modify query behavior without affecting the underlying data.

#### @cascade

The `@cascade` directive filters out nodes where any of the requested fields are null or empty:

```graphql
{
  articles(func: type(Article)) @cascade {
    Article.title
    Article.abstract
    Article.topic {
      Topic.name
    }
  }
}
```

This returns only articles that have all three fields: title, abstract, and at least one topic.

#### @normalize

The `@normalize` directive flattens nested data into a simpler structure:

```graphql
{
  articles(func: type(Article)) @normalize {
    title: Article.title
    topics: Article.topic {
      name: Topic.name
    }
    authors: Article.author {
      name: Author.name
    }
  }
}
```

This returns results with flattened, aliased field names for easier processing.

#### @facets

While not currently configured in your schema, facets let you add metadata to edges. To add and query facets, you'd update your schema like this:

```
<Article.topic>: [uid] @reverse @facets(relevance: float) .
```

Then query with:

```graphql
{
  articles(func: type(Article)) {
    Article.title
    Article.topic @facets(relevance) {
      Topic.name
    }
  }
}
```

#### @filter (with multiple conditions)

Combine multiple filter conditions using logical operators:

```graphql
{
  articles(func: type(Article)) @filter(has(Article.abstract) AND (anyoftext(Article.abstract, "climate") OR anyoftext(Article.abstract, "weather"))) {
    Article.title
    Article.abstract
  }
}
```

#### @recurse

For recursive traversals (useful if your graph has hierarchical relationships):

```graphql
{
  topics(func: type(Topic)) {
    Topic.name
    subtopics @recurse(depth: 3) {
      name
      subtopics
    }
  }
}
```

Note: This would require adding a self-referential `subtopics` predicate to your schema.

### Aggregation Queries

DQL provides functions for aggregating data:

#### Basic Count

```graphql
{
  total_articles(func: type(Article)) {
    count(uid)
  }
}
```

#### Count with Grouping

```graphql
{
  topics(func: type(Topic)) {
    Topic.name
    article_count: count(~Article.topic)
  }
}
```

This counts how many articles are associated with each topic.

#### Multiple Aggregations

```graphql
{
  articles(func: type(Article)) {
    topic_stats: Article.topic {
      # Requires @index(exact) on Topic.name
      topic_min: min(Topic.name)
      topic_max: max(Topic.name)
      topic_count: count(uid)
    }
  }
}
```

#### Value-Based Aggregations

For numeric fields with appropriate indexes (not in your current schema):

```graphql
{
  # This would require adding a numeric wordCount field with an @index(int)
  article_stats(func: type(Article)) {
    min_words: min(Article.wordCount)
    max_words: max(Article.wordCount)
    avg_words: avg(Article.wordCount)
    sum_words: sum(Article.wordCount)
  }
}
```

#### Grouping with @groupby

Group and aggregate data (requires adding @index directives to the fields used in groupby):

```graphql
{
  articles(func: type(Article)) @groupby(Article.published) {
    month: min(Article.published)
    count: count(uid)
  }
}
```

Note: This would require `<Article.published>: datetime @index(month)` in the schema.

#### Date-Based Aggregations

```graphql
{
  publications_by_month(func: type(Article)) {
    count: count(uid)
    month: datetrunc(Article.published, "month")
  } @groupby(month)
}
```

Note: This requires the proper datetime index on Article.published.

### Combined Advanced Example

This example combines multiple directives and aggregations:

```graphql
{
  topic_statistics(func: type(Topic)) @filter(has(~Article.topic)) {
    Topic.name
    articles: ~Article.topic @cascade {
      count: count(uid)
      recent_count: count(uid) @filter(ge(Article.published, "2025-01-01T00:00:00Z"))
      oldest: min(Article.published)
      newest: max(Article.published)
    }
  }
}
```

This returns each topic with article statistics, including total count, recent count, and publication date ranges.
