package main

import (
	"encoding/json"
	"fmt"

	_ "github.com/hypermodeinc/modus/sdk/go"
	"github.com/hypermodeinc/modus/sdk/go/pkg/console"
	"github.com/hypermodeinc/modus/sdk/go/pkg/http"
)

type Article struct {
	ID       string `json:"uid"`
	Title    string `json:"Article.title,omitempty"`
	Abstract string `json:"Article.abstract,omitempty"`
	URL      string `json:"Article.url,omitempty"`
	URI      string `json:"Article.uri,omitempty"`
}

func GetArticles() (string, error) {
	query := `{
		articles(func: type(Article)) {
			uid
			Article.title
			Article.abstract
			Article.url
			Article.uri
		}
	}`
	
	return executeDgraphQuery(query)
}

func executeDgraphQuery(query string) (string, error) {
	queryPayload := map[string]string{"query": query}
	
	options := &http.RequestOptions{
		Method: "POST",
		Body:   queryPayload,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	
	request := http.NewRequest("http://localhost:8080/query", options)
	response, err := http.Fetch(request)
	if err != nil {
		console.Log("Error fetching data from Dgraph: " + err.Error())
		return "", err
	}
	
	if !response.Ok() {
		return "", fmt.Errorf("Dgraph query failed: %d %s", response.Status, response.StatusText)
	}
	
	return string(response.Body), nil
}

func ParseArticles(jsonResponse string) ([]Article, error) {
	var result struct {
		Data struct {
			Articles []Article `json:"articles"`
		} `json:"data"`
	}
	
	err := json.Unmarshal([]byte(jsonResponse), &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing articles JSON: %w", err)
	}
	
	return result.Data.Articles, nil
}