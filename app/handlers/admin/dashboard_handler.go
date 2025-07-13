package admin

import (
	"net/http"
	"net/url"

	"github.com/Rakhulsr/go-ecommerce/app/helpers"

	"github.com/Rakhulsr/go-ecommerce/app/models/other"
	"github.com/Rakhulsr/go-ecommerce/app/utils/breadcrumb"
	"github.com/unrolled/render"
)

type AdminDashboardHandler struct {
	render *render.Render
}

func NewAdminDashboardHandler(render *render.Render) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		render: render,
	}
}

type AdminPageData struct {
	other.BasePageData
	// Tambahkan bidang lain khusus untuk AdminPageData di sini jika ada
}

func (h *AdminDashboardHandler) populateBaseData(r *http.Request, pageData *AdminPageData) {
	baseDataMap := helpers.GetBaseData(r, nil)

	if title, ok := baseDataMap["Title"].(string); ok {
		pageData.Title = title
	}
	if isLoggedIn, ok := baseDataMap["IsLoggedIn"].(bool); ok {
		pageData.IsLoggedIn = isLoggedIn
	}
	// === PERBAIKAN DI SINI: Cast ke *other.UserForTemplate ===
	if user, ok := baseDataMap["User"].(*other.UserForTemplate); ok { // UBAH TIPE INI
		pageData.User = user
	}
	// === AKHIR PERBAIKAN ===
	if userID, ok := baseDataMap["UserID"].(string); ok {
		pageData.UserID = userID
	}
	if cartCount, ok := baseDataMap["CartCount"].(int); ok {
		pageData.CartCount = cartCount
	}
	if csrfToken, ok := baseDataMap["CSRFToken"].(string); ok {
		pageData.CSRFToken = csrfToken
	}
	if message, ok := baseDataMap["Message"].(string); ok {
		pageData.Message = message
	}
	if messageStatus, ok := baseDataMap["MessageStatus"].(string); ok {
		pageData.MessageStatus = messageStatus
	}
	if query, ok := baseDataMap["Query"].(url.Values); ok {
		pageData.Query = query
	}
	if breadcrumbs, ok := baseDataMap["Breadcrumbs"].([]breadcrumb.Breadcrumb); ok {
		pageData.Breadcrumbs = breadcrumbs
	}
	if isAuthPage, ok := baseDataMap["IsAuthPage"].(bool); ok {
		pageData.IsAuthPage = isAuthPage
	}
	// Ambil IsAdminPage dari baseDataMap yang sudah diisi helpers.GetBaseData
	if isAdminPage, ok := baseDataMap["IsAdminPage"].(bool); ok {
		pageData.IsAdminPage = isAdminPage
	}
}

func (h *AdminDashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	data := &AdminPageData{}
	h.populateBaseData(r, data)

	data.Title = "Dashboard Admin"
	data.IsAuthPage = true // Halaman admin biasanya memerlukan otentikasi

	// IsAdminPage sudah diisi oleh populateBaseData berdasarkan role user
	// if data.User != nil && data.User.Role == "admin" {
	// 	data.IsAdminPage = true
	// }

	data.Breadcrumbs = []breadcrumb.Breadcrumb{
		{Name: "Beranda", URL: "/"},
		{Name: "Admin", URL: "/admin/dashboard"},
		{Name: "Dashboard", URL: "/admin/dashboard"},
	}

	h.render.HTML(w, http.StatusOK, "admin/dashboard_index", data)
}
