package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"google.golang.org/api/option"
)

type EncryptedAPI struct {
	IV            string `bson:"iv" json:"iv"`
	EncryptedData string `bson:"encryptedData" json:"encryptedData"`
}

type Bot struct {
	Name          string       `bson:"name"`
	Persona       string       `bson:"persona"`
	ServerID      string       `bson:"server_id"`
	UserID        string       `bson:"user_id"`
	GoogleAIAPI   EncryptedAPI `bson:"google_ai_api"`
	Image         string       `bson:"image_id"`
	Conversations []string     `bson:"conversations"`
}

var botsCollection *mongo.Collection

var dg *discordgo.Session

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error while loading .env file:", err)
	}

	var mongoClient = connectToMongo()

	var databaseName = "cordfriendAI"

	botsCollection = mongoClient.Database(databaseName).Collection("bots")

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

		var commands = []*discordgo.ApplicationCommand{
			{
				Name:        "load-name",
				Description: "Sets the nickname to the saved name from the Cordfriend AI dashboard.",
				Type:        discordgo.ChatApplicationCommand,
			},
			{
				Name:        "load-avatar",
				Description: "Sets the avatar to the saved avatar from the Cordfriend AI dashboard.",
				Type:        discordgo.ChatApplicationCommand,
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

	dg.AddHandler(handleCommand)
	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

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
}

func connectToMongo() *mongo.Client {
	var connectionString = os.Getenv("MONGODB_CONNECTION_STRING")
	if connectionString == "" {
		fmt.Println("MongoDB connection string not set.")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(connectionString).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	return client
}

func handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "load-name":
		updateBotNickname(s, i.GuildID, i)

	case "load-avatar":
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		if err != nil {
			fmt.Println("Failed to defer interaction:", err)
			return
		}

		err = updateBotImage(s, i.GuildID)

		responseContent := "Avatar updated successfully!"
		if err != nil {
			responseContent = fmt.Sprintf("Failed to update avatar: %v", err)
		}

		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &responseContent,
		})
		if err != nil {
			fmt.Println("Failed to edit interaction response:", err)
		}
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
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
		fmt.Println("Bot mentioned, responding.")

		var response = requestGenAi(m)
		fmt.Println("Returning response:", response)

		// Send response back to channel
		s.ChannelMessageSend(m.ChannelID, truncateString(response, 2000))
	}
}

func requestGenAi(m *discordgo.MessageCreate) string {
	fmt.Println("Generating response...")

	apiKey, err := fetchApiKey(m)
	if err != nil {
		fmt.Println("Error while fetching API key:", err)
		return "Could not fetch API key for this server."
	}
	if apiKey == "" {
		fmt.Println("Could not fetch API key.")
		return "Could not fetch API key for this server."
	}

	var conversationsString string

	conversations, err := fetchConversations(m.GuildID)
	if err != nil {
		fmt.Print("Error while fetching conversations:", err)
		conversationsString = "No conversations stored in history yet."
	}

	conversationsString = strings.Join(conversations, "\n")

	fmt.Println("Conversations:", conversationsString)

	var systemInstructions string
	fetchedInstructions, err := fetchBotPersona(m.GuildID)
	if err != nil {
		fmt.Println("Error while fetching bot persona:", err)
	}
	if fetchedInstructions == "" {
		systemInstructions = "You are a helpful Discord bot. Please be as concise as possible - but still give helpful information."
	} else {
		systemInstructions = fetchedInstructions
	}

	var sentUser = m.Author.DisplayName()
	fmt.Println("User who sent message:", sentUser)

	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))

	if err != nil {
		log.Fatal("Error creating new Gemini client:", err)
	}

	var promptToSend = "Conversation history: " + conversationsString + "System message: " + systemInstructions + "User '" + sentUser + "' sent the message: " + m.Content

	resp, err := client.GenerativeModel("gemini-2.5-flash").GenerateContent(
		ctx,
		genai.Text(promptToSend),
	)
	if err != nil {
		fmt.Println("Error while generating content:", err)
	}

	var response string = ""

	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				response += fmt.Sprintf("%v", part)
			}
		}
	}

	if response != "" {
		addConversations(m.GuildID, response)
	}

	return response
}

// pkcs7 unpad
func pkcs7Unpad(b []byte) ([]byte, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("pkcs7: invalid padding size")
	}
	padLen := int(b[len(b)-1])
	if padLen == 0 || padLen > len(b) {
		return nil, fmt.Errorf("pkcs7: invalid padding")
	}
	for i := len(b) - padLen; i < len(b); i++ {
		if b[i] != byte(padLen) {
			return nil, fmt.Errorf("pkcs7: invalid padding")
		}
	}
	return b[:len(b)-padLen], nil
}

// decrypt AES-256-CBC
func decryptAES256CBC(encryptedHex, ivHex string, key []byte) (string, error) {
	ciphertext, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("hex decode ciphertext: %w", err)
	}
	iv, err := hex.DecodeString(ivHex)
	if err != nil {
		return "", fmt.Errorf("hex decode iv: %w", err)
	}
	if len(iv) != aes.BlockSize {
		return "", fmt.Errorf("invalid IV length")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext not multiple of block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	unpadded, err := pkcs7Unpad(plaintext)
	if err != nil {
		return "", fmt.Errorf("unpad: %w", err)
	}
	return string(unpadded), nil
}

func fetchApiKey(m *discordgo.MessageCreate) (string, error) {
	var fetchedBot Bot

	filter := bson.M{"server_id": m.GuildID}
	err := botsCollection.FindOne(context.TODO(), filter).Decode(&fetchedBot)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("no document found for guild %s", m.GuildID)
		}
		return "", fmt.Errorf("mongo find error: %w", err)
	}

	// AES key from env (must be 32 bytes for AES-256)
	keyHex := os.Getenv("CRYPTO_SECRET_KEY")
	key, err := hex.DecodeString(keyHex) // decode hex -> raw bytes
	if err != nil {
		return "", fmt.Errorf("invalid AES key hex: %w", err)
	}
	if len(key) != 32 {
		return "", fmt.Errorf("AES key must be 32 bytes, got %d", len(key))
	}

	// Decrypt
	plain, err := decryptAES256CBC(
		fetchedBot.GoogleAIAPI.EncryptedData,
		fetchedBot.GoogleAIAPI.IV,
		key,
	)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt Google AI API key: %w", err)
	}

	return plain, nil
}

func fetchNickname(guildID string) (string, error) {
	var settings Bot
	filter := bson.M{"server_id": guildID}
	err := botsCollection.FindOne(context.TODO(), filter).Decode(&settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil // no nickname set
		}
		return "", err
	}

	fmt.Println("Nickname fetched successfully:", settings.Name)

	return settings.Name, nil
}

func updateBotNickname(s *discordgo.Session, guildID string, i *discordgo.InteractionCreate) error {
	fmt.Println("Update nickname command called.")

	nickname, err := fetchNickname(guildID)
	if err != nil {
		return fmt.Errorf("failed to fetch nickname: %w", err)
	}
	if nickname == "" {
		return nil
	}

	fmt.Println("Changing nickname for guild ID:", guildID)

	// Fix: Use "@me" to refer to the bot itself
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

func fetchBotPersona(guildID string) (string, error) {
	var settings Bot
	filter := bson.M{"server_id": guildID}
	err := botsCollection.FindOne(context.TODO(), filter).Decode(&settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}

	fmt.Println("Persona fetched successfully:", settings.Persona)

	return settings.Persona, nil
}

func truncateString(s string, maxLength int) string {
	// Convert the string to a slice of runes to handle multi-byte characters correctly.
	runes := []rune(s)

	if len(runes) <= maxLength {
		return s
	}

	// Return runes back as string
	return string(runes[:maxLength])
}

func fetchBotImage(guildID string) (string, error) {
	var settings Bot
	filter := bson.M{"server_id": guildID}
	err := botsCollection.FindOne(context.TODO(), filter).Decode(&settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}

	fmt.Println("Image fetched successfully:", settings.Image)

	return "https://cordfriendai-server.onrender.com/api/bot/image-download/" + settings.Image, nil
}

func updateBotImage(s *discordgo.Session, guildID string) error {
	var contentType string
	var base64img string

	avatarUrl, err := fetchBotImage(guildID)
	if err != nil {
		fmt.Println("Error while fetching bot image:", err)
		return err
	}

	if avatarUrl == "" {
		fmt.Println("Could not fetch avatar URL.")
		return fmt.Errorf("no avatar URL configured")
	}

	fmt.Println("Avatar URL:", avatarUrl)

	resp, err := http.Get(avatarUrl)
	if err != nil {
		fmt.Println("Error while fetching avatar:", err)
		return err
	}
	defer resp.Body.Close()

	img, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading avatar:", err)
		return err
	}

	fmt.Printf("Original image size: %d bytes\n", len(img))

	contentType = http.DetectContentType(img)

	validTypes := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !validTypes[contentType] {
		fmt.Println("Unsupported image format:", contentType)
		return fmt.Errorf("unsupported image format: %s", contentType)
	}

	base64img = base64.StdEncoding.EncodeToString(img)
	fmt.Printf("Base64 length: %d\n", len(base64img))

	avatar := fmt.Sprintf("data:%s;base64,%s", contentType, base64img)

	user, err := s.UserUpdate("", avatar, "")
	if err != nil {
		fmt.Println("Error updating avatar:", err)
		return err
	}

	fmt.Printf("Avatar updated successfully! Avatar hash: %s\n", user.Avatar)
	return nil
}

func fetchConversations(guildID string) ([]string, error) {
	var settings Bot
	filter := bson.M{"server_id": guildID}
	err := botsCollection.FindOne(context.TODO(), filter).Decode(&settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	fmt.Println("Conversations fetched successfully:", settings.Conversations)

	return settings.Conversations, nil
}

func addConversations(guildID string, conversation string) error {
	filter := bson.M{"server_id": guildID}
	update := bson.M{
		"$push": bson.M{
			"conversations": bson.M{
				"$each":     []string{conversation},
				"$position": 0, // Prepend at index 0
			},
		},
	}

	_, err := botsCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println("Error while adding to conversation history:", err)
		return err
	}

	return nil
}
