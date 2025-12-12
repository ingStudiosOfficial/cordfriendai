package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/genai"
)

func Test() {
	ctx := context.Background()

	// Create client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  "apiKey",
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 1. Define your function declaration
	getCurrentTimeFunc := &genai.FunctionDeclaration{
		Name:        "getCurrentTime",
		Description: "Returns the current date and time",
		Parameters: &genai.Schema{
			Type: "object",
			Properties: map[string]*genai.Schema{
				"timezone": {
					Type:        "string",
					Description: "The timezone (optional, defaults to UTC)",
				},
			},
		},
	}

	// 2. Configure the generation with the tool
	config := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{
			{FunctionDeclarations: []*genai.FunctionDeclaration{getCurrentTimeFunc}},
		},
		Temperature: genai.Ptr(float32(0.0)),
	}

	// 3. Send initial prompt
	contents := []*genai.Content{
		{
			Parts: []*genai.Part{{Text: "What's the current time?"}},
			Role:  "user",
		},
	}

	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, config)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Check if model requested a function call
	functionCalls := resp.FunctionCalls()
	if len(functionCalls) > 0 {
		// 5. Execute the function and get the result
		for _, fc := range functionCalls {
			fmt.Printf("Model requested function: %s\n", fc.Name)

			// Your actual function implementation
			currentTime := getCurrentTime()

			// 6. Send function response back to model
			contents = append(contents, resp.Candidates[0].Content)
			contents = append(contents, &genai.Content{
				Parts: []*genai.Part{
					genai.NewPartFromFunctionResponse(fc.Name, map[string]any{
						"time": currentTime,
					}),
				},
				Role: "user",
			})
		}

		// 7. Get final response with function results
		finalResp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, config)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Final answer: %s\n", finalResp.Text())
	} else {
		fmt.Printf("Direct answer: %s\n", resp.Text())
	}
}

// Your actual function implementation
func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}
