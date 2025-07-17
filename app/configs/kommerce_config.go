package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// KomerceAPIConfig holds environment variables specific to Komerce API V2
// Ini adalah duplikat yang disesuaikan untuk Komerce API RajaOngkir V2.
type KomerceAPIConfig struct {
	API_ONGKIR_BASE_URL_KOMERCE string
	API_ONGKIR_KEY_KOMERCE      string
	API_ONGKIR_ORIGIN_KOMERCE   string // Jika Anda memiliki origin default untuk Komerce
}

// LoadKomerceAPIEnv loads environment variables specific to Komerce API V2.
// Fungsi ini akan mencari variabel lingkungan dengan nama yang berbeda
// agar tidak bertabrakan dengan konfigurasi RajaOngkir standar Anda.
func LoadKomerceAPIEnv() KomerceAPIConfig {
	// Memuat .env file lagi di sini untuk memastikan variabel tersedia,
	// terutama jika fungsi ini dipanggil secara independen.
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
