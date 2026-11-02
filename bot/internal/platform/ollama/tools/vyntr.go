package tools

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

type VyntrSearchArgs struct {
	SearchQuery string `json:"search_query"`
}

func GetVyntrSearchTool() openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		Function: shared.FunctionDefinitionParam{
			Name:        "search_web",
			Description: openai.String("Searches the web for up-to-date information"),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"search_query": map[string]any{
						"type":        "string",
						"description": "The query you want to search for",
					},
				},
				"required": []string{"search_query"},
			},
		},
	}
}

func VyntrSearch(apiKey string, query string) (string, error) {
	fmt.Println("Query:", query)

	urlToFetch := "https://vyntr.com/api/v1/search?q=" + url.QueryEscape(query)

	req, err := http.NewRequest("GET", urlToFetch, nil)
	if err != nil {
		fmt.Println("Error while creating new request:", err)
		return "", fmt.Errorf("failed to creqte request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error while fetching Vyntr API search results:", err)
		return "", fmt.Errorf("failed to fetch search results: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Error while fetching searching with Vyntr:", string(body))
		return "", fmt.Errorf("failed to search with vyntr: %v", string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))

	return string(body), nil
}
