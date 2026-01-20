package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bot/internal/discord"
	"bot/internal/scheduler"
	"bot/internal/storage/mongodb"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const VERSION = "1.6.0"

var STARTTIME time.Time

var dg *discordgo.Session

func main() {
	fmt.Println("Cordfriend AI [Version " + VERSION + "]")

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	STARTTIME = time.Now()

	scheduler.StartUptimePingScheduler(STARTTIME, ctx)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error while loading .env file:", err)
	}

	var mongoClient = mongodb.ConnectToMongo()

	var databaseName = "cordfriendAI"

	var discordToken = os.Getenv("DISCORD_TOKEN")

	if discordToken == "" {
		log.Fatal("Discord token not set.")
	}

	fmt.Println("Discord token:", discordToken)

	// Create Discord session
	var errds error
	dg, errds = discordgo.New("Bot " + discordToken)
	if errds != nil {
		log.Fatal("Error starting Discord session:", err)
	}

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)

		var minCount float64 = 1.0

		var commands = []*discordgo.ApplicationCommand{
			{
				Name:        "load-name",
				Description: "Sets the nickname to the saved name from the Cordfriend AI dashboard.",
				Type:        discordgo.ChatApplicationCommand,
			},
			{
				Name:        "fetch-neko",
				Description: "Fetches an image of a husbando/kitsune/neko/waifu of your choice and count.",
				Type:        discordgo.ChatApplicationCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "type",
						Description: "Can be husbando/kitsune/neko/waifu",
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{
								Name:  "husbando",
								Value: "husbando",
							},
							{
								Name:  "kitsune",
								Value: "kitsune",
							},
							{
								Name:  "neko",
								Value: "neko",
							},
							{
								Name:  "waifu",
								Value: "waifu",
							},
						},
						Required: true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "qty",
						Description: "Amount of images to fetch",
						MinValue:    &minCount,
						MaxValue:    10.0,
						Required:    false,
					},
				},
			},
		}

		registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))

		for i, v := range commands {
			cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
			if err != nil {
				log.Panicf("Cannot create '%v' command: %v", v.Name, err)
			}
			registeredCommands[i] = cmd
		}
	})

	messageHandler := discord.MessageHandler(mongoClient.Database(databaseName))
	commandHandler := discord.CommandHandler(mongoClient.Database(databaseName))

	dg.AddHandler(commandHandler.HandleCommand)
	dg.AddHandler(messageHandler.HandleMessageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	// Open a websocket to connect to Discord
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()

	// Close MongoDB connection
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			log.Println("Error disconnecting from MongoDB:", err)
		}
	}()

	select {
	case <-sigChan:
		fmt.Println("Bot shutting down...")
	case <-ctx.Done():
		fmt.Println("Waiting for services to stop...")
	}
}
