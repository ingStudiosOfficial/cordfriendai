package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"bot/internal/commands"
)

func HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *mongo.Database) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "load-name":
		err := commands.UpdateBotNickname(s, i.GuildID, i, db)

		if err != nil {
			fmt.Println("Error while updating bot nickname:", err)
		}
	}
}
