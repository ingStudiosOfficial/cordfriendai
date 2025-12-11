package serverping

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func SendUptime(durationSinceStart string) {
	fmt.Println("Attempting to ping server of uptime: " + durationSinceStart)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error while loading .env file:", err)
	}

	type UptimeReport struct {
		Uptime string `json:"uptime"`
	}

	report := UptimeReport{
		Uptime: durationSinceStart,
	}

	jsonReport, err := json.Marshal(report)
	if err != nil {
		fmt.Println("Error while creating report:", err)
		return
	}

	resp, err := http.Post(
		os.Getenv("SERVER_TO_PING"),
		"application/json",
		bytes.NewBuffer(jsonReport),
	)

	if err != nil {
		fmt.Println("Error while sending uptime:", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)

		fmt.Println("Ping failed:", bodyString)

		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading response body:", err)
	}

	bodyString := string(bodyBytes)

	fmt.Println("Successfully pinged server and reported uptime:", bodyString)
}
