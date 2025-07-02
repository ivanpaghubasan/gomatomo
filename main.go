package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ivanpaghubasan/go-matomo/matomo"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := &matomo.MatomoClient{
		BaseURL:    os.Getenv("MATOMO_URL"),
		ScriptHost: os.Getenv("MATOMO_SCRIPT_URL"),
		TokenAuth:  os.Getenv("MATOMO_TOKEN"),
	}

	userLogin := "ivan.paghubasan"

	siteID, login, password, script, err := client.ProvisionTelemetry(
		userLogin,
		fmt.Sprintf("%s@example.com", userLogin),
		"SnapToApp",
		"https://snaptoapp.com",
	)

	if err != nil {
		log.Fatalf("Provisioning failed: %v", err)
	}

	fmt.Println("Matomo site ID:", siteID)
	fmt.Println("Matomo login:", login)
	if password != "" {
		fmt.Println("Initial password:", password)
	}
	fmt.Println("Tracking script:\n", script)
}
