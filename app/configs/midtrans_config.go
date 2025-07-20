package configs

import (
	"log"
	"os"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

var (
	snapClient            snap.Client
	coreAPIClient         coreapi.Client
	isMidtransInitialized bool
)

func InitMidtransClient() {
	if isMidtransInitialized {
		return
	}

	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	if serverKey == "" {
		log.Fatal("MIDTRANS_SERVER_KEY environment variable not set.")
	}

	env := midtrans.Sandbox
	if LoadEnv().APP_ENV == "production" {
		env = midtrans.Production
	}

	snapClient.New(serverKey, env)
	log.Println("Midtrans Snap Client initialized.")

	coreAPIClient.New(serverKey, env)
	log.Println("Midtrans CoreAPI Client initialized.")

	isMidtransInitialized = true
}

func GetMidtransSnapClient() snap.Client {
	if !isMidtransInitialized {
		log.Fatal("Midtrans client not initialized. Call InitMidtransClient() first.")
	}
	return snapClient
}

func GetMidtransCoreAPIClient() coreapi.Client {
	if !isMidtransInitialized {
		log.Fatal("Midtrans client not initialized. Call InitMidtransClient() first.")
	}
	return coreAPIClient
}

func GetAppBaseURL() string {
	baseURL := LoadEnv().APP_URL
	if baseURL == "" {
		log.Fatal("APP_BASE_URL environment variable not set. This is required for Midtrans callbacks.")
	}
	return baseURL
}
