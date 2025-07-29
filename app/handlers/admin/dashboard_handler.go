package admin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/models/other"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/services"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/Rakhulsr/go-ecommerce/app/utils/calc"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/unrolled/render"
)

type AdminHandler struct {
	render       *render.Render
	validator    *validator.Validate
	productRepo  repositories.ProductRepositoryImpl
	categoryRepo repositories.CategoryRepositoryImpl
	sectionRepo  repositories.SectionRepositoryImpl
	userRepo     repositories.UserRepositoryImpl
	cartRepo     repositories.CartRepositoryImpl
	cartItemRepo repositories.CartItemRepositoryImpl
	cartSvc      services.CartService
	orderRepo    repositories.OrderRepository
}

func NewAdminHandler(
	render *render.Render,
	validator *validator.Validate,
	productRepo repositories.ProductRepositoryImpl,
	categoryRepo repositories.CategoryRepositoryImpl,
	sectionRepo repositories.SectionRepositoryImpl,
	userRepo repositories.UserRepositoryImpl,
	cartRepo repositories.CartRepositoryImpl,
	cartItemRepo repositories.CartItemRepositoryImpl,
	cartSvc services.CartService,
	orderRepo repositories.OrderRepository,
) *AdminHandler {
	return &AdminHandler{
		render:       render,
		validator:    validator,
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		sectionRepo:  sectionRepo,
		userRepo:     userRepo,
		cartRepo:     cartRepo,
		cartItemRepo: cartItemRepo,
		cartSvc:      cartSvc,
		orderRepo:    orderRepo,
	}
}

type AdminPageData struct {
	other.BasePageData
	TotalUsers       int64
	TotalProducts    int64
	TotalOrders      int64
	RecentOrders     []models.Order
	RecentActivities []struct {
		Activity string
		Time     time.Time
	}
	CurrentGlobalDiscount float64
}

type AdminProductPageData struct {
	other.BasePageData
	Products    []models.Product
	ProductData *ProductForm
	IsEdit      bool
	FormAction  string
	Errors      map[string]string
	Categories  []models.Category
}

type ProductForm struct {
	ID              string
	Name            string `form:"name" validate:"required,min=3,max=100"`
	Description     string `form:"description" validate:"required,min=10"`
	SKU             string `form:"sku" validate:"required,alphanum,min=3,max=20"`
	Price           string `form:"price" validate:"required,numeric,min=0"`
	Stock           string `form:"stock" validate:"required,numeric,min=0"`
	Weight          string `form:"weight" validate:"required,numeric,min=0"`
	CategoryID      string `form:"category_id" validate:"required"`
	DiscountPercent string `form:"discount_percent" validate:"omitempty,numeric,min=0,max=100"`

	ExistingImages []models.ProductImage
}

type AdminCategoryPageData struct {
	other.BasePageData
	Categories   []models.Category
	CategoryData *CategoryForm
	IsEdit       bool
	FormAction   string
	Errors       map[string]string
	Sections     []models.Section
}

type CategoryForm struct {
	ID   string `form:"id"`
	Name string `form:"name" validate:"required,min=3,max=100"`
	Slug string

	SectionID string `form:"section_id"`
}

type AdminUserPageData struct {
	other.BasePageData
	Users      []models.User
	UserData   *UserForm
	IsEdit     bool
	FormAction string
	Errors     map[string]string
}

type UserForm struct {
	ID        string `form:"id"`
	FirstName string `form:"first_name" validate:"required,min=2,max=50"`
	LastName  string `form:"last_name" validate:"min=2,max=50"`
	Email     string `form:"email" validate:"required,email"`
	Password  string `form:"password"`
	Role      string `form:"role" validate:"required,oneof=admin customer"`
}

func (h *AdminHandler) populateBaseDataForAdmin(r *http.Request, pageData interface{}) {
	baseDataMap := helpers.GetBaseData(r, nil)

	var base *other.BasePageData
	switch pd := pageData.(type) {
	case *AdminPageData:
		base = &pd.BasePageData
	case *AdminProductPageData:
		base = &pd.BasePageData
	case *AdminCategoryPageData:
		base = &pd.BasePageData
	case *AdminUserPageData:
		base = &pd.BasePageData
	default:
		log.Printf("populateBaseDataForAdmin: Unknown pageData type: %T", pageData)
		return
	}

	if title, ok := baseDataMap["Title"].(string); ok {
		base.Title = title
	}
	if isLoggedIn, ok := baseDataMap["IsLoggedIn"].(bool); ok {
		base.IsLoggedIn = isLoggedIn
	}
	if user, ok := baseDataMap["User"].(*other.UserForTemplate); ok {
		base.User = user
	}
	if userID, ok := baseDataMap["UserID"].(string); ok {
		base.UserID = userID
	}
	if cartCount, ok := baseDataMap["CartCount"].(int); ok {
		base.CartCount = cartCount
	}
	if csrfToken, ok := baseDataMap["CSRFToken"].(string); ok {
		base.CSRFToken = csrfToken
	}
	if message, ok := baseDataMap["Message"].(string); ok {
		base.Message = message
	}
	if messageStatus, ok := baseDataMap["MessageStatus"].(string); ok {
		base.MessageStatus = messageStatus
	}
	if query, ok := baseDataMap["Query"].(url.Values); ok {
		base.Query = query
	}
	if breadcrumbs, ok := baseDataMap["Breadcrumbs"].([]breadcrumb.Breadcrumb); ok {
		base.Breadcrumbs = breadcrumbs
	}
	if isAuthPage, ok := baseDataMap["IsAuthPage"].(bool); ok {
		base.IsAuthPage = isAuthPage
	}
	if isAdminPage, ok := baseDataMap["IsAdminPage"].(bool); ok {
		base.IsAdminPage = isAdminPage
	}
	if hideAdminWelcomeMessage, ok := baseDataMap["HideAdminWelcomeMessage"].(bool); ok {
		base.HideAdminWelcomeMessage = hideAdminWelcomeMessage
	}
	base.CurrentPath = r.URL.Path
	if strings.HasPrefix(r.URL.Path, "/admin/") {
		base.IsAdminRoute = true
	} else {
		base.IsAdminRoute = false
	}
}
func (h *AdminHandler) applyGlobalDiscount(ctx context.Context, discountPercent float64) error {
	discountDecimal := decimal.NewFromFloat(discountPercent)

	products, err := h.productRepo.GetProducts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all products for global discount: %w", err)
	}

	for _, product := range products {

		productDiscountAmount := calc.CalculateDiscount(product.Price, discountDecimal)

		if err := h.productRepo.UpdateProductDiscount(ctx, product.ID, discountDecimal, productDiscountAmount); err != nil {
			log.Printf("applyGlobalDiscount: Failed to update discount for product %s: %v", product.ID, err)

		}
	}

	carts, err := h.cartRepo.GetAllCarts(ctx)
	if err != nil {
		log.Printf("applyGlobalDiscount: Failed to get all carts to update summaries after product discount: %v", err)
		return fmt.Errorf("failed to update cart summaries after global product discount: %w", err)
	}

	for _, cart := range carts {

		if err := h.cartRepo.UpdateCartSummary(ctx, cart.ID); err != nil {
			log.Printf("applyGlobalDiscount: Failed to update cart summary for cart %s after product discount: %v", cart.ID, err)

		}
	}

	return nil
}

func (h *AdminHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := &AdminPageData{}
	h.populateBaseDataForAdmin(r, data)

	data.Title = "Dashboard Admin"
	data.IsAuthPage = true

	data.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"},
		{Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Dashboard", URL: "/admin/dashboard"},
	}

	totalUsers, err := h.userRepo.GetUserCount(ctx)
	if err != nil {
		log.Printf("GetDashboard: Gagal mengambil total pengguna: %v", err)
		data.TotalUsers = 0
	} else {
		data.TotalUsers = totalUsers
	}

	totalProducts, err := h.productRepo.GetProductCount(ctx)
	if err != nil {
		log.Printf("GetDashboard: Gagal mengambil total produk: %v", err)
		data.TotalProducts = 0
	} else {
		data.TotalProducts = totalProducts
	}

	totalOrders, err := h.orderRepo.GetOrderCount(ctx)
	if err != nil {
		log.Printf("GetDashboard: Gagal mengambil total pesanan: %v", err)
		data.TotalOrders = 0
	} else {
		data.TotalOrders = totalOrders
	}

	const limitOrders = 5
	recentOrders, err := h.orderRepo.GetTopNOrders(ctx, limitOrders)
	if err != nil {
		log.Printf("GetDashboard: Gagal mengambil %d pesanan terbaru: %v", limitOrders, err)
		data.RecentOrders = []models.Order{}
	} else {
		data.RecentOrders = recentOrders
		log.Printf("GetDashboard: Berhasil mengambil %d pesanan terbaru.", len(recentOrders))
	}

	products, err := h.productRepo.GetProducts(ctx)
	if err == nil && len(products) > 0 {
		data.CurrentGlobalDiscount = products[0].DiscountPercent.InexactFloat64()
	} else {
		data.CurrentGlobalDiscount = 0.0
	}

	data.RecentActivities = []struct {
		Activity string
		Time     time.Time
	}{
		{Activity: "Produk baru 'Smartwatch X' ditambahkan.", Time: time.Now().Add(-30 * time.Minute)},
		{Activity: "Pengguna 'budi.s@example.com' mendaftar.", Time: time.Now().Add(-2 * time.Hour)},
		{Activity: "Pesanan ORD002 diselesaikan.", Time: time.Now().Add(-7 * time.Hour)},
	}

	h.render.HTML(w, http.StatusOK, "admin/dashboard/index", data)
}

func (h *AdminHandler) ApplyGlobalDiscountPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ApplyGlobalDiscountPost: Kesalahan parsing form: %v", err)
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	discountPercentStr := r.PostFormValue("global_discount_percent")
	discountPercent, err := strconv.ParseFloat(discountPercentStr, 64)
	if err != nil || discountPercent < 0 || discountPercent > 100 {
		log.Printf("ApplyGlobalDiscountPost: Persentase diskon tidak valid: %v", err)
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	err = h.applyGlobalDiscount(r.Context(), discountPercent)
	if err != nil {
		log.Printf("ApplyGlobalDiscountPost: Gagal menerapkan diskon global: %v", err)
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}
