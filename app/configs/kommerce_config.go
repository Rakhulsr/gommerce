package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type KomerceAPIConfig struct {
	API_ONGKIR_BASE_URL_KOMERCE string
	API_ONGKIR_KEY_KOMERCE      string
	API_ONGKIR_ORIGIN_KOMERCE   string
}

func LoadKomerceAPIEnv() KomerceAPIConfig {

	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: No .env file found when loading Komerce API config.")
	}

	return KomerceAPIConfig{
		API_ONGKIR_BASE_URL_KOMERCE: os.Getenv("API_ONGKIR_BASE_URL"),
		API_ONGKIR_KEY_KOMERCE:      os.Getenv("API_ONGKIR_KEY"),
		API_ONGKIR_ORIGIN_KOMERCE:   os.Getenv("API_ONGKIR_ORIGIN_KOMERCE"),
	}
}

var KomerceAPIEnv = LoadKomerceAPIEnv()
