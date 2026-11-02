package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"bot/internal/commands"
	"bot/internal/platform/ollama"
	"bot/internal/storage/mongodb"
)

type CommandParams struct {
	Db *mongo.Database
	Client *ollama.OllamaRequest
}

func CommandHandler(db *mongo.Database, client *ollama.OllamaRequest) *CommandParams {
	return &CommandParams{
		Db: db,
		Client: client,
	}
}

func (r *CommandParams) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "load-name":
		err := commands.UpdateBotNickname(s, i.GuildID, i, r.Db)

		if err != nil {
			fmt.Println("Error while updating bot nickname:", err)
			errorMessage := fmt.Sprintf("Error while updating bot nickname: %v", err)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &errorMessage,
			})
		}
	case "fetch-neko":
		err := commands.GenerateNeko(s, i)

		if err != nil {
			fmt.Println("Error while fetching neko:", err)
			errorMessage := fmt.Sprintf("Error while fetching husbando/kitsune/neko/waifu: %v", err)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &errorMessage,
			})
		}
	case "talk":
		err := r.Client.GenerateFromCommand(s, i, *mongodb.NewBotRepository(r.Db))
		if err != nil {
			fmt.Println("Error while fetching neko:", err)
			errorMessage := fmt.Sprintf("Error while fetching husbando/kitsune/neko/waifu: %v", err)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &errorMessage,
			})
		}
	}
}
