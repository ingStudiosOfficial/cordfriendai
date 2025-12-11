package structs

type EncryptedAPI struct {
	IV            string `bson:"iv" json:"iv"`
	EncryptedData string `bson:"encryptedData" json:"encryptedData"`
}
