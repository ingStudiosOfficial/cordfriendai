package tools

import (
	"fmt"
	"time"

	"github.com/ollama/ollama/api"
)

type GetTimeArgs struct {
	LocationIana string `json:"location_iana"`
}

func GetTimeTool() api.Tool {
	properties := api.NewToolPropertiesMap()

	properties.Set("location_iana", api.ToolProperty{
		Type:        api.PropertyType{"string"},
		Description: "The IANA location you want to find the time for",
	})

	var TimeTool = api.Tool {
		Type: "function",
		Function: api.ToolFunction{
			Name: "get_time",
			Description: "Gets the time given a location",
			Parameters: api.ToolFunctionParameters{
				Type: "object",
				Properties: properties,
				Required: []string{"location_iana"},
			},
		},
	}

	return TimeTool
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
