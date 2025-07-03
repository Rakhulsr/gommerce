package routes

import (
	"net/http"

	"github.com/Rakhulsr/go-ecommerce/app/handlers"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/utils/renderer"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *mux.Router {
	router := mux.NewRouter()
	render := renderer.New()

	productRepo := repositories.NewProductRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	productHandler := handlers.NewProductHandler(productRepo, categoryRepo, render)

	homeHandler := handlers.NewHomeHandler(render, categoryRepo, productRepo)

	router.HandleFunc("/", homeHandler.Home).Methods("GET")
	router.HandleFunc("/products", productHandler.Products).Methods("GET")
	router.HandleFunc("/products/{slug}", productHandler.ProductDetail).Methods("GET")

	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("assets/css"))))

	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("assets/js"))))

	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("assets/images"))))

	return router

}
