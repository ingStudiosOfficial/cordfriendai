package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"bot/internal/storage/mongodb"
	"bot/internal/structs"

	"github.com/bwmarrin/discordgo"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
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
		fmt.Print("Error while fetching conversations:", err)
		conversationsString = "No conversations stored in history yet."
	} else {
		conversationsByte, err := json.Marshal(conversations)
		if err != nil {
			fmt.Print("Error while converting conversations:", err)
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
		systemInstructions = "You are a helpful Discord bot. Please be as concise as possible - but still give helpful information."
	} else {
		systemInstructions = fetchedInstructions
	}

	var sentUser = r.M.Author.DisplayName()
	fmt.Println("User who sent message:", sentUser)

	var sentUserId = r.M.Author.ID
	fmt.Println("User ID who sent message:", sentUserId)

	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))

	if err != nil {
		fmt.Println("Error creating new Gemini client:", err)
		return "Error creating new Gemini client."
	}

	var promptToSend = "Conversation history: " + conversationsString + " System message: " + systemInstructions + " User '" + sentUser + "' sent the message: " + r.M.Content
	fmt.Println("Sending prompt:", promptToSend)

	resp, err := client.GenerativeModel("gemini-2.5-flash").GenerateContent(
		ctx,
		genai.Text(promptToSend),
	)
	if err != nil {
		fmt.Println("Error while generating content:", err)
		return "There was an error while generating your content. If this persists, try deleting your bots conversations"
	}

	var messageSent structs.User
	messageSent.Name = r.M.Author.DisplayName()
	messageSent.Message = r.M.Content

	var response string = ""

	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				response += fmt.Sprintf("%v", part)
			}
		}
	}

	if response != "" {
		r.Repository.AddConversations(r.M.GuildID, messageSent, response)
	}

	return "<@" + sentUserId + "> " + response
}
