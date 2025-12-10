package discord

import (
	"fmt"

	"bot/internal/platform/gemini"
	"bot/internal/storage/mongodb"
	"bot/internal/utils"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MessageCreateParams struct {
	Db *mongo.Database
}

func MessageHandler(db *mongo.Database) *MessageCreateParams {
	return &MessageCreateParams{
		Db: db,
	}
}

func (r *MessageCreateParams) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
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
		fmt.Println("Bot mentioned, responding.")

		response := geminiAPIClient.RequestGenAi()
		fmt.Println("Returning response:", response)

		s.ChannelMessageSend(m.ChannelID, utils.TruncateString(response, 2000))
	}
}
