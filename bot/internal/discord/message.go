package discord

import (
	"fmt"

	"bot/internal/platform/gemini"
	"bot/internal/storage/mongodb"
	"bot/internal/strings"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MessageParams struct {
	Db *mongo.Database
}

func MessageHandler(db *mongo.Database) *MessageParams {
	return &MessageParams{
		Db: db,
	}
}

func (r *MessageParams) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("Message created, running checks.")

	botRepository := mongodb.NewBotRepository(r.Db)
	geminiAPIClient := gemini.NewAPIRequest(botRepository, m)

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
		err := s.ChannelTyping(m.ChannelID)
		if err != nil {
			fmt.Println("Failed to add typing indicator:", err)
			s.ChannelMessageSend(m.ChannelID, strings.TruncateString("Failed to respond.", 2000))
			return
		}

		fmt.Println("Bot mentioned, responding.")

		response := geminiAPIClient.RequestGenAi()
		fmt.Println("Returning response:", response)

		s.ChannelMessageSend(m.ChannelID, strings.TruncateString(response, 2000))
	}
}
