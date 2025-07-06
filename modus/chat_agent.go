package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hypermodeinc/modus/sdk/go/pkg/agents"
	"github.com/hypermodeinc/modus/sdk/go/pkg/dgraph"
	"github.com/hypermodeinc/modus/sdk/go/pkg/models"
	"github.com/hypermodeinc/modus/sdk/go/pkg/models/openai"
)

const (
	MODEL_NAME      = "text-generator"
	MAX_HISTORY     = 20
	TOOL_LOOP_LIMIT = 3
)

// Chat agent implementation for HyperNews
type HyperNewsChatAgent struct {
	agents.AgentBase
	conversationId string
	items          []interface{}
	chatHistory    []openai.RequestMessage
	lastActivity   time.Time
}

func (c *HyperNewsChatAgent) Name() string {
	return "HyperNewsChatAgent"
}

func (c *HyperNewsChatAgent) GetState() *string {
	state := ChatAgentState{
		ConversationId: c.conversationId,
		Items:          c.items,
		ChatHistory:    c.chatHistory,
		LastActivity:   c.lastActivity,
	}

	data, err := json.Marshal(state)
	if err != nil {
		fmt.Printf("Error marshaling state: %v\n", err)
		return nil
	}

	stateStr := string(data)
	return &stateStr
}

func (c *HyperNewsChatAgent) SetState(data *string) {
	if data == nil {
		return
	}

	var state ChatAgentState
	if err := json.Unmarshal([]byte(*data), &state); err != nil {
		fmt.Printf("Error unmarshaling state: %v\n", err)
		return
	}

	c.conversationId = state.ConversationId
	c.items = state.Items
	c.chatHistory = state.ChatHistory
	c.lastActivity = state.LastActivity
}

func (c *HyperNewsChatAgent) OnInitialize() error {
	c.lastActivity = time.Now()
	c.chatHistory = []openai.RequestMessage{}
	return nil
}

func (c *HyperNewsChatAgent) OnSuspend() error {
	return nil
}

func (c *HyperNewsChatAgent) OnResume() error {
	return nil
}

func (c *HyperNewsChatAgent) OnTerminate() error {
	return nil
}

func (c *HyperNewsChatAgent) OnReceiveMessage(msgName string, data *string) (*string, error) {
	switch msgName {
	case "chat":
		return c.handleChat(data)
	case "get_items":
		return c.getConversationItems()
	case "clear_items":
		return c.clearConversationItems()
	default:
		return nil, fmt.Errorf("unknown message type: %s", msgName)
	}
}

func (c *HyperNewsChatAgent) handleChat(data *string) (*string, error) {
	if data == nil {
		return nil, fmt.Errorf("no message data provided")
	}

	var request ChatRequest
	if err := json.Unmarshal([]byte(*data), &request); err != nil {
		return nil, fmt.Errorf("failed to parse chat request: %v", err)
	}

	if c.conversationId == "" {
		c.conversationId = fmt.Sprintf("conv_%d", time.Now().UnixNano())
	}

	// Add user message to items and chat history
	userMessage := MessageItem{
		ResponseItem: ResponseItem{
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			Type:      ResponseTypeMessage,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Content: request.Message,
		Role:    "user",
	}
	c.items = append(c.items, userMessage)
	c.chatHistory = append(c.chatHistory, openai.NewUserMessage(request.Message))
	c.lastActivity = time.Now()

	var responseItems []interface{}

	// Generate AI response with tools
	response, toolItems, err := c.generateAIResponseWithTools(request.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI response: %v", err)
	}

	// Add tool call items to response
	for _, item := range toolItems {
		c.items = append(c.items, item)
		responseItems = append(responseItems, item)
	}

	// Add assistant message
	if response != "" {
		assistantMessage := MessageItem{
			ResponseItem: ResponseItem{
				ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
				Type:      ResponseTypeMessage,
				Timestamp: time.Now().Format(time.RFC3339),
			},
			Content: response,
			Role:    "assistant",
		}
		c.items = append(c.items, assistantMessage)
		responseItems = append(responseItems, assistantMessage)
	}

	// Limit chat history size
	if len(c.chatHistory) > MAX_HISTORY {
		c.chatHistory = c.chatHistory[len(c.chatHistory)-MAX_HISTORY:]
	}

	// Create response
	chatResponse := struct {
		Items          []interface{} `json:"items"`
		ConversationId string        `json:"conversationId"`
	}{
		Items:          responseItems,
		ConversationId: c.conversationId,
	}

	responseData, err := json.Marshal(chatResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %v", err)
	}

	responseStr := string(responseData)
	return &responseStr, nil
}

func (c *HyperNewsChatAgent) generateAIResponseWithTools(userMessage string) (string, []interface{}, error) {
	model, err := models.GetModel[openai.ChatModel](MODEL_NAME)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get model: %v", err)
	}

	tools := c.getNewsTools()
	systemPrompt := c.getSystemPrompt()

	var toolItems []interface{}
	loops := 0

	// Create a working copy of chat history for this conversation
	workingHistory := make([]openai.RequestMessage, len(c.chatHistory))
	copy(workingHistory, c.chatHistory)

	for loops < TOOL_LOOP_LIMIT {
		input, err := model.CreateInput()
		if err != nil {
			return "", nil, fmt.Errorf("failed to create input: %v", err)
		}

		// Build messages: system + history
		input.Messages = []openai.RequestMessage{openai.NewSystemMessage(systemPrompt)}
		input.Messages = append(input.Messages, workingHistory...)

		input.Temperature = 0.7
		input.Tools = tools
		input.ToolChoice = openai.ToolChoiceAuto

		output, err := model.Invoke(input)
		if err != nil {
			return "", nil, fmt.Errorf("model invocation failed: %v", err)
		}

		message := output.Choices[0].Message

		// Add assistant message to working history
		workingHistory = append(workingHistory, message.ToAssistantMessage())

		// Check if there are tool calls
		if len(message.ToolCalls) > 0 {
			// Process each tool call
			for _, toolCall := range message.ToolCalls {
				// Create tool call item for UI
				toolCallItem := ToolCallItem{
					ResponseItem: ResponseItem{
						ID:        fmt.Sprintf("tool_%d", time.Now().UnixNano()),
						Type:      ResponseTypeToolCall,
						Timestamp: time.Now().Format(time.RFC3339),
					},
					ToolCall: ToolCallData{
						ID:        toolCall.Id,
						Name:      toolCall.Function.Name,
						Arguments: c.parseToolArguments(toolCall.Function.Arguments),
						Status:    "executing",
					},
				}

				// Execute news tool
				result, err := c.executeNewsTool(toolCall)
				if err != nil {
					toolCallItem.ToolCall.Status = "error"
					toolCallItem.ToolCall.Error = err.Error()
				} else {
					toolCallItem.ToolCall.Status = "completed"
					toolCallItem.ToolCall.Result = result
				}

				toolItems = append(toolItems, toolCallItem)

				// Add tool response to working history
				var toolResponse string
				if err != nil {
					toolResponse = fmt.Sprintf("Error: %s", err.Error())
				} else {
					resultJSON, _ := json.Marshal(result)
					toolResponse = string(resultJSON)
				}
				workingHistory = append(workingHistory, openai.NewToolMessage(&toolResponse, toolCall.Id))
			}
		} else {
			// No more tool calls, we have our final response
			c.chatHistory = workingHistory
			return message.Content, toolItems, nil
		}

		loops++
	}

	// If we hit the loop limit, return what we have
	c.chatHistory = workingHistory
	return "I've processed your request with the available tools.", toolItems, nil
}

func (c *HyperNewsChatAgent) getNewsTools() []openai.Tool {
	return []openai.Tool{
		openai.NewToolForFunction("search_articles", "Search for news articles in the HyperNews database").
			WithParameter("query", "string", "Search query for articles").
			WithParameter("limit", "number", "Maximum number of articles to return (default: 5)"),

		openai.NewToolForFunction("get_article_by_id", "Get a specific article by its ID").
			WithParameter("article_id", "string", "The ID of the article to retrieve"),

		openai.NewToolForFunction("analyze_topics", "Analyze trending topics in recent articles").
			WithParameter("days", "number", "Number of days to look back (default: 7)").
			WithParameter("limit", "number", "Maximum number of topics to return (default: 10)"),

		openai.NewToolForFunction("get_articles_by_location", "Find articles related to a specific location").
			WithParameter("location", "string", "Geographic location to search for").
			WithParameter("limit", "number", "Maximum number of articles to return (default: 5)"),

		openai.NewToolForFunction("get_articles_by_organization", "Find articles mentioning specific organizations").
			WithParameter("organization", "string", "Organization name to search for").
			WithParameter("limit", "number", "Maximum number of articles to return (default: 5)"),

		openai.NewToolForFunction("summarize_article", "Generate a summary of an article").
			WithParameter("article_id", "string", "The ID of the article to summarize"),
	}
}

func (c *HyperNewsChatAgent) getSystemPrompt() string {
	return fmt.Sprintf(`Today is %s. You are HyperNews Assistant, an AI helper for exploring and analyzing news content.

You have access to a comprehensive news database with articles, topics, organizations, people, and locations. 

You can help users:
- Search for specific news articles
- Analyze trending topics and themes
- Find articles by location or organization
- Provide summaries and analysis
- Answer questions about current events

When users ask about news, always use the appropriate tools to search the database and provide accurate, up-to-date information. Create informative cards when displaying article information to make the content more engaging and actionable.

Be helpful, informative, and focus on providing valuable insights about the news content.`,
		time.Now().UTC().Format(time.RFC3339))
}

func (c *HyperNewsChatAgent) parseToolArguments(argsJSON string) map[string]interface{} {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return map[string]interface{}{"raw": argsJSON}
	}
	return args
}

func (c *HyperNewsChatAgent) executeNewsTool(toolCall openai.ToolCall) (interface{}, error) {
	var args map[string]interface{}
	json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

	switch toolCall.Function.Name {
	case "search_articles":
		return c.searchArticles(args)
	case "get_article_by_id":
		return c.getArticleById(args)
	case "analyze_topics":
		return c.analyzeTopics(args)
	case "get_articles_by_location":
		return c.getArticlesByLocation(args)
	case "get_articles_by_organization":
		return c.getArticlesByOrganization(args)
	case "summarize_article":
		return c.summarizeArticle(args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
	}
}

func (c *HyperNewsChatAgent) searchArticles(args map[string]interface{}) (interface{}, error) {
	query := c.getStringArg(args, "query", "")
	limit := c.getIntArg(args, "limit", 5)

	// Build vector similarity search query
	dqlQuery := fmt.Sprintf(`
	query search_articles($query: string, $limit: int) {
		articles(func: anyoftext(Article.title Article.abstract, $query), first: $limit) {
			uid
			Article.title
			Article.abstract
			Article.url
			Article.published
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
	}`)

	dgraphQuery := dgraph.NewQuery(dqlQuery).
		WithVariable("$query", query).
		WithVariable("$limit", limit)

	response, err := dgraph.ExecuteQuery(connection, dgraphQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search articles: %v", err)
	}

	var result struct {
		Articles []*Article `json:"articles"`
	}
	if err := json.Unmarshal([]byte(response.Json), &result); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %v", err)
	}

	// Create article cards for the results
	if len(result.Articles) > 0 {
		articlesCard := CardItem{
			ResponseItem: ResponseItem{
				ID:        fmt.Sprintf("card_%d", time.Now().UnixNano()),
				Type:      ResponseTypeCard,
				Timestamp: time.Now().Format(time.RFC3339),
			},
			Card: CardData{
				ID:    fmt.Sprintf("articles_card_%d", time.Now().UnixNano()),
				Type:  "articles",
				Title: fmt.Sprintf("Found %d articles for \"%s\"", len(result.Articles), query),
				Content: map[string]interface{}{
					"query":         query,
					"results_count": len(result.Articles),
					"articles":      result.Articles,
				},
				Actions: []CardAction{
					{
						ID:     "search_more",
						Label:  "Search more articles",
						Type:   "button",
						Action: "search_articles",
						Data:   map[string]interface{}{"query": query},
					},
				},
			},
		}
		c.items = append(c.items, articlesCard)
	}

	return map[string]interface{}{
		"query":          query,
		"articles_found": len(result.Articles),
		"articles":       result.Articles,
	}, nil
}

func (c *HyperNewsChatAgent) getArticleById(args map[string]interface{}) (interface{}, error) {
	articleId := c.getStringArg(args, "article_id", "")
	if articleId == "" {
		return nil, fmt.Errorf("article_id is required")
	}

	dqlQuery := `
	query get_article($id: string) {
		article(func: uid($id)) {
			uid
			Article.title
			Article.abstract
			Article.url
			Article.published
			Article.topic {
				Topic.name
			}
			Article.org {
				Organization.name
			}
			Article.geo {
				Geo.name
			}
			Article.person {
				Person.name
			}
		}
	}`

	dgraphQuery := dgraph.NewQuery(dqlQuery).WithVariable("$id", articleId)
	response, err := dgraph.ExecuteQuery(connection, dgraphQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %v", err)
	}

	var result struct {
		Article []*Article `json:"article"`
	}
	if err := json.Unmarshal([]byte(response.Json), &result); err != nil {
		return nil, fmt.Errorf("failed to parse article: %v", err)
	}

	if len(result.Article) == 0 {
		return nil, fmt.Errorf("article not found")
	}

	article := result.Article[0]

	// Create detailed article card
	articleCard := CardItem{
		ResponseItem: ResponseItem{
			ID:        fmt.Sprintf("card_%d", time.Now().UnixNano()),
			Type:      ResponseTypeCard,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Card: CardData{
			ID:    fmt.Sprintf("article_card_%d", time.Now().UnixNano()),
			Type:  "article_detail",
			Title: article.Title,
			Content: map[string]interface{}{
				"abstract":      article.Abstract,
				"url":           article.Url,
				"published":     article.Published,
				"topics":        article.Topics,
				"organizations": article.Organizations,
				"locations":     article.Geos,
				"people":        article.People,
			},
			Actions: []CardAction{
				{
					ID:     "view_article",
					Label:  "Read Full Article",
					Type:   "link",
					Action: article.Url,
				},
				{
					ID:     "summarize",
					Label:  "Get Summary",
					Type:   "button",
					Action: "summarize_article",
					Data:   map[string]interface{}{"article_id": articleId},
				},
			},
		},
	}
	c.items = append(c.items, articleCard)

	return article, nil
}

func (c *HyperNewsChatAgent) analyzeTopics(args map[string]interface{}) (interface{}, error) {
	days := c.getIntArg(args, "days", 7)
	limit := c.getIntArg(args, "limit", 10)

	dqlQuery := `
	query analyze_topics($limit: int) {
		topics(func: type(Topic), first: $limit) {
			Topic.name
			article_count: count(~Article.topic)
		}
	}`

	dgraphQuery := dgraph.NewQuery(dqlQuery).WithVariable("$limit", limit)
	response, err := dgraph.ExecuteQuery(connection, dgraphQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze topics: %v", err)
	}

	var result struct {
		Topics []struct {
			Name         string `json:"Topic.name"`
			ArticleCount int    `json:"article_count"`
		} `json:"topics"`
	}
	if err := json.Unmarshal([]byte(response.Json), &result); err != nil {
		return nil, fmt.Errorf("failed to parse topics: %v", err)
	}

	// Create topics analysis card
	topicsCard := CardItem{
		ResponseItem: ResponseItem{
			ID:        fmt.Sprintf("card_%d", time.Now().UnixNano()),
			Type:      ResponseTypeCard,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Card: CardData{
			ID:    fmt.Sprintf("topics_card_%d", time.Now().UnixNano()),
			Type:  "topics_analysis",
			Title: fmt.Sprintf("Top %d Topics (Last %d days)", len(result.Topics), days),
			Content: map[string]interface{}{
				"days":   days,
				"topics": result.Topics,
			},
			Actions: []CardAction{
				{
					ID:     "analyze_more",
					Label:  "Analyze More Topics",
					Type:   "button",
					Action: "analyze_topic",
					Data:   map[string]interface{}{"days": days * 2},
				},
			},
		},
	}
	c.items = append(c.items, topicsCard)

	return map[string]interface{}{
		"days":         days,
		"topics_found": len(result.Topics),
		"topics":       result.Topics,
	}, nil
}

func (c *HyperNewsChatAgent) getArticlesByLocation(args map[string]interface{}) (interface{}, error) {
	location := c.getStringArg(args, "location", "")
	limit := c.getIntArg(args, "limit", 5)

	dqlQuery := `
	query articles_by_location($location: string, $limit: int) {
		articles(func: type(Article)) @filter(anyoftext(Article.geo, $location)) {
			uid
			Article.title
			Article.abstract
			Article.url
			Article.published
			Article.geo {
				Geo.name
			}
		}
	}`

	dgraphQuery := dgraph.NewQuery(dqlQuery).
		WithVariable("$location", location).
		WithVariable("$limit", limit)

	response, err := dgraph.ExecuteQuery(connection, dgraphQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles by location: %v", err)
	}

	var result struct {
		Articles []*Article `json:"articles"`
	}
	if err := json.Unmarshal([]byte(response.Json), &result); err != nil {
		return nil, fmt.Errorf("failed to parse location articles: %v", err)
	}

	return map[string]interface{}{
		"location":       location,
		"articles_found": len(result.Articles),
		"articles":       result.Articles,
	}, nil
}

func (c *HyperNewsChatAgent) getArticlesByOrganization(args map[string]interface{}) (interface{}, error) {
	organization := c.getStringArg(args, "organization", "")
	limit := c.getIntArg(args, "limit", 5)

	dqlQuery := `
	query articles_by_org($org: string, $limit: int) {
		articles(func: type(Article)) @filter(anyoftext(Article.org, $org)) {
			uid
			Article.title
			Article.abstract
			Article.url
			Article.published
			Article.org {
				Organization.name
			}
		}
	}`

	dgraphQuery := dgraph.NewQuery(dqlQuery).
		WithVariable("$org", organization).
		WithVariable("$limit", limit)

	response, err := dgraph.ExecuteQuery(connection, dgraphQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles by organization: %v", err)
	}

	var result struct {
		Articles []*Article `json:"articles"`
	}
	if err := json.Unmarshal([]byte(response.Json), &result); err != nil {
		return nil, fmt.Errorf("failed to parse organization articles: %v", err)
	}

	return map[string]interface{}{
		"organization":   organization,
		"articles_found": len(result.Articles),
		"articles":       result.Articles,
	}, nil
}

func (c *HyperNewsChatAgent) summarizeArticle(args map[string]interface{}) (interface{}, error) {
	articleId := c.getStringArg(args, "article_id", "")
	if articleId == "" {
		return nil, fmt.Errorf("article_id is required")
	}

	// First get the article
	article, err := c.getArticleById(args)
	if err != nil {
		return nil, err
	}

	// Generate summary using the AI model
	model, err := models.GetModel[openai.ChatModel](MODEL_NAME)
	if err != nil {
		return nil, fmt.Errorf("failed to get model for summarization: %v", err)
	}

	articleData := article.(*Article)

	input, err := model.CreateInput(
		openai.NewSystemMessage("You are a helpful assistant that creates concise, informative summaries of news articles. Focus on the key points, main themes, and important details."),
		openai.NewUserMessage(fmt.Sprintf("Please summarize this article:\n\nTitle: %s\n\nAbstract: %s\n\nCreate a 2-3 sentence summary focusing on the most important points.", articleData.Title, articleData.Abstract)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create summarization input: %v", err)
	}

	input.Temperature = 0.3

	output, err := model.Invoke(input)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %v", err)
	}

	summary := output.Choices[0].Message.Content

	return map[string]interface{}{
		"article_id": articleId,
		"title":      articleData.Title,
		"summary":    summary,
		"url":        articleData.Url,
	}, nil
}

func (c *HyperNewsChatAgent) getStringArg(args map[string]interface{}, key, defaultValue string) string {
	if val, ok := args[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (c *HyperNewsChatAgent) getIntArg(args map[string]interface{}, key string, defaultValue int) int {
	if val, ok := args[key]; ok {
		if num, ok := val.(float64); ok {
			return int(num)
		}
		if num, ok := val.(int); ok {
			return num
		}
	}
	return defaultValue
}

func (c *HyperNewsChatAgent) getConversationItems() (*string, error) {
	response := struct {
		Items []interface{} `json:"items"`
		Count int           `json:"count"`
	}{
		Items: c.items,
		Count: len(c.items),
	}

	itemsData, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal items: %v", err)
	}

	itemsStr := string(itemsData)
	return &itemsStr, nil
}

func (c *HyperNewsChatAgent) clearConversationItems() (*string, error) {
	c.items = []interface{}{}
	c.chatHistory = []openai.RequestMessage{}
	c.lastActivity = time.Now()
	return nil, nil
}
