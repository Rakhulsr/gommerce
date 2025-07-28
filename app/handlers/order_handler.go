package handlers

import (
	"log"
	"net/http"
	"net/url"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models/other"
	"github.com/Rakhulsr/go-ecommerce/app/repositories"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

type OrderHandler struct {
	render      *render.Render
	orderRepo   repositories.OrderRepository
	userRepo    repositories.UserRepositoryImpl
	paymentRepo *repositories.PaymentRepositoryImpl
}

func NewOrderHandler(render *render.Render, orderRepo repositories.OrderRepository, userRepo repositories.UserRepositoryImpl, paymentRepo repositories.PaymentRepositoryImpl) *OrderHandler {
	return &OrderHandler{
		render:      render,
		orderRepo:   orderRepo,
		userRepo:    userRepo,
		paymentRepo: &paymentRepo,
	}
}

func (h *OrderHandler) OrderListGet(w http.ResponseWriter, r *http.Request) {

	userID := helpers.GetUserIDFromContext(r.Context())

	if userID == "" {
		http.Redirect(w, r, "/login?status=error&message="+url.QueryEscape("Anda harus login untuk melihat pesanan."), http.StatusSeeOther)
		return
	}

	ctx := r.Context()

	orders, err := h.orderRepo.FindByUserID(ctx, userID)
	if err != nil {
		log.Printf("OrderListGet: Gagal mendapatkan daftar pesanan untuk UserID %s: %v", userID, err)
		http.Redirect(w, r, "/dashboard?status=error&message="+url.QueryEscape("Gagal memuat daftar pesanan."), http.StatusSeeOther)
		return
	}

	pageData := other.BasePageData{}
	baseDataMap := helpers.GetBaseData(r, nil)
	helpers.PopulateBaseData(&pageData, baseDataMap)

	pageData.Title = "Daftar Pesanan Saya"
	pageData.Orders = orders

	h.render.HTML(w, http.StatusOK, "order_list", pageData)
}

func (h *OrderHandler) OrderDetailGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderCode := vars["orderCode"]

	if orderCode == "" {
		http.Redirect(w, r, "/orders?status=error&message="+url.QueryEscape("Kode pesanan tidak ditemukan."), http.StatusSeeOther)
		return
	}

	ctx := r.Context()

	order, err := h.orderRepo.FindByCodeWithDetails(ctx, orderCode)
	if err != nil {
		log.Printf("OrderDetailGet: Gagal mendapatkan detail pesanan %s: %v", orderCode, err)
		http.Redirect(w, r, "/orders?status=error&message="+url.QueryEscape("Pesanan tidak ditemukan atau terjadi kesalahan."), http.StatusSeeOther)
		return
	}

	if order == nil {
		log.Printf("OrderDetailGet: Pesanan %s tidak ditemukan.", orderCode)
		http.Redirect(w, r, "/orders?status=error&message="+url.QueryEscape("Pesanan tidak ditemukan."), http.StatusSeeOther)
		return
	}

	userID := helpers.GetUserIDFromContext(r.Context())
	if order.UserID != userID {
		http.Redirect(w, r, "/orders?status=error&message="+url.QueryEscape("Anda tidak memiliki akses ke pesanan ini."), http.StatusForbidden)
		return
	}

	payment, err := h.paymentRepo.FindByOrderID(ctx, order.ID)
	if err != nil {
		log.Printf("OrderDetailGet: Gagal mendapatkan detail pembayaran untuk OrderID %s: %v", order.ID, err)

	}

	if order.OrderItems == nil || len(order.OrderItems) == 0 {
		log.Printf("DEBUG: OrderDetailGet: OrderItems kosong atau nil untuk pesanan %s.", order.OrderCode)
	} else {
		for i, item := range order.OrderItems {
			log.Printf("DEBUG: OrderDetailGet: OrderItem %d: ProductName=%s, Qty=%d, ProductID=%s", i, item.ProductName, item.Qty, item.ProductID)
			if item.Product.ProductImages == nil || len(item.Product.ProductImages) == 0 {
				log.Printf("DEBUG: OrderDetailGet: ProductImages kosong atau nil untuk produk %s (OrderItem %d).", item.ProductName, i)
			} else {
				for j, img := range item.Product.ProductImages {
					log.Printf("DEBUG: OrderDetailGet: ProductImage %d untuk %s: Path=%s, ID=%s", j, item.ProductName, img.Path, img.ID)
				}
			}
		}
	}

	pageData := other.BasePageData{}
	baseDataMap := helpers.GetBaseData(r, nil)
	helpers.PopulateBaseData(&pageData, baseDataMap)

	pageData.Title = "Detail Pesanan #" + order.OrderCode
	pageData.Order = order
	pageData.Payment = payment

	h.render.HTML(w, http.StatusOK, "order_detail", pageData)
}
