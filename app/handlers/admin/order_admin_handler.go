package admin

import (
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"
	"github.com/Rakhulsr/go-ecommerce/app/models"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
)

func (h *AdminHandler) GetOrdersPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orders, err := h.orderRepo.GetAllOrders(ctx)
	if err != nil {
		log.Printf("AdminHandler.GetOrdersPage: Gagal mendapatkan semua pesanan: %v", err)
		http.Redirect(w, r, "/admin/dashboard?status=error&message="+url.QueryEscape("Gagal memuat daftar pesanan."), http.StatusSeeOther)
		return
	}

	pageData := helpers.AdminOrderPageData{}
	baseDataMap := helpers.GetBaseData(r, nil)
	helpers.PopulateBaseData(&pageData.BasePageData, baseDataMap)

	pageData.Title = "Daftar Pesanan Admin"
	pageData.Orders = orders

	pageData.Title = "Manajemen Kategori"
	pageData.IsAuthPage = true
	pageData.IsAdminPage = true
	pageData.HideAdminWelcomeMessage = true

	pageData.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"},
		{Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Orders", URL: "/admin/orders"},
	}

	pageData.OrderStatusOptions = map[int]string{
		models.OrderStatusPending:    "Menunggu Pembayaran",
		models.OrderStatusProcessing: "Sedang Diproses",
		models.OrderStatusShipped:    "Dalam Pengiriman",
		models.OrderStatusCompleted:  "Selesai",
		models.OrderStatusCancelled:  "Dibatalkan",
		models.OrderStatusRefunded:   "Pengembalian Dana",
		models.OrderStatusFailed:     "Gagal",
	}

	h.render.HTML(w, http.StatusOK, "admin/orders/list", pageData)
}

func (h *AdminHandler) UpdateOrderStatusPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orderID := r.FormValue("order_id")
	newStatusStr := r.FormValue("new_status")

	if orderID == "" || newStatusStr == "" {
		http.Redirect(w, r, "/admin/orders?status=error&message="+url.QueryEscape("ID pesanan atau status baru tidak valid."), http.StatusSeeOther)
		return
	}

	newStatus, err := strconv.Atoi(newStatusStr)
	if err != nil {
		http.Redirect(w, r, "/admin/orders?status=error&message="+url.QueryEscape("Status baru tidak valid."), http.StatusSeeOther)
		return
	}

	err = h.orderRepo.UpdateStatus(ctx, orderID, newStatus)
	if err != nil {
		log.Printf("AdminHandler.UpdateOrderStatusPost: Gagal memperbarui status pesanan %s ke %d: %v", orderID, newStatus, err)
		http.Redirect(w, r, "/admin/orders?status=error&message="+url.QueryEscape("Gagal memperbarui status pesanan."), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/orders?status=success&message="+url.QueryEscape("Status pesanan berhasil diperbarui."), http.StatusSeeOther)
}
