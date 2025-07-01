package routes

import (
	"net/http"

	"github.com/Rakhulsr/go-ecommerce/app/handlers"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/", handlers.Home).Methods("GET")

	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("assets/css"))))

	return router

}
