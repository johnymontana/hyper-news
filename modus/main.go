package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hypermodeinc/modus/sdk/go/pkg/agents"
	"github.com/hypermodeinc/modus/sdk/go/pkg/console"
	"github.com/hypermodeinc/modus/sdk/go/pkg/dgraph"
	"github.com/hypermodeinc/modus/sdk/go/pkg/models"
	"github.com/hypermodeinc/modus/sdk/go/pkg/models/openai"
	"github.com/hypermodeinc/modus/sdk/go/pkg/utils"
)

var connection = "dgraph"

func init() {
	agents.Register(&HyperNewsChatAgent{})
}

func CreateConversation() (string, error) {
	info, err := agents.Start("HyperNewsChatAgent")
	if err != nil {
		return "", err
	}
	return info.Id, nil
}

func ContinueChat(id string, query string) (ChatResponse, error) {
	request := ChatRequest{
		Message: query,
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to marshal request: %v", err)
	}

	requestStr := string(requestData)
	response, err := agents.SendMessage(id, "chat", agents.WithData(requestStr))
	if err != nil {
		return ChatResponse{}, err
	}

	if response == nil {
		return ChatResponse{}, fmt.Errorf("no response received")
	}

	var agentResponse struct {
		Items          []interface{} `json:"items"`
		ConversationId string        `json:"conversationId"`
	}
	if err := json.Unmarshal([]byte(*response), &agentResponse); err != nil {
		return ChatResponse{}, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	itemsJson, err := json.Marshal(agentResponse.Items)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to marshal items: %v", err)
	}

	return ChatResponse{
		Items:          string(itemsJson),
		ConversationId: agentResponse.ConversationId,
	}, nil
}

func ChatHistory(id string) (HistoryResponse, error) {
	response, err := agents.SendMessage(id, "get_items")
	if err != nil {
		return HistoryResponse{}, err
	}

	if response == nil {
		return HistoryResponse{Items: "[]", Count: 0}, nil
	}

	var agentResponse struct {
		Items []interface{} `json:"items"`
		Count int           `json:"count"`
	}
	if err := json.Unmarshal([]byte(*response), &agentResponse); err != nil {
		return HistoryResponse{}, fmt.Errorf("failed to unmarshal items: %v", err)
	}

	itemsJson, err := json.Marshal(agentResponse.Items)
	if err != nil {
		return HistoryResponse{}, fmt.Errorf("failed to marshal items: %v", err)
	}

	return HistoryResponse{
		Items: string(itemsJson),
		Count: agentResponse.Count,
	}, nil
}

func DeleteAgent(id string) (string, error) {
	_, err := agents.Stop(id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func DeleteConversationHistory(id string) (bool, error) {
	_, err := agents.SendMessage(id, "clear_items")
	if err != nil {
		return false, err
	}
	return true, nil
}

// News query functions
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

	query := dgraph.NewQuery(`
	query vector_search($embedding: float32vector) {
		articles(func: similar_to(Article.embedding, 5, $embedding), orderdesc:Article.published,first:100) {
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

	b, err := json.MarshalIndent(articleData, "", "  ")
	if err != nil {
		fmt.Println("error marshaling articleData:", err)
	} else {
		console.Log(string(b))
	}

	return articleData.Articles, nil
}

func QueryLocations(lon float64, lat float64, distance int64) ([]*GeoData, error) {
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

	response, err := dgraph.ExecuteQuery(connection, query)
	if err != nil {
		return nil, err
	}

	var articleData ArticleData
	if err := json.Unmarshal([]byte(response.Json), &articleData); err != nil {
		return nil, err
	}

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
		people(func: type(Person)) {
			uid
			Person.name
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

func GeocodeLocation(location string) (*Coordinate, error) {
	sampleCoordinateJson, _ := utils.JsonSerialize(Coordinate{
		Latitude:  54.001,
		Longitude: -74.23904,
	})

	instruction := "I need the location for a given location. Only respond with valid JSON object in this format:\n" + string(sampleCoordinateJson)
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
