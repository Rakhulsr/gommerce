package main

import (
	"log"
	"net/http"

	"github.com/Rakhulsr/go-ecommerce/app/configs"
	"github.com/Rakhulsr/go-ecommerce/app/routes"
)

func main() {

	db, err := configs.OpenConnection()
	if err != nil {
		log.Fatal(err)
	}

	router := routes.NewRouter(db)

	server := http.Server{
		Addr:    configs.LoadENV.Port,
		Handler: router,
	}
	log.Println("Server up at port 8080")
	if err := server.ListenAndServe(); err != nil {
		log.Println("failed to connecting to the server")
	}

}
