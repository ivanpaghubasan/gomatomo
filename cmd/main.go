package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ivanpaghubasan/gomatomo/matomo"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := matomo.NewClient(os.Getenv("MATOMO_URL"), os.Getenv("MATOMO_TOKEN"))
	list, err := client.GetMockSessionsByDevice()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(list)

	list2, err := client.GetMockAudienceByCountry()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(list2)
	var appID int64 = 543
	appUrl := fmt.Sprintf("https://dapp.snaptoapp.com/app/%s/landing-page", fmt.Sprint(appID))
	fmt.Println("App URL ", appUrl)
	// userLogin := "john.doe"

	// siteID, login, password, script, err := client.ProvisionTelemetry(
	// 	userLogin,
	// 	fmt.Sprintf("%s@example.com", userLogin),
	// 	"John Doe Website",
	// 	"http://johndoe.com",
	// )

	// if err != nil {
	// 	log.Fatalf("Provisioning failed: %v", err)
	// }

	// fmt.Println("Matomo site ID:", siteID)
	// fmt.Println("Matomo user login:", login)
	// if password != "" {
	// 	fmt.Println("Initial password:", password)
	// }
	// fmt.Println("Tracking script:\n", script)
}
