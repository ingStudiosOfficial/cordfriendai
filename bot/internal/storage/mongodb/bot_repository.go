package mongodb

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	"bot/internal/decryption"
	"bot/internal/structs"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type BotRepository struct {
	collection *mongo.Collection
}

func NewBotRepository(db *mongo.Database) *BotRepository {
	return &BotRepository{
		collection: db.Collection("bots"),
	}
}

func (r *BotRepository) FetchConversations(guildID string) ([]structs.Conversation, error) {
	var settings structs.Bot
	filter := bson.M{"server_id": guildID}
	err := r.collection.FindOne(context.TODO(), filter).Decode(&settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	fmt.Println("Conversations fetched successfully:", settings.Conversations)

	return settings.Conversations, nil
}

func (r *BotRepository) AddConversations(guildID string, userConv structs.User, botConv string) error {
	var conversation structs.Conversation = structs.Conversation{
		User: userConv,
		Bot:  botConv,
	}

	filter := bson.M{"server_id": guildID}
	update := bson.M{
		"$push": bson.M{
			"conversations": bson.M{
				"$each":     []structs.Conversation{conversation},
				"$position": 0, // Prepend at index 0
			},
		},
	}

	_, err := r.collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println("Error while adding to conversation history:", err)
		return err
	}

	return nil
}

func (r *BotRepository) FetchNickname(guildID string) (string, error) {
	var settings structs.Bot
	filter := bson.M{"server_id": guildID}
	err := r.collection.FindOne(context.TODO(), filter).Decode(&settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil // no nickname set
		}
		return "", err
	}

	fmt.Println("Nickname fetched successfully:", settings.Name)

	return settings.Name, nil
}

func (r *BotRepository) FetchBotPersona(guildID string) (string, error) {
	var settings structs.Bot
	filter := bson.M{"server_id": guildID}
	err := r.collection.FindOne(context.TODO(), filter).Decode(&settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}

	fmt.Println("Persona fetched successfully:", settings.Persona)

	return settings.Persona, nil
}

func (r *BotRepository) FetchApiKey(serverId string) (string, error) {
	var fetchedBot structs.Bot

	filter := bson.M{"server_id": serverId}
	err := r.collection.FindOne(context.TODO(), filter).Decode(&fetchedBot)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("no document found for guild %s", serverId)
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
	plain, err := decryption.DecryptAES256CBC(
		fetchedBot.GoogleAIAPI.EncryptedData,
		fetchedBot.GoogleAIAPI.IV,
		key,
	)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt Google AI API key: %w", err)
	}

	return plain, nil
}
