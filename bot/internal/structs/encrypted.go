package structs

type EncryptedAPI struct {
	IV            string `bson:"iv" json:"iv"`
	EncryptedData string `bson:"encryptedData" json:"encryptedData"`
}

type Conversation struct {
	User string `bson:"user"`
	Bot  string `bson:"bot"`
}

type Bot struct {
	Name          string         `bson:"name"`
	Persona       string         `bson:"persona"`
	ServerID      string         `bson:"server_id"`
	UserID        string         `bson:"user_id"`
	GoogleAIAPI   EncryptedAPI   `bson:"google_ai_api"`
	Image         string         `bson:"image_id"`
	Conversations []Conversation `bson:"conversations"`
}
