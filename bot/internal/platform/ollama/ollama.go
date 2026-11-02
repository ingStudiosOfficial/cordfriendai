package ollama

import (
	"bot/internal/platform/ollama/tools"
	"bot/internal/response"
	"bot/internal/storage/mongodb"
	"bot/internal/structs"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OllamaRequest struct {
	Ctx    context.Context
	Client openai.Client
	Model  string
}

func InitializeOllama() *OllamaRequest {
	if (os.Getenv("USE_CLOUD") == "true") {
		client := openai.NewClient(
			option.WithBaseURL("https://api.groq.com/openai/v1"),
			option.WithAPIKey(os.Getenv("GROQ_API_KEY")),
		)
		ctx := context.Background()

		model := "openai/gpt-oss-20b"

		return &OllamaRequest{
			Ctx:    ctx,
			Client: client,
			Model:  model,
		}
	} else {
		client := openai.NewClient(
			option.WithBaseURL("http://localhost:11434/v1/"),
			option.WithAPIKey("ollama"),
		)

		ctx := context.Background()

		model := "qwen3.5:4b"

		return &OllamaRequest{
			Ctx:    ctx,
			Client: client,
			Model:  model,
		}
	}
}

func (r *OllamaRequest) GenerateFromText(s *discordgo.Session, m *discordgo.MessageCreate, repository mongodb.BotRepository) error {
	message := m.Content

	history := []openai.ChatCompletionMessageParamUnion{}

	conv, fcerr := repository.FetchConversations(m.GuildID)
	if fcerr != nil {
		return fcerr
	}

	persona, err := repository.FetchBotPersona(m.GuildID)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		history = append(history, openai.SystemMessage(persona))
	}

	for _, val := range conv {
		history = append(history, openai.UserMessage(val.User.Name+val.User.Message))
		history = append(history, openai.AssistantMessage(val.Bot))
	}

	history = append(history, openai.UserMessage(message))

	toolsCalled := []string{}

	for {
		params := openai.ChatCompletionNewParams{
			Model:    r.Model,
			Messages: history,
			Tools: []openai.ChatCompletionToolParam{
				tools.GetTimeTool(),
				tools.GetVyntrSearchTool(),
				tools.GetWeatherTool(),
			},
			ReasoningEffort: "low",
		}

		stream := r.Client.Chat.Completions.NewStreaming(r.Ctx, params)
		acc := openai.ChatCompletionAccumulator{}

		finalResp := ""

		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				content := chunk.Choices[0].Delta.Content
				fmt.Print(content)
				finalResp += content
			}
		}

		if err := stream.Err(); err != nil {
			return fmt.Errorf("error during generation: %v\n", err)
		}

		if len(acc.Choices) == 0 {
			return fmt.Errorf("error during generation: no choices returned\n")
		}

		toolCalls := acc.Choices[0].Message.ToolCalls

		if len(toolCalls) == 0 {
			s.ChannelMessageSendReply(m.ChannelID, finalResp, m.Reference())
			history = append(history, openai.AssistantMessage(finalResp))
			repository.AddConversations(m.GuildID, structs.User{
				Name:    m.Author.DisplayName(),
				Message: message,
			}, finalResp)
			return nil
		}

		history = append(history, acc.Choices[0].Message.ToParam())

		for _, toolCall := range toolCalls {
			result := ""

			switch toolCall.Function.Name {
			case "get_time":
				var timeArgs tools.GetTimeArgs
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &timeArgs); err != nil {
					result = "An error occurred while getting the current time."
					break
				}

				result = tools.GetTime(timeArgs.LocationIana).String()

			case "fetch_weather":
				var weatherArgs tools.GetWeatherArgs
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &weatherArgs); err != nil {
					result = "An error occurred while getting the weather."
					break
				}

				apiKey, err := repository.FetchWeatherApiKey(m.GuildID)
				if err != nil {
					fmt.Println("Error while fetching weather API key:", err)
					result = "There was an error while fetching the weather. Please check whether your API key is valid and your rate limits."
					break
				}

				weather, err := tools.GetWeather(apiKey, weatherArgs.Locations)
				if err != nil {
					fmt.Println("Error while fetching weather:", err)
					result = "There was an error while fetching the weather. Please check whether your API key is valid and your rate limits."
					break
				}

				result = weather

			case "search_web":
				var vyntrSearchArgs tools.VyntrSearchArgs
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &vyntrSearchArgs); err != nil {
					result = "An error occurred while searching the web."
					break
				}

				apiKey, err := repository.FetchVyntrApiKey(m.GuildID)
				if err != nil {
					fmt.Println("Error while fetching Vyntr API key:", err)
					result = "An error occurred while searching the web."
					break
				}

				searchResult, err := tools.VyntrSearch(apiKey, vyntrSearchArgs.SearchQuery)
				if err != nil {
					fmt.Println("Error while searching the web:", err)
					result = "An error occurred while searching the web."
					break
				}

				result = searchResult

			default:
				result = "Unknown tool: " + toolCall.Function.Name
			}

			history = append(history, openai.ToolMessage(result, toolCall.ID))

			toolsCalled = append(toolsCalled, toolCall.Function.Name)

			var builder strings.Builder

			for i, tool := range toolsCalled {
				fmt.Fprintf(&builder, "%s. %s", strconv.Itoa(i+1), tool)
			}

			fmt.Println(history)
		}
	}
}

func (r *OllamaRequest) GenerateFromCommand(s *discordgo.Session, i *discordgo.InteractionCreate, repository mongodb.BotRepository) error {
	err := response.DeferResponse(s, i, "Pondering...")
	if err != nil {
		return err
	}
	
	interactionData := i.ApplicationCommandData()

	options := interactionData.Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))

	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	message := optionMap["message"].StringValue()

	history := []openai.ChatCompletionMessageParamUnion{}

	conv, fcerr := repository.FetchConversations(i.GuildID)
	if fcerr != nil {
		return fcerr
	}

	persona, err := repository.FetchBotPersona(i.GuildID)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		history = append(history, openai.SystemMessage(persona))
	}

	for _, val := range conv {
		history = append(history, openai.UserMessage(val.User.Name+val.User.Message))
		history = append(history, openai.AssistantMessage(val.Bot))
	}

	history = append(history, openai.UserMessage(message))

	toolsCalled := []string{}

	for {
		params := openai.ChatCompletionNewParams{
			Model:    r.Model,
			Messages: history,
			Tools: []openai.ChatCompletionToolParam{
				tools.GetTimeTool(),
				tools.GetVyntrSearchTool(),
				tools.GetWeatherTool(),
			},
			ReasoningEffort: "low",
		}

		stream := r.Client.Chat.Completions.NewStreaming(r.Ctx, params)
		acc := openai.ChatCompletionAccumulator{}

		finalResp := ""

		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				content := chunk.Choices[0].Delta.Content
				fmt.Print(content)
				finalResp += content
			}
		}

		if err := stream.Err(); err != nil {
			return fmt.Errorf("error during generation: %v\n", err)
		}

		if len(acc.Choices) == 0 {
			return fmt.Errorf("error during generation: no choices returned\n")
		}

		toolCalls := acc.Choices[0].Message.ToolCalls

		if len(toolCalls) == 0 {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &finalResp,
			})
			history = append(history, openai.AssistantMessage(finalResp))
			repository.AddConversations(i.GuildID, structs.User{
				Name:    i.Member.DisplayName(),
				Message: message,
			}, finalResp)
			return nil
		}

		history = append(history, acc.Choices[0].Message.ToParam())

		for _, toolCall := range toolCalls {
			result := ""

			switch toolCall.Function.Name {
			case "get_time":
				var timeArgs tools.GetTimeArgs
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &timeArgs); err != nil {
					result = "An error occurred while getting the current time."
					break
				}

				result = tools.GetTime(timeArgs.LocationIana).String()

			case "fetch_weather":
				var weatherArgs tools.GetWeatherArgs
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &weatherArgs); err != nil {
					result = "An error occurred while getting the weather."
					break
				}

				apiKey, err := repository.FetchWeatherApiKey(i.GuildID)
				if err != nil {
					fmt.Println("Error while fetching weather API key:", err)
					result = "There was an error while fetching the weather. Please check whether your API key is valid and your rate limits."
					break
				}

				weather, err := tools.GetWeather(apiKey, weatherArgs.Locations)
				if err != nil {
					fmt.Println("Error while fetching weather:", err)
					result = "There was an error while fetching the weather. Please check whether your API key is valid and your rate limits."
					break
				}

				result = weather

			case "search_web":
				var vyntrSearchArgs tools.VyntrSearchArgs
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &vyntrSearchArgs); err != nil {
					result = "An error occurred while searching the web."
					break
				}

				apiKey, err := repository.FetchVyntrApiKey(i.GuildID)
				if err != nil {
					fmt.Println("Error while fetching Vyntr API key:", err)
					result = "An error occurred while searching the web."
					break
				}

				searchResult, err := tools.VyntrSearch(apiKey, vyntrSearchArgs.SearchQuery)
				if err != nil {
					fmt.Println("Error while searching the web:", err)
					result = "An error occurred while searching the web."
					break
				}

				result = searchResult

			default:
				result = "Unknown tool: " + toolCall.Function.Name
			}

			history = append(history, openai.ToolMessage(result, toolCall.ID))

			toolsCalled = append(toolsCalled, toolCall.Function.Name)

			var builder strings.Builder

			for i, tool := range toolsCalled {
				fmt.Fprintf(&builder, "%s. %s", strconv.Itoa(i+1), tool)
			}

			fmt.Println(history)
		}
	}
}