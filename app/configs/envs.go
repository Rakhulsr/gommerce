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
	JWTSecret           string
	AppAuthKey          string
	AppEncKey           string
	EmailHost           string
	EmailPort           string
	EmailUsername       string
	EmailPassword       string
	EmailFrom           string
}

func LoadEnv() ENV {

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
		JWTSecret:           os.Getenv("JWT_SECRET"),
		AppAuthKey:          os.Getenv("APP_AUTH_KEY"),
		AppEncKey:           os.Getenv("APP_ENC_KEY"),
		EmailHost:           os.Getenv("EMAIL_HOST"),
		EmailPort:           os.Getenv("EMAIL_PORT"),
		EmailUsername:       os.Getenv("EMAIL_USERNAME"),
		EmailPassword:       os.Getenv("EMAIL_PASSWORD"),
		EmailFrom:           os.Getenv("EMAIL_USERNAME"),
	}

}

var LoadENV = LoadEnv()
