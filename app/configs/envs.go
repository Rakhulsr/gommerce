package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type ENV struct {
	DBHost              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBPort              string
	Port                string
	SESSION_KEY         string
	API_ONGKIR_BASE_URL string
	API_ONGKIR_KEY      string
	API_ONGKIR_ORIGIN   string
}

func LoadEnv() ENV {
	// cwd, err := os.Getwd()
	// if err != nil {
	// 	log.Fatalf("Failed to get current DIR : %v", err)
	// }

	// fmt.Printf("Current DIR is : %s\n", cwd)

	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: No .env file found ")
	}

	return ENV{
		DBHost:              os.Getenv("DB_HOST"),
		DBUser:              os.Getenv("DB_USER"),
		DBPassword:          os.Getenv("DB_PASSWORD"),
		DBName:              os.Getenv("DB_NAME"),
		DBPort:              os.Getenv("DB_PORT"),
		Port:                os.Getenv("APP_PORT"),
		SESSION_KEY:         os.Getenv("SESSION_KEY"),
		API_ONGKIR_BASE_URL: os.Getenv("API_ONGKIR_BASE_URL"),
		API_ONGKIR_KEY:      os.Getenv("API_ONGKIR_KEY"),
		API_ONGKIR_ORIGIN:   os.Getenv("API_ONGKIR_ORIGIN"),
	}

}

var LoadENV = LoadEnv()
