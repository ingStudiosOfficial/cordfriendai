package tools

import (
	"fmt"
	"io"
	"net/http"

	"github.com/ollama/ollama/api"
)

type GetWeatherArgs struct {
	Locations string `json:"location"`
}

func GetWeatherTool() api.Tool {
	properties := api.NewToolPropertiesMap()

	properties.Set("location", api.ToolProperty{
		Type:        api.PropertyType{"string"},
		Description: "The name of the location you want to fetch the weather for",
	})

	var WeatherTool = api.Tool {
		Type: "function",
		Function: api.ToolFunction{
			Name: "fetch_weather",
			Description: "Fetches the weather given a location",
			Parameters: api.ToolFunctionParameters{
				Type: "object",
				Properties: properties,
				Required: []string{"location"},
			},
		},
	}

	return WeatherTool
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
