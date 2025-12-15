package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"bot/internal/platform/gemini/tools"
	"bot/internal/storage/mongodb"
	"bot/internal/structs"

	"github.com/bwmarrin/discordgo"
	"google.golang.org/genai"
)

type APIRequest struct {
	Repository *mongodb.BotRepository
	M          *discordgo.MessageCreate
}

func NewAPIRequest(repository *mongodb.BotRepository, m *discordgo.MessageCreate) *APIRequest {
	return &APIRequest{
		Repository: repository,
		M:          m,
	}
}

func (r *APIRequest) RequestGenAi() string {
	fmt.Println("Generating response...")

	apiKey, err := r.Repository.FetchApiKey(r.M.GuildID)
	if err != nil {
		fmt.Println("Error while fetching API key:", err)
		return "Could not fetch API key for this server."
	}
	if apiKey == "" {
		fmt.Println("Could not fetch API key.")
		return "Could not fetch API key for this server."
	}

	var conversationsString string

	conversations, err := r.Repository.FetchConversations(r.M.GuildID)
	if err != nil {
		fmt.Println("Error while fetching conversations:", err)
		conversationsString = "No conversations stored in history yet."
	} else {
		conversationsByte, err := json.Marshal(conversations)
		if err != nil {
			fmt.Println("Error while converting conversations:", err)
			conversationsString = "No conversations stored in history yet."
		} else {
			conversationsString = string(conversationsByte)
		}
	}

	fmt.Println("Conversations:", conversationsString)

	var systemInstructions string
	fetchedInstructions, err := r.Repository.FetchBotPersona(r.M.GuildID)
	if err != nil {
		fmt.Println("Error while fetching bot persona:", err)
	}
	if fetchedInstructions == "" {
		systemInstructions = ""
	} else {
		systemInstructions = "User defined instructions: '" + fetchedInstructions + "'"
	}

	var sentUser = r.M.Author.DisplayName()
	fmt.Println("User who sent message:", sentUser)

	var sentUserId = r.M.Author.ID
	fmt.Println("User ID who sent message:", sentUserId)

	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		fmt.Println("Error creating new Gemini client:", err)
		return "Error creating new Gemini client."
	}

	var promptToSend = "Conversation history: '" + conversationsString + "' System message: '" + systemInstructions + "' User '" + sentUser + "' sent the message: '" + r.M.Content + "'"
	fmt.Println("Sending prompt:", promptToSend)

	config := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{
			{FunctionDeclarations: []*genai.FunctionDeclaration{tools.TimeTool, tools.WeatherTool, tools.SearchTool}},
		},
	}

	contents := []*genai.Content{
		{
			Parts: []*genai.Part{{Text: promptToSend}},
		},
	}

	resp, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash-lite",
		contents,
		config,
	)
	if err != nil {
		fmt.Println("Error while generating content:", err)
		return "There was an error while generating your content. If this persists, try deleting your bots conversations or checking your rate limits."
	}

	functionCalls := resp.FunctionCalls()
	if len(functionCalls) > 0 {
		for _, fc := range functionCalls {
			switch fc.Name {
			case "getTime":
				time := tools.GetTime(fc.Args["location_iana"].(string)).String()

				contents = append(contents, resp.Candidates[0].Content)
				contents = append(contents, &genai.Content{
					Parts: []*genai.Part{
						genai.NewPartFromFunctionResponse(fc.Name, map[string]any{
							"time": time,
						}),
					},
				})
			case "getWeather":
				apiKey, err := r.Repository.FetchWeatherApiKey(r.M.GuildID)
				if err != nil {
					fmt.Println("Error while fetching weather API key:", err)
					return "There was an error while fetching the weather. Please check whether your API key is valid and your rate limits."
				}

				weather, err := tools.GetWeather(apiKey, fc.Args["location"].(string))
				if err != nil {
					fmt.Println("Error while fetching weather data:", err)
					return "There was an error while fetching the weather. Please check whether your API key is valid and your rate limits."
				}

				contents = append(contents, resp.Candidates[0].Content)
				contents = append(contents, &genai.Content{
					Parts: []*genai.Part{
						genai.NewPartFromFunctionResponse(fc.Name, map[string]any{
							"weather": weather,
						}),
					},
				})
			case "vyntrSearch":
				apiKey, err := r.Repository.FetchVyntrApiKey(r.M.GuildID)
				if err != nil {
					fmt.Println("Error while fetching Vyntr API key:", err)
					return "There was an error while searching with Vyntr. Please check whether your API key is valid and your rate limits."
				}

				results, err := tools.VyntrSearch(apiKey, fc.Args["query"].(string))
				if err != nil {
					fmt.Println("Error while searching with Vyntr:", err)
					return "There was an error while searching with Vyntr. Please check whether your API key is valid and your rate limits."
				}

				contents = append(contents, resp.Candidates[0].Content)
				contents = append(contents, &genai.Content{
					Parts: []*genai.Part{
						genai.NewPartFromFunctionResponse(fc.Name, map[string]any{
							"results": results,
						}),
					},
				})
			}
		}
	}

	finalResp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash-lite", contents, config)
	if err != nil {
		fmt.Println("Error while generating content:", err)
		return "There was an error while generating your content. If this persists, try deleting your bots conversations or checking your rate limits."
	}

	var messageSent structs.User
	messageSent.Name = r.M.Author.DisplayName()
	messageSent.Message = r.M.Content

	var response string = finalResp.Text()

	if response != "" {
		r.Repository.AddConversations(r.M.GuildID, messageSent, response)
	}

	return "<@" + sentUserId + "> " + response
}
