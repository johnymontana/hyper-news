package main

import (
	"errors"
	"fmt"
	"strings"

	"encoding/json"

	_ "github.com/hypermodeinc/modus/sdk/go"
	"github.com/hypermodeinc/modus/sdk/go/pkg/dgraph"
	"github.com/hypermodeinc/modus/sdk/go/pkg/models"
	"github.com/hypermodeinc/modus/sdk/go/pkg/models/openai"
	"github.com/tidwall/gjson"
)

var connection = "dgraph"

type Person struct {
	Uid   string   `json:"uid,omitempty"`
	Name  string   `json:"Person.name,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

type Author struct {
	Uid  string `json:"uid,omitempty"`
	Name string `json:"Author.name,omitempty"`
}

type Article struct {
	Uid           string          `json:"uid,omitempty"`
	Title         string          `json:"Article.title,omitempty"`
	Abstract      string          `json:"Article.abstract,omitempty"`
	Url           string          `json:"Article.url,omitempty"`
	People        []*Person       `json:"Article.person,omitempty"`
	Authors       []*Author       `json:"Article.author,omitempty"`
	Organizations []*Organization `json:"Article.org,omitempty"`
	Topics        []*Topic        `json:"Article.topic,omitempty"`
	Geos          []*Geo          `json:"Article.geo,omitempty"`
}

type Geo struct {
	Uid      string `json:"uid,omitempty"`
	Name     string `json:"Geo.name,omitempty"`
	Location string `json:"Geo.location,omitempty"`
}

type Organization struct {
	Uid  string `json:"uid,omitempty"`
	Name string `json:"Organization.name,omitempty"`
}

type Topic struct {
	Uid  string `json:"uid,omitempty"`
	Name string `json:"Topic.name,omitempty"`
	// Articles []*Article `json:"Topic.article,omitempty"`
}

type TopicData struct {
	Topics []*SearchTopic `json:"topics"`
}

type SearchTopic struct {
	Uid      string     `json:"uid,omitempty"`
	Name     string     `json:"Topic.name,omitempty"`
	Articles []*Article `json:"Topic.article,omitempty"`
}

type ArticleData struct {
	Articles []*Article `json:"articles"`
}

type PeopleData struct {
	People []*Person `json:"people"`
}

func GetEmbeddingsForText(texts ...string) ([][]float32, error) {
	model, err := models.GetModel[openai.EmbeddingsModel]("nomic-embed")

	if err != nil {
		return nil, err
	}

	input, err := model.CreateInput(texts)
	if err != nil {
		return nil, err
	}

	output, err := model.Invoke(input)
	if err != nil {
		return nil, err
	}

	results := make([][]float32, len(output.Data))
	for i, d := range output.Data {
		results[i] = d.Embedding
	}

	return results, nil
}

func QuerySimilar(userQuery *string) ([]*Article, error) {
	// TODO: embed query and search

	embedding, err := GetEmbeddingsForText(*userQuery)
	if err != nil {
		return nil, err
	}

	query := dgraph.NewQuery(`
	query vector_search($embedding: string) {
		articles(func: similar_to(Article.embedding, 100, $embedding)) {
		  uid
			Article.title
			Article.abstract
			Article.url
			Article.geo {
				Geo.name
			}

			Article.org {
			Organization.name
			}

			Article.topic {
			Topic.name
			}


			dgraph.type
		}
	  }
	`).WithVariable("$embedding", embedding[0])

	response, err := dgraph.ExecuteQuery(connection, query)
	if err != nil {
		return nil, err
	}

	var articleData ArticleData
	if err := json.Unmarshal([]byte(response.Json), &articleData); err != nil {
		return nil, err
	}
	fmt.Printf("ArticleData: %+v\n", articleData)

	return articleData.Articles, nil
}

func QueryLocations(location *string) ([]*Article, error) {
	// TODO: add index to locations and search for them by fulltext
	return nil, nil
}

func QueryTopics(topic string) ([]*SearchTopic, error) {
	query := dgraph.NewQuery(`
	query queryTopics($topic: string!) {
  topics(func: anyoftext(Topic.name, $topic), first: 10) {
   Topic.name 
    uid
    Topic.article: ~Article.topic {
      Article.title
      Article.abstract
      Article.author: ~Author.article {
        Author.name
      }
      Article.org {
        Organization.name
      }
      Article.topic {
        Topic.name
      }
	Article.geo {
		Geo.name
		Geo.location
	  }
    }
  }
}
`).WithVariable("$topic", topic)

	response, err := dgraph.ExecuteQuery(connection, query)
	if err != nil {
		return nil, err
	}

	var topicData TopicData
	if err := json.Unmarshal([]byte(response.Json), &topicData); err != nil {
		return nil, err
	}

	return topicData.Topics, nil
}

// QueryArticles retrieves articles from the database
// If num is nil, it defaults to 10 articles
func QueryArticles(num int) ([]*Article, error) {

	query := dgraph.NewQuery(`

	query queryArticles($num: int!) {
		articles(func: type(Article), orderdesc:Article.published,first: $num) {
			uid
			Article.title
			Article.abstract
			Article.url
			
			Article.geo {
				Geo.name
			}

			Article.org {
			Organization.name
			}

			Article.topic {
			Topic.name
			}

			Article.person {
			Person.name
			}

			Article.author: ~Author.article {
			Author.name
			}

			dgraph.type
		}
	}
	`).WithVariable("$num", num)

	// Execute the query with variables
	response, err := dgraph.ExecuteQuery(connection, query)
	if err != nil {
		return nil, err
	}

	var articleData ArticleData
	if err := json.Unmarshal([]byte(response.Json), &articleData); err != nil {
		return nil, err
	}

	return articleData.Articles, nil
}

func QueryPeople() ([]*Person, error) {
	query := dgraph.NewQuery(`
	{
		people(func: has(dgraph.type, "Person")) {
			uid
			firstName
			lastName
			dgraph.type
		}
		}
	`)

	response, err := dgraph.ExecuteQuery(connection, query)
	if err != nil {
		return nil, err
	}

	var peopleData PeopleData
	if err := json.Unmarshal([]byte(response.Json), &peopleData); err != nil {
		return nil, err
	}

	return peopleData.People, nil
}

func GenerateTextWithTools(prompt string) (string, error) {
	model, err := models.GetModel[openai.ChatModel]("text-generator")
	if err != nil {
		return "", err
	}

	instruction := `
	You are a helpful assistant who is very knowledgeable about recent news. Use your tools to answer the user's question.
	
	Important: When returning articles, ALWAYS format your response as a valid JSON array of article objects with the following structure:
	[
		{
			"uid": "article-1",
			"title": "Article Title",
			"abstract": "Brief description of the article",
			"url": "https://full-url-to-article.com",
			"uri": "https://full-url-to-article.com"
		}
	]
	
	Your response MUST be valid JSON with no surrounding text or markdown formatting. Only include the JSON array of articles.`

	input, err := model.CreateInput(
		openai.NewSystemMessage(instruction),
		openai.NewUserMessage(prompt),
	)
	if err != nil {
		return "", err
	}

	input.Temperature = 0.2

	input.Tools = []openai.Tool{
		openai.NewToolForFunction("QueryArticles", "Queries news articles from the database, sorted by publication date returning the newest first.").WithParameter("num", "integer", "Number of articles to return"),
		openai.NewToolForFunction("QueryTopics", "Queries news topics from the database, sorted by publication date returning the newest first.").WithParameter("topic", "string", "Topic to search for"),
	}

	for {
		output, err := model.Invoke(input)
		if err != nil {
			return "", err
		}

		msg := output.Choices[0].Message

		if len(msg.ToolCalls) > 0 {
			input.Messages = append(input.Messages, msg.ToAssistantMessage())

			for _, tc := range msg.ToolCalls {
				var toolMsg *openai.ToolMessage[string]
				switch tc.Function.Name {
				case "QueryArticles":
					// Convert int64 to int and create a pointer to it
					numInt64 := gjson.Get(tc.Function.Arguments, "num").Int()
					numInt := int(numInt64)
					if result, err := QueryArticles(numInt); err == nil {
						toolMsg = openai.NewToolMessage(result, tc.Id)
					} else {
						toolMsg = openai.NewToolMessage(err, tc.Id)
					}

				case "QueryTopics":
					topic := gjson.Get(tc.Function.Arguments, "topic").String()
					if result, err := QueryTopics(topic); err == nil {
						toolMsg = openai.NewToolMessage(result, tc.Id)
					} else {
						toolMsg = openai.NewToolMessage(err, tc.Id)
					}

				default:
					return "", fmt.Errorf("unknown tool call: %s", tc.Function.Name)
				}

				input.Messages = append(input.Messages, toolMsg)
			}
		} else if msg.Content != "" {
			content := strings.TrimSpace(msg.Content)

			if strings.HasPrefix(content, "```json") {
				content = strings.TrimPrefix(content, "```json")
				if idx := strings.LastIndex(content, "```"); idx != -1 {
					content = content[:idx]
				}
				content = strings.TrimSpace(content)
			} else if strings.HasPrefix(content, "```") {
				content = strings.TrimPrefix(content, "```")
				if idx := strings.LastIndex(content, "```"); idx != -1 {
					content = content[:idx]
				}
				content = strings.TrimSpace(content)
			}

			return content, nil
		} else {
			return "", errors.New("invalid response from model")
		}
	}
}
