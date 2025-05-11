package main

import (
	"errors"
	"fmt"
	"strings"

	"encoding/json"

	"github.com/hypermodeinc/modus/sdk/go/pkg/console"
	"github.com/hypermodeinc/modus/sdk/go/pkg/dgraph"
	"github.com/hypermodeinc/modus/sdk/go/pkg/models"
	"github.com/hypermodeinc/modus/sdk/go/pkg/models/openai"
	"github.com/hypermodeinc/modus/sdk/go/pkg/utils"
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

type GeoSearch struct {
	Name     string     `json:"Geo.name,omitempty"`
	Articles []*Article `json:"articles"`
}

type GeoData struct {
	Geos []*GeoSearch `json:"geos"`
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

type ResponseWithLogs struct {
	Response string   `json:"response"`
	Logs     []string `json:"logs"`
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

	embeddingJson, err := json.Marshal(embedding)
	if err != nil {
		fmt.Println("error marshaling embedding:", err)
	} else {
		console.Log(string(embeddingJson))
	}

	//
	query := dgraph.NewQuery(`
	query vector_search($embedding: float32vector) {
		articles(func: type(Article), orderdesc:Article.published,first:100) {
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

	responseJson, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Println("error marshaling response:", err)
	} else {
		console.Log(string(responseJson))
	}

	var articleData ArticleData
	if err := json.Unmarshal([]byte(response.Json), &articleData); err != nil {
		return nil, err
	}
	// Print articleData as pretty JSON for debugging
	b, err := json.MarshalIndent(articleData, "", "  ")
	if err != nil {
		fmt.Println("error marshaling articleData:", err)
	} else {
		console.Log(string(b))
	}

	return articleData.Articles, nil
}

func QueryLocations(lon float64, lat float64, distance int64) ([]*GeoData, error) {

	// locationCoordinate, err := GeocodeLocation(location)
	// if err != nil {
	// 	return nil, err
	// }

	query := dgraph.NewQuery(fmt.Sprintf(`
	query NearbyLocations($distance: int){
  geos(func: near(Geo.location, [%f, %f], $distance)) {
    Geo.name
    articles: ~Article.geo {
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
	  }
    }
  }
}
	`, lon, lat)).WithVariable("$distance", distance)
	response, err := dgraph.ExecuteQuery(connection, query)
	if err != nil {
		return nil, err
	}

	var geoData GeoData
	if err := json.Unmarshal([]byte(response.Json), &geoData); err != nil {
		return nil, err
	}

	return []*GeoData{&geoData}, nil
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
	// Print articleData as pretty JSON for debugging
	b, err := json.MarshalIndent(articleData, "", "  ")
	if err != nil {
		fmt.Println("error marshaling articleData:", err)
	} else {
		fmt.Println(string(b))
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

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

var sampleCoordianteJson string = func() string {
	bytes, _ := utils.JsonSerialize(Coordinate{
		Latitude:  54.001,
		Longitude: -74.23904,
	})
	return string(bytes)
}()

func GeocodeLocation(location string) (*Coordinate, error) {
	instruction := "I need the location for a given location. Only respond with valid JSON object in this format:\n" + sampleCoordianteJson
	prompt := fmt.Sprintf(`The location is "%s".`, location)

	model, err := models.GetModel[openai.ChatModel]("text-generator")
	if err != nil {
		return nil, err
	}
	input, err := model.CreateInput(
		openai.NewSystemMessage(instruction),
		openai.NewUserMessage(prompt),
	)
	if err != nil {
		return nil, err
	}

	input.ResponseFormat = openai.ResponseFormatJson

	output, err := model.Invoke(input)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(output.Choices[0].Message.Content)

	var coordinate Coordinate
	if err := json.Unmarshal([]byte(content), &coordinate); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &coordinate, nil
}

func ExploreNewsWithTools(prompt string) (ResponseWithLogs, error) {
	var logs []string

	model, err := models.GetModel[openai.ChatModel]("text-generator")
	if err != nil {
		return ResponseWithLogs{}, err
	}

	instruction := `
	You are a helpful assistant who is very knowledgeable about recent news. Use your tools to answer the user's question.

	You may need to use multiple tools to resolve the users request, for example if they ask about news in a certain area you may need to first geocode the location and then query for news by location.
	
	`

	input, err := model.CreateInput(
		openai.NewSystemMessage(instruction),
		openai.NewUserMessage(prompt),
	)
	if err != nil {
		return ResponseWithLogs{}, err
	}

	input.Temperature = 0.2

	input.Tools = []openai.Tool{
		openai.NewToolForFunction("QueryArticles", "Queries news articles from the database, sorted by publication date returning the newest first.").WithParameter("num", "integer", "Number of articles to return"),
		openai.NewToolForFunction("QueryTopics", "Queries news topics from the database, sorted by publication date returning the newest first.").WithParameter("topic", "string", "Topic to search for"),
		openai.NewToolForFunction("QueryLocations", "Find news articles based on latitude and longitude location coordinates.").WithParameter("lon", "number", "Longitude of the location").WithParameter("lat", "number", "Latitude of the location").WithParameter("distance", "integer", "Distance in meters, recommend at least 50000"),
		openai.NewToolForFunction("GeocodeLocation", "Convert a location string to latitude and longitude coordinates.").WithParameter("location", "string", "Location string"),
	}

	for {
		output, err := model.Invoke(input)
		if err != nil {
			return ResponseWithLogs{}, err
		}

		msg := output.Choices[0].Message

		if len(msg.ToolCalls) > 0 {

			input.Messages = append(input.Messages, msg.ToAssistantMessage())

			for _, tc := range msg.ToolCalls {
				var toolMsg *openai.ToolMessage[string]

				logs = append(logs, fmt.Sprintf("Calling function : %s with %s",
					tc.Function.Name,
					tc.Function.Arguments))

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

				case "QueryLocations":
					lon := gjson.Get(tc.Function.Arguments, "lon").Float()
					lat := gjson.Get(tc.Function.Arguments, "lat").Float()
					distance := gjson.Get(tc.Function.Arguments, "distance").Int()
					if result, err := QueryLocations(lon, lat, distance); err == nil {
						toolMsg = openai.NewToolMessage(result, tc.Id)
					} else {
						toolMsg = openai.NewToolMessage(err, tc.Id)
					}

				case "GeocodeLocation":
					location := gjson.Get(tc.Function.Arguments, "location").String()
					if result, err := GeocodeLocation(location); err == nil {
						toolMsg = openai.NewToolMessage(result, tc.Id)
					} else {
						toolMsg = openai.NewToolMessage(err, tc.Id)
					}

				default:
					return ResponseWithLogs{}, fmt.Errorf("unknown tool call: %s", tc.Function.Name)
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

			return ResponseWithLogs{
				Response: content,
				Logs:     logs,
			}, nil
		} else {
			return ResponseWithLogs{}, errors.New("invalid response from model")
		}
	}
}
