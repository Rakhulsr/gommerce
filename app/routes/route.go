// routes/routes.go (Updated)

package routes

import (
	"net/http"

	"github.com/Rakhulsr/go-ecommerce/app/handlers"
	"github.com/Rakhulsr/go-ecommerce/app/middlewares"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/services"
	"github.com/Rakhulsr/go-ecommerce/app/utils/renderer"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *mux.Router {
	router := mux.NewRouter()
	render := renderer.New()

	productRepo := repositories.NewProductRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	cartRepo := repositories.NewCartRepository(db)
	cartItemRepo := repositories.NewCartItemRepository(db)

	shippingSvc := services.NewRajaOngkirService()

	productHandler := handlers.NewProductHandler(productRepo, categoryRepo, render)
	homeHandler := handlers.NewHomeHandler(render, categoryRepo, productRepo)
	cartHandler := handlers.NewCartHandler(productRepo, cartRepo, *render, cartItemRepo, shippingSvc)
	locationAPIHandler := handlers.NewLocationAPIHandler(shippingSvc, render)

	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("assets/css"))))
	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("assets/js"))))
	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("assets/images"))))

	router.Use(mux.MiddlewareFunc(middlewares.MethodOverrideMiddleware))
	router.Use(mux.MiddlewareFunc(middlewares.SessionManagerMiddleware))

	router.Use(mux.MiddlewareFunc(middlewares.CartCountMiddleware(cartRepo)))

	router.HandleFunc("/", homeHandler.Home).Methods("GET")
	router.HandleFunc("/products", productHandler.Products).Methods("GET")
	router.HandleFunc("/products/{slug}", productHandler.ProductDetail).Methods("GET")

	router.HandleFunc("/carts", cartHandler.GetCart).Methods("GET")
	router.HandleFunc("/carts/add", cartHandler.AddItemCart).Methods("POST")
	router.HandleFunc("/carts/update", cartHandler.UpdateCartItem).Methods("POST")
	router.HandleFunc("/carts/delete", cartHandler.DeleteCartItem).Methods("POST")

	router.HandleFunc("/provinces", locationAPIHandler.GetProvincesAPI).Methods("GET")
	router.HandleFunc("/cities", locationAPIHandler.GetCitiesAPI).Methods("GET")
	router.HandleFunc("/calculate-shipping-cost", locationAPIHandler.CalculateShippingCostAPI).Methods("POST")

	return router
}
