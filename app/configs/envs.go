package configs

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type ENV struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	Port       string
}

func LoadEnv() ENV {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current DIR : %v", err)
	}

	fmt.Printf("Current DIR is : %s\n", cwd)

	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: No .env file found ")
	}

	return ENV{
		DBHost:     os.Getenv("DB_HOST"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBPort:     os.Getenv("DB_PORT"),
		Port:       os.Getenv("APP_PORT"),
	}

}

var LoadENV = LoadEnv()
