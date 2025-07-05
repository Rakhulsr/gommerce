package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/Rakhulsr/go-ecommerce/app/utils/calc"
	"github.com/Rakhulsr/go-ecommerce/app/utils/sessions"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/unrolled/render"
	"gorm.io/gorm"
)

type CartHandler struct {
	productRepo  repositories.ProductRepository
	cartRepo     repositories.CartRepository
	cartItemRepo repositories.CartItemRepository
	render       *render.Render
}

func NewCartHandler(productRepo repositories.ProductRepository, cartRepo repositories.CartRepository, render render.Render, cartItemRepo repositories.CartItemRepository) *CartHandler {
	return &CartHandler{productRepo, cartRepo, cartItemRepo, &render}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	cartID, err := sessions.GetCartID(w, r)
	if err != nil {
		http.Error(w, "Gagal mengakses cart", http.StatusInternalServerError)
		return
	}

	cart, err := h.cartRepo.GetCartWithItems(r.Context(), cartID)
	if err != nil {
		http.Error(w, "Gagal mengambil data cart", http.StatusInternalServerError)
		return
	}

	_ = h.render.HTML(w, http.StatusOK, "cart", map[string]interface{}{
		"title": "Keranjang Belanja",
		"cart":  cart,
	})
}

func (h *CartHandler) AddItemCart(w http.ResponseWriter, r *http.Request) {

	productID := r.FormValue("product_id")
	qtyStr := r.FormValue("qty")

	qty, err := strconv.Atoi(qtyStr)
	if err != nil || qty <= 0 {
		http.Error(w, "Jumlah tidak valid", http.StatusBadRequest)
		return
	}

	product, err := h.productRepo.GetByID(r.Context(), productID)
	if err != nil {
		http.Redirect(w, r, "/products/"+product.Slug, http.StatusSeeOther)
		return
	}

	if qty > product.Stock {
		http.Redirect(w, r, "/products/"+product.Slug, http.StatusSeeOther)
		return
	}

	cartID, err := sessions.GetCartID(w, r)
	if err != nil {
		http.Error(w, "Gagal mendapatkan cart session", http.StatusInternalServerError)
		return
	}

	cart, err := h.cartRepo.GetByID(r.Context(), cartID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Gagal mengambil cart: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if cart == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		cart = &models.Cart{
			ID:              cartID,
			BaseTotalPrice:  decimal.Decimal{},
			TaxAmount:       decimal.Decimal{},
			TaxPercent:      decimal.Decimal{},
			DiscountAmount:  decimal.Decimal{},
			DiscountPercent: decimal.Decimal{},
			GrandTotal:      decimal.Decimal{},
		}
		if err := h.cartRepo.CreateCart(r.Context(), cart); err != nil {
			log.Printf("Gagal membuat cart: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	basePrice := product.Price
	baseTotal := basePrice.Mul(decimal.NewFromInt(int64(qty)))
	taxPercent := calc.GetTaxPercent()
	taxAmount := calc.CalculateTax(baseTotal)
	discountPercent := product.DiscountPercent
	discountAmount := calc.CalculateDiscount(baseTotal, discountPercent)
	grandTotal := calc.CalculateGrandTotal(baseTotal, taxAmount, discountAmount)
	subTotal := grandTotal

	existingItem, err := h.cartItemRepo.GetCartAndProduct(r.Context(), cartID, productID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Gagal mengecek item existing: %v", err)
		http.Error(w, "Gagal memproses permintaan", http.StatusInternalServerError)
		return
	}

	if existingItem != nil {

		existingItem.Qty = qty
		existingItem.BaseTotal = existingItem.BasePrice.Mul(decimal.NewFromInt(int64(qty)))
		existingItem.TaxPercent = calc.GetTaxPercent()
		existingItem.TaxAmount = calc.CalculateTax(existingItem.BaseTotal)
		existingItem.DiscountAmount = calc.CalculateDiscount(existingItem.BaseTotal, discountPercent)
		existingItem.DiscountPercent = product.DiscountPercent
		existingItem.GrandTotal = calc.CalculateGrandTotal(existingItem.BaseTotal, existingItem.TaxAmount, existingItem.DiscountAmount)
		existingItem.SubTotal = existingItem.GrandTotal

		if err := h.cartItemRepo.Update(r.Context(), existingItem); err != nil {
			log.Printf("Gagal update item di cart: %v", err)
			http.Error(w, "Gagal menambahkan item ke keranjang", http.StatusInternalServerError)
			return
		}
	} else {

		item := &models.CartItem{
			ID:              uuid.New().String(),
			CartID:          cartID,
			ProductID:       productID,
			Qty:             qty,
			BasePrice:       basePrice,
			BaseTotal:       baseTotal,
			TaxAmount:       taxAmount,
			TaxPercent:      taxPercent,
			DiscountAmount:  discountAmount,
			DiscountPercent: discountPercent,
			GrandTotal:      grandTotal,
			SubTotal:        subTotal,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		if err := h.cartItemRepo.Add(r.Context(), item); err != nil {
			log.Printf("Gagal menambahkan item baru: %v", err)
			http.Error(w, "Gagal menambahkan item", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/carts", http.StatusSeeOther)
}
