package main

import (
	"time"

	"github.com/hypermodeinc/modus/sdk/go/pkg/models/openai"
)

// Request/Response types for GraphQL API
type ChatRequest struct {
	Message string `json:"message"`
}

type ChatResponse struct {
	Items          string `json:"items"` // JSON string of items array
	ConversationId string `json:"conversationId"`
}

type HistoryResponse struct {
	Items string `json:"items"` // JSON string of items array
	Count int    `json:"count"`
}

// Response item types
type ResponseItemType string

const (
	ResponseTypeMessage  ResponseItemType = "message"
	ResponseTypeToolCall ResponseItemType = "tool_call"
	ResponseTypeCard     ResponseItemType = "card"
)

type ResponseItem struct {
	ID        string           `json:"id"`
	Type      ResponseItemType `json:"type"`
	Timestamp string           `json:"timestamp,omitempty"`
}

type MessageItem struct {
	ResponseItem
	Content string `json:"content"`
	Role    string `json:"role"`
}

type ToolCallItem struct {
	ResponseItem
	ToolCall ToolCallData `json:"toolCall"`
}

type CardItem struct {
	ResponseItem
	Card CardData `json:"card"`
}

type ToolCallData struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	Status    string                 `json:"status"`
	Result    interface{}            `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

type CardData struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Title   string                 `json:"title,omitempty"`
	Content map[string]interface{} `json:"content"`
	Actions []CardAction           `json:"actions,omitempty"`
}

type CardAction struct {
	ID     string                 `json:"id"`
	Label  string                 `json:"label"`
	Type   string                 `json:"type"`
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type ChatAgentState struct {
	ConversationId string                  `json:"conversationId"`
	Items          []interface{}           `json:"items"`
	ChatHistory    []openai.RequestMessage `json:"chatHistory"`
	LastActivity   time.Time               `json:"lastActivity"`
}

// Article and related data types
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
	Published     string          `json:"Article.published,omitempty"`
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
}

// Query result types
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

// Coordinate type for geocoding
type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
