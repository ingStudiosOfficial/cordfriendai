package tools

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
)

type VyntrSearchArgs struct {
	SearchQuery string `json:"search_query"`
}

func GetVyntrSearchTool() api.Tool {
	properties := api.NewToolPropertiesMap()

	properties.Set("search_query", api.ToolProperty{
		Type:        api.PropertyType{"string"},
		Description: "The query you want to search for",
	})

	var VyntrSearchTool = api.Tool {
		Type: "function",
		Function: api.ToolFunction{
			Name: "search_web",
			Description: "Searches the web for up-to-date information",
			Parameters: api.ToolFunctionParameters{
				Type: "object",
				Properties: properties,
				Required: []string{"search_query"},
			},
		},
	}

	return VyntrSearchTool
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
