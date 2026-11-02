package discord

import (
	"fmt"
	"log"

	"bot/internal/platform/ollama"
	"bot/internal/storage/mongodb"
	"bot/internal/strings"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MessageParams struct {
	Db *mongo.Database
	Client *ollama.OllamaRequest
}

func MessageHandler(db *mongo.Database) (*MessageParams, error) {
	ollamaApiClient := ollama.InitializeOllama()

	return &MessageParams{
		Db: db,
		Client: ollamaApiClient,
	}, nil
}

func (r *MessageParams) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("Message created, running checks.")

	// Ignore messages created by the bot
	if m.Author.ID == s.State.User.ID {
		fmt.Println("Returning as message is created by bot itself.")
		return
	}

	mentioned := false

	for _, user := range m.Mentions {
		if user.ID == s.State.User.ID {
			mentioned = true
			break
		}
	}

	if mentioned {
		botRepository := mongodb.NewBotRepository(r.Db)

		err := s.ChannelTyping(m.ChannelID)
		if err != nil {
			fmt.Println("Failed to add typing indicator:", err)
			s.ChannelMessageSend(m.ChannelID, strings.TruncateString("Failed to respond.", 2000))
			return
		}

		fmt.Println("Bot mentioned, responding.")

		genErr := r.Client.GenerateFromText(s, m, *botRepository)
		if genErr != nil {
			log.Printf("Error while generating: %v", err)
			s.ChannelMessageSend(m.ChannelID, genErr.Error())
		}
	}
}
