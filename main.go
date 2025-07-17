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
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

var MidtransSnapClient snap.Client

func init() {

	MidtransSnapClient.New(configs.LoadENV.MIDTRANS_SERVER_KEY, midtrans.Sandbox)

	midtrans.ClientKey = configs.LoadENV.MIDTRANS_CLIENT_KEY

	midtrans.ServerKey = configs.LoadENV.MIDTRANS_SERVER_KEY

	midtrans.Environment = midtrans.Sandbox

	log.Println("✅ Midtrans Snap Client initialized.")
}

func main() {

	env := configs.LoadEnv()
	if len(os.Args) > 1 {
		cmd.RunCli()
		return
	}

	rand.Seed(time.Now().UnixNano())

	log.Printf("Loaded API_ONGKIR_KEY_KOMERCE: '%s'", env.API_ONGKIR_KEY_KOMERCE)
	if env.API_ONGKIR_KEY_KOMERCE == "" {
		log.Fatalf("API_ONGKIR_KEY_KOMERCE is empty! Please check your .env file.")
	}

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

	log.Printf("🚀 Server starting on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Println("failed to connecting to the server")
	}

}
