package structs

type User struct {
	Name    string `bson:"name"`
	Message string `bson:"message"`
}

type Conversation struct {
	User User   `bson:"user"`
	Bot  string `bson:"bot"`
}

type Bot struct {
	Name              string         `bson:"name"`
	Persona           string         `bson:"persona"`
	ServerID          string         `bson:"server_id"`
	UserID            string         `bson:"user_id"`
	GoogleAIAPI       EncryptedAPI   `bson:"google_ai_api"`
	OpenWeatherMapAPI EncryptedAPI   `bson:"openweathermap_api"`
	Image             string         `bson:"image_id"`
	Conversations     []Conversation `bson:"conversations"`
}
