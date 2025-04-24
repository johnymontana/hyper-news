import json
import uuid
import os
import re
from geocode_ollama import geocode_location

def parse_byline(byline):
    """Extract author names from byline like 'By Author1, Author2 and Author3'"""
    if not byline or not isinstance(byline, str):
        return []
    
    # Remove "By " prefix
    if byline.startswith("By "):
        byline = byline[3:]
    
    # Split by commas and 'and'
    authors = re.split(r',\s*|\s+and\s+', byline)
    return [author.strip() for author in authors if author.strip()]

def trim_all_whitespace(text):
    no_whitespace = ''.join(text.split())
    clean_text = re.sub(r'[^a-zA-Z0-9]', '', no_whitespace)
    return clean_text

def json_to_nquads(articles):
    all_nquads = []
    
    for article in articles:
        article_uid = f"_:Article_{article['id']}"
        
        # Article basic info
        all_nquads.append(f'{article_uid} <dgraph.type> "Article" .')
        
        if "title" in article:
            all_nquads.append(f'{article_uid} <Article.title> "{escape_string(article["title"])}" .')
        
        if "abstract" in article:
            all_nquads.append(f'{article_uid} <Article.abstract> "{escape_string(article["abstract"])}" .')
        
        if "uri" in article:
            all_nquads.append(f'{article_uid} <Article.uri> "{escape_string(article["uri"])}" .')
        
        if "url" in article:
            all_nquads.append(f'{article_uid} <Article.url> "{escape_string(article["url"])}" .')
        
        if "published_date" in article:
            all_nquads.append(f'{article_uid} <Article.published> "{article["published_date"]}"^^<xs:dateTime> .')
        
        # Authors
        if "byline" in article:
            authors = parse_byline(article["byline"])
            for author_name in authors:
                author_uid = f"_:Author_{trim_all_whitespace(author_name)}"
                all_nquads.append(f'{author_uid} <dgraph.type> "Author" .')
                all_nquads.append(f'{author_uid} <Author.name> "{escape_string(author_name)}" .')
                all_nquads.append(f'{author_uid} <Author.article> {article_uid} .')
        
        # Topics (des_facet)
        if "des_facet" in article and isinstance(article["des_facet"], list):
            for topic in article["des_facet"]:
                topic_uid = f"_:Topic_{trim_all_whitespace(topic)}"
                all_nquads.append(f'{topic_uid} <dgraph.type> "Topic" .')
                all_nquads.append(f'{topic_uid} <Topic.name> "{escape_string(topic)}" .')
                all_nquads.append(f'{article_uid} <Article.topic> {topic_uid} .')
        
        # Organizations (org_facet)
        if "org_facet" in article and isinstance(article["org_facet"], list):
            for org in article["org_facet"]:
                org_uid = f"_:Organization_{trim_all_whitespace(org)}"
                all_nquads.append(f'{org_uid} <dgraph.type> "Organization" .')
                all_nquads.append(f'{org_uid} <Organization.name> "{escape_string(org)}" .')
                all_nquads.append(f'{article_uid} <Article.org> {org_uid} .')

        if "per_facet" in article and isinstance(article["per_facet"], list):
            for person in article["per_facet"]:
                person_uid = f"_:Person_{trim_all_whitespace(person)}"
                all_nquads.append(f'{person_uid} <dgraph.type> "Person" .')
                all_nquads.append(f'{person_uid} <Person.name> "{escape_string(person)}" .')
                all_nquads.append(f'{article_uid} <Article.person> {person_uid} .')
        
        # Geo locations (geo_facet)
        if "geo_facet" in article and isinstance(article["geo_facet"], list):
            for geo in article["geo_facet"]:

                try:
                    geojsonstr = geocode_location(geo)
                except Exception as e:
                    print(f"Error geocoding location '{geo}': {str(e)}")
                    continue

                geo_uid = f"_:Geo_{trim_all_whitespace(geo)}"
                all_nquads.append(f'{geo_uid} <dgraph.type> "Geo" .')
                all_nquads.append(f'{geo_uid} <Geo.name> "{escape_string(geo)}" .')

                if geojsonstr:
                    all_nquads.append(f'{geo_uid} <Geo.location> "{geojsonstr}"^^<geo:geojson> .')

                all_nquads.append(f'{article_uid} <Article.geo> {geo_uid} .')
        
        # Images
        if "media" in article and isinstance(article["media"], list):
            for media_item in article["media"]:
                if media_item.get("type") == "image":
                    image_uid = f"_:{uuid.uuid4()}"
                    all_nquads.append(f'{image_uid} <dgraph.type> "Image" .')
                    
                    if "caption" in media_item:
                        all_nquads.append(f'{image_uid} <Image.caption> "{escape_string(media_item["caption"])}" .')
                    
                    # Get the first image URL from metadata
                    if "media-metadata" in media_item and isinstance(media_item["media-metadata"], list) and len(media_item["media-metadata"]) > 0:
                        url = media_item["media-metadata"][0].get("url", "")
                        all_nquads.append(f'{image_uid} <Image.url> "{escape_string(url)}" .')
                    
                    all_nquads.append(f'{image_uid} <Image.article> {article_uid} .')
    
    return "\n".join(all_nquads)

def escape_string(s):
    """Escape special characters in strings for N-Quads"""
    if not isinstance(s, str):
        s = str(s)
    return s.replace('\\', '\\\\').replace('"', '\\"').replace('\n', '\\n').replace('\r', '\\r')

def process_json_files(directory):
    all_articles = []
    for filename in os.listdir(directory):
        if filename.endswith('.json'):
            with open(os.path.join(directory, filename), 'r') as f:
                try:
                    data = json.load(f)
                    # Handle both single articles and arrays of articles
                    if isinstance(data.get("results"), list):
                        all_articles.extend(data.get("results"))
                    elif isinstance(data, dict):
                        # Check if it's a container with articles
                        if "response" in data and "docs" in data["response"]:
                            all_articles.extend(data["response"]["docs"])
                        else:
                            # Single article
                            all_articles.append(data)
                except json.JSONDecodeError:
                    print(f"Error parsing {filename}")
    
    return all_articles

# Main process
def main(directory, output_file):
    articles = process_json_files(directory)
    nquads = json_to_nquads(articles)
    
    with open(output_file, 'w') as f:
        f.write(nquads)
    
    print(f"Processed {len(articles)} articles and wrote RDF to {output_file}")

if __name__ == "__main__":
    # Replace with your directory containing JSON files
    input_directory = "nyt/test"
    output_file = "nyt_articles_versions.rdf"
    main(input_directory, output_file)
