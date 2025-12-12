package tools

import (
	"fmt"
	"time"

	"google.golang.org/genai"
)

var TimeTool = &genai.FunctionDeclaration{
	Name:        "getTime",
	Description: "Gets the current time",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"location_iana": {Type: genai.TypeString},
		},
		Required: []string{"location_iana"},
	},
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
