package tools

import (
	"fmt"
	"io"
	"net/http"

	"google.golang.org/genai"
)

var WeatherTool = &genai.FunctionDeclaration{
	Name:        "getWeather",
	Description: "Gets the current weather in a town, city, state, prefecture, province, or country",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"location": {Type: genai.TypeString},
		},
		Required: []string{"location"},
	},
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
