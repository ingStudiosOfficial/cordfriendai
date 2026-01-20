package commands

import (
	"bot/internal/response"
	"fmt"
	"net/http"
	"strings"

	nb "github.com/Yakiyo/nekos_best.go"
	"github.com/bwmarrin/discordgo"
)

func GenerateNeko(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	err := response.DeferResponse(s, i, "Please wait while we fetch the images...")
	if err != nil {
		return err
	}

	interactionData := i.ApplicationCommandData()

	options := interactionData.Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))

	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	var imageType string
	var imageCount int

	if opt, ok := optionMap["type"]; ok {
		imageType = strings.ToLower(opt.StringValue())
	}

	if opt, ok := optionMap["qty"]; ok {
		imageCount = int(opt.IntValue())
	}

	if imageCount > 10 {
		return fmt.Errorf("Please enter valid image count to generate (1 - 10 images).")
	}

	if imageCount == 0 {
		imageCount = 1
	}

	switch imageType {
	case "husbando":
	case "kitsune":
	case "neko":
	case "waifu":
		fmt.Println("Valid image type.")
	default:
		fmt.Printf("Image type '%v' inavid.\n", imageType)
		return fmt.Errorf("Image type '%v' inavid.", imageType)
	}

	res, err := nb.FetchMany(imageType, imageCount)

	if err != nil {
		fmt.Println("Error while fetching images:", err)
		return fmt.Errorf("Failed to fetch %v: %v", imageType, err)
	}

	imageUrls := make([]string, 0, len(res))

	for _, url := range res {
		fmt.Println("Fetched URL:", url.Url)
		imageUrls = append(imageUrls, url.Url)
	}

	files := make([]*discordgo.File, 0, len(imageUrls))
	embeds := make([]*discordgo.MessageEmbed, 0, len(imageUrls))

	for idx, url := range imageUrls {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("An error occurred while fetching images: %v", err)
		}
		defer resp.Body.Close()

		fileName := fmt.Sprintf("embed_%v.png", idx)

		files = append(files, &discordgo.File{
			Name:   fileName,
			Reader: resp.Body,
		})

		embedMessage := fmt.Sprintf("Fetched from [nekos.best](https://nekos.best). View the original image [here](%v). [Artist](%v): %v", res[idx].Source_url, res[idx].Artist_href, res[idx].Artist_name)

		embeds = append(embeds, &discordgo.MessageEmbed{
			Title:       imageType,
			Description: embedMessage,
			Type:        discordgo.EmbedTypeImage,
			Image: &discordgo.MessageEmbedImage{
				URL: "attachment://" + fileName,
			},
		})
	}

	responseMessage := fmt.Sprintf("[%v](https://nekos.best) fetched successfully", imageType)

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Files:   files,
		Embeds:  &embeds,
		Content: &responseMessage,
	})
	if err != nil {
		return fmt.Errorf("An error occurred while fetching images: %v", err)
	}

	return nil
}
