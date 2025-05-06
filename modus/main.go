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

// TODO: add property with minilm embedding value in rdf
// TODO: add modus function to embed with minilm
// TODO: add modus function to implement vector search using minilm embedding value - does it work?
// TODO: if not then push to github and write it up

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

func QueryLocations(lon float64, lat float64, distance int) ([]*GeoData, error) {

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

type Product struct {
	Id          string  `json:"id,omitempty"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

var sampleProductJson string = func() string {
	bytes, _ := utils.JsonSerialize(Product{
		Id:          "123",
		Name:        "Shoes",
		Price:       50.0,
		Description: "Great shoes for walking.",
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

// This function generates a single product.
func GenerateProduct(category string) (*Product, error) {

	// We can get creative with the instruction and prompt to guide the model
	// in generating the desired output.  Here we provide a sample JSON of the
	// object we want the model to generate.
	instruction := "Generate a product for the category provided.\n" +
		"Only respond with valid JSON object in this format:\n" + sampleProductJson
	prompt := fmt.Sprintf(`The category is "%s".`, category)

	// Set up the input for the model, creating messages for the instruction and prompt.
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

	// Let's increase the temperature to get more creative responses.
	// Be careful though, if the temperature is too high, the model may generate invalid JSON.
	input.Temperature = 1.2

	// This model also has a response format parameter that can be set to JSON,
	// Which, along with the instruction, can help guide the model in generating valid JSON output.
	input.ResponseFormat = openai.ResponseFormatJson

	// Here we invoke the model with the input we created.
	output, err := model.Invoke(input)
	if err != nil {
		return nil, err
	}

	// The output should contain the JSON string we asked for.
	content := strings.TrimSpace(output.Choices[0].Message.Content)

	// We can now parse the JSON string as a Product object.
	var product Product
	if err := json.Unmarshal([]byte(content), &product); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &product, nil
}

// This function generates multiple product.
func GenerateProducts(category string, quantity int) ([]Product, error) {

	// Similar to the generateText example, we can tailor the instruction and prompt
	// to guide the model in generating the desired output.  Note that understanding the behavior
	// of the model is important to get the desired results.  In this case, we need the model
	// to return an _object_ containing an array, not an array of objects directly.
	// That's because the model will not reliably generate an array of objects directly.
	instruction := fmt.Sprintf("Generate %d products for the category provided.\n"+
		"Only respond with a valid JSON object containing a valid JSON array named 'list', in this format:\n"+
		`{"list":[%s]}`, quantity, sampleProductJson)
	prompt := fmt.Sprintf(`The category is "%s".`, category)

	// Set up the input for the model, creating messages for the instruction and prompt.
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

	// Adjust the model inputs, just like in the previous example.
	// Be careful, if the temperature is too high, the model may generate invalid JSON.
	input.Temperature = 1.2
	input.ResponseFormat = openai.ResponseFormatJson

	// Here we invoke the model with the input we created.
	output, err := model.Invoke(input)
	if err != nil {
		return nil, err
	}

	// The output should contain the JSON string we asked for.
	content := strings.TrimSpace(output.Choices[0].Message.Content)

	// We can parse that JSON to a compatible object, to get the data we're looking for.
	var data map[string][]Product
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Now we can extract the list of products from the data.
	products, found := data["list"]
	if !found {
		return nil, fmt.Errorf("expected 'list' key in JSON object")
	}
	return products, nil
}

func GenerateTextWithTools(prompt string) (string, error) {
	model, err := models.GetModel[openai.ChatModel]("text-generator")
	if err != nil {
		return "", err
	}

	instruction := `
	You are a helpful assistant who is very knowledgeable about recent news. Use your tools to answer the user's question.
	
	`

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
		openai.NewToolForFunction("QueryLocations", "Find news articles based on location.").WithParameter("lon", "number", "Longitude of the location").WithParameter("lat", "number", "Latitude of the location").WithParameter("distance", "integer", "Distance in meters, recommend at least 50000"),
		openai.NewToolForFunction("GeocodeLocation", "Convert a location string to latitude and longitude coordinates.").WithParameter("location", "string", "Location string"),
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

				case "QueryLocations":
					topic := gjson.Get(tc.Function.Arguments, "topic").String()
					if result, err := QueryTopics(topic); err == nil {
						toolMsg = openai.NewToolMessage(result, tc.Id)
					} else {
						toolMsg = openai.NewToolMessage(err, tc.Id)
					}

				case "GeocodeLocation":
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
