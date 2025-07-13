package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/cmd"
	"github.com/Rakhulsr/go-ecommerce/app/configs"
	"github.com/Rakhulsr/go-ecommerce/app/routes"
)

func main() {

	if len(os.Args) > 1 {
		cmd.RunCli()
		return
	}
	env := configs.LoadENV
	log.Printf("DEBUG: RajaOngkir Base URL: %s", env.API_ONGKIR_BASE_URL)
	log.Printf("DEBUG: RajaOngkir API Key: %s", env.API_ONGKIR_KEY)
	rand.Seed(time.Now().UnixNano())

	db, err := configs.OpenConnection()
	if err != nil {
		log.Fatal("DB connection failed:", err)

	}
	log.Println("✅ Database connected.")

	log.Println("✅ Session store initialized.")
	router := routes.NewRouter(db)

	server := http.Server{
		Addr:    configs.LoadENV.Port,
		Handler: router,
	}

	log.Printf("🚀 Server starting on :%s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Println("failed to connecting to the server")
	}

}
