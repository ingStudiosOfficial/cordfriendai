package tools

import (
	"fmt"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

type GetTimeArgs struct {
	LocationIana string `json:"location_iana"`
}

func GetTimeTool() openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		Function: shared.FunctionDefinitionParam{
			Name:        "get_time",
			Description: openai.String("Gets the time given a location"),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"location_iana": map[string]any{
						"type":        "string",
						"description": "The IANA location you want to find the time for",
					},
				},
				"required": []string{"location_iana"},
			},
		},
	}
}

func GetTime(location string) time.Time {
	timeNow := time.Now()

	timeLocation, err := time.LoadLocation(location)
	if err != nil {
		fmt.Println("Error while loading time location:", err)
		return timeNow
	}

	timeInLocation := timeNow.In(timeLocation)
	fmt.Println("Time in location:", timeInLocation)

	return timeInLocation
}