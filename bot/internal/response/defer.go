package response

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func DeferResponse(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		fmt.Println("Error deferring:", err)
		return fmt.Errorf("failed to defer response: %v", err)
	}

	temporaryMessage := msg
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &temporaryMessage,
	})
	if err != nil {
		fmt.Println("Error setting temporary message:", err)
		return fmt.Errorf("failed to set temporary message: %v", err)
	}

	return nil
}
