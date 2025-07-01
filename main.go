package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Rakhulsr/go-ecommerce/app/cmd"
	"github.com/Rakhulsr/go-ecommerce/app/configs"
	"github.com/Rakhulsr/go-ecommerce/app/routes"
)

func main() {

	if len(os.Args) > 1 {
		cmd.RunCli()
		return
	}

	db, err := configs.OpenConnection()
	if err != nil {
		log.Fatal("DB connection failed:", err)

	}

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("assets/css"))))

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
