package routes

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Rakhulsr/go-ecommerce/app/configs"
	"github.com/Rakhulsr/go-ecommerce/app/handlers"
	"github.com/Rakhulsr/go-ecommerce/app/handlers/admin"
	"github.com/Rakhulsr/go-ecommerce/app/middlewares"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/services"
	"github.com/Rakhulsr/go-ecommerce/app/utils/renderer"
	"github.com/Rakhulsr/go-ecommerce/app/utils/sessions"
	"github.com/gorilla/mux"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *mux.Router {
	configs.InitMidtransClient()
	env := configs.LoadEnv()

	router := mux.NewRouter()
	render := renderer.New()

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
	orderRepo := repositories.NewOrderRepository(db)
	orderItemRepo := repositories.NewOrderItemRepository(db)
	orderCustomerRepo := repositories.NewOrderCustomerRepository(db)

	cartSvc := services.NewCartService(cartRepo, cartItemRepo, productRepo, db)
	komerceShippingSvc := services.NewKomerceRajaOngkirClient(env.API_ONGKIR_KEY_KOMERCE)

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
	checkoutSvc := services.NewCheckoutService(db, cartRepo, cartItemRepo, productRepo, userRepo, addressRepo, orderRepo, orderItemRepo, orderCustomerRepo)

	originID, _ := strconv.Atoi(env.API_ONGKIR_ORIGIN)

	productHandler := handlers.NewProductHandler(productRepo, categoryRepo, render)
	homeHandler := handlers.NewHomeHandler(render, categoryRepo, productRepo)
	komerceCartHandler := handlers.NewKomerceCartHandler(productRepo, cartRepo, render, cartItemRepo, komerceShippingSvc, userRepo, addressRepo, cartSvc, originID)

	authHandler := handlers.NewAuthHandler(render, userRepo, cartRepo, sessionStore, mailer, validate)
	// Inisialisasi KomerceAddressHandler tanpa locationRepo
	komerceAddressHandler := handlers.NewKomerceAddressHandler(render, addressRepo, userRepo, komerceShippingSvc, validate)

	adminHandler := admin.NewAdminHandler(render, validate, productRepo, categoryRepo, sectionRepo, userRepo, cartRepo, cartItemRepo, *cartSvc)
	komerceCheckoutHandler := handlers.NewKomerceCheckoutHandler(render, validate, checkoutSvc, cartRepo, userRepo, orderRepo, productRepo, db, komerceShippingSvc, addressRepo, sessionStore)

	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("assets/css"))))
	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("assets/js"))))
	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("assets/images"))))

	router.Use(mux.MiddlewareFunc(middlewares.MethodOverrideMiddleware))
	router.Use(mux.MiddlewareFunc((middlewares.ContentSecurityPolicyMiddleware)))
	router.Use(mux.MiddlewareFunc(middlewares.AuthAndCartSessionMiddleware(userRepo, cartRepo, sessionStore)))
	router.Use(mux.MiddlewareFunc(middlewares.CartCountMiddleware(cartRepo)))

	router.HandleFunc("/cart-count", komerceCartHandler.GetCartCount).Methods("GET")

	router.HandleFunc("/", homeHandler.Home).Methods("GET")
	router.HandleFunc("/products", productHandler.Products).Methods("GET")
	router.HandleFunc("/products/{slug}", productHandler.ProductDetail).Methods("GET")

	router.HandleFunc("/carts", komerceCartHandler.GetCart).Methods("GET")
	router.HandleFunc("/carts/add", komerceCartHandler.AddItemCart).Methods("POST")
	router.HandleFunc("/carts/update", komerceCartHandler.UpdateCartItem).Methods("POST")
	router.HandleFunc("/carts/delete", komerceCartHandler.DeleteCartItem).Methods("POST", "DELETE")

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

	apiKomerceRouter := router.PathPrefix("/api/komerce").Subrouter()
	apiKomerceRouter.HandleFunc("/calculate-shipping-cost", komerceCartHandler.CalculateShippingCost).Methods("POST")

	// --- PERUBAHAN KRUSIAL DI SINI ---
	// Hapus routes lama untuk lokasi granular
	// apiKomerceRouter.HandleFunc("/provinces", komerceAddressHandler.GetProvinces).Methods("GET")
	// apiKomerceRouter.HandleFunc("/cities", komerceAddressHandler.GetCitiesByProvince).Methods("GET")
	// apiKomerceRouter.HandleFunc("/districts", komerceAddressHandler.GetDistrictsByCity).Methods("GET")
	// apiKomerceRouter.HandleFunc("/subdistricts", komerceAddressHandler.GetSubdistrictsByDistrict).Methods("GET")

	// Tambahkan route baru untuk pencarian lokasi (autocomplete)
	apiKomerceRouter.HandleFunc("/search-destinations", komerceAddressHandler.SearchDomesticDestinationsHandler).Methods("GET")
	// --- AKHIR PERUBAHAN KRUSIAL ---

	authenticated.HandleFunc("/profile", authHandler.ProfileHandler).Methods("GET")
	authenticated.HandleFunc("/profile/edit", authHandler.UpdateProfilePost).Methods("POST", "PUT")
	authenticated.HandleFunc("/profile/edit", authHandler.UpdateProfilePage).Methods("GET")
	authenticated.HandleFunc("/logout", authHandler.LogoutHandler).Methods("POST")

	authenticated.HandleFunc("/addresses", komerceAddressHandler.GetAddressesPage).Methods("GET")
	authenticated.HandleFunc("/addresses/add", komerceAddressHandler.AddAddressPage).Methods("GET")
	authenticated.HandleFunc("/addresses/add", komerceAddressHandler.AddAddressPost).Methods("POST")
	authenticated.HandleFunc("/addresses/edit/{id}", komerceAddressHandler.EditAddressPage).Methods("GET")
	authenticated.HandleFunc("/addresses/edit/{id}", komerceAddressHandler.EditAddressPost).Methods("POST")
	authenticated.HandleFunc("/addresses/delete/{id}", komerceAddressHandler.DeleteAddressPost).Methods("POST")
	authenticated.HandleFunc("/addresses/set-primary/{id}", komerceAddressHandler.SetPrimaryAddressPost).Methods("POST")

	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middlewares.AuthRequiredMiddleware)
	adminRouter.Use(middlewares.AdminAuthMiddleware(userRepo))
	adminRouter.HandleFunc("/dashboard/apply-discount", adminHandler.ApplyGlobalDiscountPost).Methods("POST")

	adminRouter.HandleFunc("/dashboard", adminHandler.GetDashboard).Methods("GET")
	adminRouter.HandleFunc("/products", adminHandler.GetProductsPage).Methods("GET")
	adminRouter.HandleFunc("/products/add", adminHandler.AddProductPage).Methods("GET")
	adminRouter.HandleFunc("/products/add", adminHandler.AddProductPost).Methods("POST")
	adminRouter.HandleFunc("/products/edit/{id}", adminHandler.EditProductPage).Methods("GET")
	adminRouter.HandleFunc("/products/edit/{id}", adminHandler.EditProductPost).Methods("POST")
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

	checkoutRouter := router.PathPrefix("/checkout").Subrouter()

	checkoutRouter.HandleFunc("/process", komerceCheckoutHandler.DisplayCheckoutConfirmation).Methods("POST")
	checkoutRouter.HandleFunc("/initiate-midtrans", komerceCheckoutHandler.InitiateMidtransTransactionPost).Methods("POST")

	checkoutRouter.HandleFunc("/finish", komerceCheckoutHandler.CheckoutFinishGet).Methods("GET")
	checkoutRouter.HandleFunc("/unfinish", komerceCheckoutHandler.CheckoutUnfinishGet).Methods("GET")
	checkoutRouter.HandleFunc("/error", komerceCheckoutHandler.CheckoutErrorGet).Methods("GET")

	router.HandleFunc("/midtrans-notification", komerceCheckoutHandler.MidtransNotificationPost).Methods("POST")

	return router
}
