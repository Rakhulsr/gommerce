package routes

import (
	"log"
	"net/http"

	"github.com/Rakhulsr/go-ecommerce/app/configs"
	"github.com/Rakhulsr/go-ecommerce/app/handlers"
	"github.com/Rakhulsr/go-ecommerce/app/handlers/admin"
	"github.com/Rakhulsr/go-ecommerce/app/middlewares"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/services"
	"github.com/Rakhulsr/go-ecommerce/app/utils/renderer"
	"github.com/Rakhulsr/go-ecommerce/app/utils/sessions"
	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"github.com/go-playground/validator/v10"
)

func NewRouter(db *gorm.DB) *mux.Router {
	router := mux.NewRouter()
	render := renderer.New()
	env := configs.LoadENV

	sessionKeys, err := configs.LoadSessionKeysFromEnv()
	if err != nil {
		log.Fatalf("Failed to load session keys for router initialization: %v", err)
	}

	productRepo := repositories.NewProductRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	cartItemRepo := repositories.NewCartItemRepository(db)
	cartRepo := repositories.NewCartRepository(db, cartItemRepo)
	sectionRepo := repositories.NewSectionRepository(db)
	userRepo := repositories.NewUserRepository(db)
	addressRepo := repositories.NewGormAddressRepository(db)

	cartSvc := services.NewCartService(cartRepo, cartItemRepo, productRepo)
	shippingSvc := services.NewRajaOngkirService()
	sessionStore := sessions.NewCookieSessionStore(sessionKeys.AuthKey, sessionKeys.EncKey)

	emailConfig := services.Config{
		Host:     env.EmailHost,
		Port:     env.EmailPort,
		Username: env.EmailUsername,
		Password: env.EmailPassword,
		From:     env.EmailFrom,
	}
	mailer := services.NewMailer(emailConfig)

	validate := validator.New()

	productHandler := handlers.NewProductHandler(productRepo, categoryRepo, render)
	homeHandler := handlers.NewHomeHandler(render, categoryRepo, productRepo)
	cartHandler := handlers.NewCartHandler(productRepo, cartRepo, render, cartItemRepo, shippingSvc)
	locationAPIHandler := handlers.NewLocationAPIHandler(shippingSvc, render)
	authHandler := handlers.NewAuthHandler(render, userRepo, cartRepo, sessionStore, mailer, validate)
	addressHandler := handlers.NewAddressHandler(render, addressRepo, userRepo, shippingSvc, validate)

	adminHandler := admin.NewAdminHandler(render, validate, productRepo, categoryRepo, sectionRepo, userRepo, cartRepo, cartItemRepo, *cartSvc)

	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("assets/css"))))
	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("assets/js"))))
	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("assets/images"))))

	router.Use(mux.MiddlewareFunc(middlewares.MethodOverrideMiddleware))
	router.Use(mux.MiddlewareFunc(middlewares.AuthAndCartSessionMiddleware(userRepo, cartRepo, sessionStore)))
	router.Use(mux.MiddlewareFunc(middlewares.CartCountMiddleware(cartRepo)))

	router.HandleFunc("/", homeHandler.Home).Methods("GET")
	router.HandleFunc("/products", productHandler.Products).Methods("GET")
	router.HandleFunc("/products/{slug}", productHandler.ProductDetail).Methods("GET")
	router.HandleFunc("/carts", cartHandler.GetCart).Methods("GET")
	router.HandleFunc("/carts/add", cartHandler.AddItemCart).Methods("POST")
	router.HandleFunc("/carts/update", cartHandler.UpdateCartItem).Methods("POST")
	router.HandleFunc("/carts/delete", cartHandler.DeleteCartItem).Methods("POST")
	router.HandleFunc("/login", authHandler.LoginGetHandler).Methods("GET")
	router.HandleFunc("/login", authHandler.LoginPostHandler).Methods("POST")
	router.HandleFunc("/register", authHandler.RegisterGetHandler).Methods("GET")
	router.HandleFunc("/register", authHandler.RegisterPostHandler).Methods("POST")
	router.HandleFunc("/forgot-password", authHandler.ForgotPasswordGetHandler).Methods("GET")
	router.HandleFunc("/forgot-password", authHandler.ForgotPasswordPostHandler).Methods("POST")
	router.HandleFunc("/verify-otp", authHandler.VerifyOTPGetHandler).Methods("GET")
	router.HandleFunc("/verify-otp", authHandler.VerifyOTPPostHandler).Methods("POST")
	router.HandleFunc("/reset-password", authHandler.ResetPasswordGetHandler).Methods("GET")
	router.HandleFunc("/reset-password", authHandler.ResetPasswordPostHandler).Methods("POST")

	authenticated := router.PathPrefix("/").Subrouter()
	authenticated.Use(mux.MiddlewareFunc(middlewares.AuthRequiredMiddleware))
	authenticated.HandleFunc("/provinces", locationAPIHandler.GetProvincesAPI).Methods("GET")
	authenticated.HandleFunc("/cities", locationAPIHandler.GetCitiesAPI).Methods("GET")
	authenticated.HandleFunc("/calculate-shipping-cost", locationAPIHandler.CalculateShippingCostAPI).Methods("POST")

	authenticated.HandleFunc("/profile", authHandler.ProfileHandler).Methods("GET")
	authenticated.HandleFunc("/profile/edit", authHandler.UpdateProfilePost).Methods("PUT", "POST")
	authenticated.HandleFunc("/profile/edit", authHandler.UpdateProfilePage).Methods("GET")
	authenticated.HandleFunc("/logout", authHandler.LogoutHandler).Methods("POST")

	authenticated.HandleFunc("/addresses", addressHandler.GetAddressesPage).Methods("GET")
	authenticated.HandleFunc("/addresses/add", addressHandler.AddAddressPage).Methods("GET")
	authenticated.HandleFunc("/addresses/add", addressHandler.AddAddressPost).Methods("POST")
	authenticated.HandleFunc("/addresses/edit/{id}", addressHandler.EditAddressPage).Methods("GET")
	authenticated.HandleFunc("/addresses/edit/{id}", addressHandler.EditAddressPost).Methods("POST")
	authenticated.HandleFunc("/addresses/delete/{id}", addressHandler.DeleteAddressPost).Methods("POST")
	authenticated.HandleFunc("/addresses/set-primary/{id}", addressHandler.SetPrimaryAddressPost).Methods("POST")

	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middlewares.AuthRequiredMiddleware)
	adminRouter.Use(middlewares.AdminAuthMiddleware(userRepo))
	adminRouter.HandleFunc("/dashboard/apply-discount", adminHandler.ApplyGlobalDiscountPost).Methods("POST")

	adminRouter.HandleFunc("/dashboard", adminHandler.GetDashboard).Methods("GET")
	adminRouter.HandleFunc("/products", adminHandler.GetProductsPage).Methods("GET")
	adminRouter.HandleFunc("/products/add", adminHandler.AddProductPage).Methods("GET")
	adminRouter.HandleFunc("/products/add", adminHandler.AddProductPost).Methods("POST")
	adminRouter.HandleFunc("/products/edit/{id}", adminHandler.EditProductPage).Methods("GET")
	adminRouter.HandleFunc("/products/edit/{id}", adminHandler.EditProductPost).Methods("POST", "PUT")
	adminRouter.HandleFunc("/products/delete/{id}", adminHandler.DeleteProductPost).Methods("POST", "DELETE")

	adminRouter.HandleFunc("/categories", adminHandler.GetCategoriesPage).Methods("GET")
	adminRouter.HandleFunc("/categories/add", adminHandler.AddCategoryPage).Methods("GET")
	adminRouter.HandleFunc("/categories/add", adminHandler.AddCategoryPost).Methods("POST")
	adminRouter.HandleFunc("/categories/edit/{id}", adminHandler.EditCategoryPage).Methods("GET")
	adminRouter.HandleFunc("/categories/edit/{id}", adminHandler.EditCategoryPost).Methods("POST")
	adminRouter.HandleFunc("/categories/delete/{id}", adminHandler.DeleteCategoryPost).Methods("POST")

	adminRouter.HandleFunc("/users", adminHandler.GetUsersPage).Methods("GET")
	adminRouter.HandleFunc("/users/add", adminHandler.AddUserPage).Methods("GET")
	adminRouter.HandleFunc("/users/add", adminHandler.AddUserPost).Methods("POST")
	adminRouter.HandleFunc("/users/edit/{id}", adminHandler.EditUserPage).Methods("GET")
	adminRouter.HandleFunc("/users/edit/{id}", adminHandler.EditUserPost).Methods("POST", "PUT")
	adminRouter.HandleFunc("/users/delete/{id}", adminHandler.DeleteUserPost).Methods("POST", "DELETE")

	return router
}
