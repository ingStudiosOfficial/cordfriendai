package tools

import (
	"fmt"
	"io"
	"net/http"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

type GetWeatherArgs struct {
	Locations string `json:"location"`
}

func GetWeatherTool() openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		Function: shared.FunctionDefinitionParam{
			Name:        "fetch_weather",
			Description: openai.String("Fetches the weather given a location"),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "The name of the location you want to fetch the weather for",
					},
				},
				"required": []string{"location"},
			},
		},
	}
}

func GetWeather(apiKey string, location string) (string, error) {
	fmt.Println("Location to fetch:", location)

	urlToFetch := "https://api.openweathermap.org/data/2.5/weather?q=" + location + "&appid=" + apiKey

	resp, err := http.Get(urlToFetch)

	if err != nil {
		return "", fmt.Errorf("error fetching weather: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)

		fmt.Println("Fetch failed:", bodyString)

		return "", fmt.Errorf("error fetching weather: %v", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading body:", err)
		return "", fmt.Errorf("failed to read body: %v", err)
	}

	return string(bodyBytes), nil
}
