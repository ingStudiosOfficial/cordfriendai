package ollama

import (
	"bot/internal/platform/ollama/tools"
	"bot/internal/storage/mongodb"
	"bot/internal/structs"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ollama/ollama/api"
)

type OllamaRequest struct {
	Ctx     context.Context
	Client  *api.Client
	Model   string
}

func InitializeOllama() *OllamaRequest {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatalf("Failed to create client from environment: %v\n", err)
	}

	ctx := context.Background()

	model := "qwen3.5:4b"

	return &OllamaRequest{
		Ctx:     ctx,
		Client:  client,
		Model:   model,
	}
}

func (r *OllamaRequest) GenerateFromText(s *discordgo.Session, m *discordgo.MessageCreate, repository mongodb.BotRepository) error {
	message := m.Content

	history := []api.Message{}

	conv, fcerr := repository.FetchConversations(m.GuildID)
	if fcerr != nil {
		return fcerr
	}

	for _, val := range conv {
		history = append(history, api.Message{
			Role: "user",
			Content: val.User.Name + val.User.Message,
		})

		history = append(history, api.Message{
			Role: "assistant",
			Content: val.Bot,
		})
	}

	history = append(history, api.Message{
		Role:    "user",
		Content: message,
	})

	var initialMessage, err = s.ChannelMessageSendReply(m.ChannelID, "Pondering...", m.Reference())
	if err != nil {
		return err
	}

	toolsCalled := []string{}

	for {
		req := &api.ChatRequest{
			Model:    r.Model,
			Messages: history,
			// Tools: []api.Tool{
			// 	tools.GetTimeTool(),
			// 	tools.GetVyntrSearchTool(),
			// 	tools.GetWeatherTool(),
			// },
			Think: &api.ThinkValue{
				Value: false,
			},
		}

		var toolCalls []api.ToolCall

		finalResp := ""

		respFunc := func(resp api.ChatResponse) error {
			if resp.Message.Content != "" {
				fmt.Print(resp.Message.Content)
				finalResp += resp.Message.Content
				initialMessage, err = s.ChannelMessageEdit(m.ChannelID, initialMessage.ID, finalResp)
				if err != nil {
					log.Printf("Error: %v", err)
				}
			}

			if len(resp.Message.ToolCalls) > 0 {
				toolCalls = append(toolCalls, resp.Message.ToolCalls...)
			}

			return nil
		}

		err := r.Client.Chat(r.Ctx, req, respFunc)
		if err != nil {
			return fmt.Errorf("error during generation: %v\n", err)
		}

		if len(toolCalls) == 0 {
			history = append(history, api.Message{
				Role:    "assistant",
				Content: finalResp,
			})
			repository.AddConversations(m.GuildID, structs.User{
				Name: m.Author.DisplayName(),
				Message: message,
			}, finalResp)
			return nil
		}

		history = append(history, api.Message{
			Role:      "assistant",
			Content:   finalResp,
			ToolCalls: toolCalls,
		})

		for _, toolCall := range toolCalls {
			switch toolCall.Function.Name {
			case "get_time":
				jsonBytes, err := json.Marshal(toolCall.Function.Arguments)
				if err != nil {
					history = append(history, api.Message{
						Role:    "tool",
						Content: "An error occurred while getting the current time.",
					})
					break
				}

				var timeArgs tools.GetTimeArgs
				if err := json.Unmarshal(jsonBytes, &timeArgs); err != nil {
					history = append(history, api.Message{
						Role:    "tool",
						Content: "An error occurred while getting the current time.",
					})
					break
				}

				time := tools.GetTime(timeArgs.LocationIana).String()

				history = append(history, api.Message{
					Role:    "tool",
					Content: time,
				})

			case "fetch_weather":
				jsonBytes, err := json.Marshal(toolCall.Function.Arguments)
				if err != nil {
					history = append(history, api.Message{
						Role:    "tool",
						Content: "An error occurred while getting the weather.",
					})
					break
				}

				var weatherArgs tools.GetWeatherArgs
				if err := json.Unmarshal(jsonBytes, &weatherArgs); err != nil {
					history = append(history, api.Message{
						Role:    "tool",
						Content: "An error occurred while getting the weather.",
					})
					break
				}

				apiKey, err := repository.FetchWeatherApiKey(m.GuildID)
				if err != nil {
					fmt.Println("Error while fetching weather API key:", err)
					history = append(history, api.Message{
						Role: "tool",
						Content: "There was an error while fetching the weather. Please check whether your API key is valid and your rate limits.",
					})
				}

				weather, err := tools.GetWeather(apiKey, weatherArgs.Locations)
				if err != nil {
					fmt.Println("Error while fetching weather:", err)
					history = append(history, api.Message{
						Role: "tool",
						Content: "There was an error while fetching the weather. Please check whether your API key is valid and your rate limits.",
					})
				}

				history = append(history, api.Message{
					Role:    "tool",
					Content: weather,
				})

			case "search_web":
				jsonBytes, err := json.Marshal(toolCall.Function.Arguments)
				if err != nil {
					history = append(history, api.Message{
						Role:    "tool",
						Content: "An error occurred while searching the web.",
					})
					break
				}

				var vyntrSearchArgs tools.VyntrSearchArgs
				if err := json.Unmarshal(jsonBytes, &vyntrSearchArgs); err != nil {
					history = append(history, api.Message{
						Role:    "tool",
						Content: "An error occurred while searching the web.",
					})
					break
				}

				apiKey, err := repository.FetchVyntrApiKey(m.GuildID)
				if err != nil {
					fmt.Println("Error while fetching weather API key:", err)
					history = append(history, api.Message{
						Role: "tool",
						Content: "An error occurred while searching the web.",
					})
				}

				result, err := tools.VyntrSearch(apiKey, vyntrSearchArgs.SearchQuery)
				if err != nil {
					fmt.Println("Error while searching the web:", err)
					history = append(history, api.Message{
						Role: "tool",
						Content: "An error occurred while searching the web.",
					})
				}

				history = append(history, api.Message{
					Role:    "tool",
					Content: result,
				})
			}

			toolsCalled = append(toolsCalled, toolCall.Function.Name)

			var builder strings.Builder

			for i, tool := range toolsCalled {
				fmt.Fprintf(&builder, "%s. %s", strconv.Itoa(i+1), tool)
			}

			fmt.Println(history)
		}
	}
}