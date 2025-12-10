package commands

import (
	"bot/internal/storage/mongodb"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func UpdateBotNickname(s *discordgo.Session, guildID string, i *discordgo.InteractionCreate, db *mongo.Database) error {
	fmt.Println("Update nickname command called.")

	botRepository := mongodb.NewBotRepository(db)

	nickname, err := botRepository.FetchNickname(guildID)
	if err != nil {
		return fmt.Errorf("failed to fetch nickname: %w", err)
	}
	if nickname == "" {
		return nil
	}

	fmt.Println("Changing nickname for guild ID:", guildID)

	err = s.GuildMemberNickname(guildID, "@me", nickname)
	if err != nil {
		return fmt.Errorf("failed to set nickname: %w", err)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Nickname updated successfully!",
		},
	})
	if err != nil {
		fmt.Println("Failed to respond to interaction:", err)
	}

	return nil
}
