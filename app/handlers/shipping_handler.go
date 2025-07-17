package handlers

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"

// 	"github.com/Rakhulsr/go-ecommerce/app/services"
// 	"github.com/unrolled/render"
// )

// type LocationAPIHandler struct {
// 	rajaOngkirSvc services.RajaOngkirClient
// 	render        *render.Render
// }

// func NewLocationAPIHandler(rajaOngkirSvc services.RajaOngkirClient, render *render.Render) *LocationAPIHandler {
// 	return &LocationAPIHandler{
// 		rajaOngkirSvc: rajaOngkirSvc,
// 		render:        render,
// 	}
// }

// func (h *LocationAPIHandler) GetProvincesAPI(w http.ResponseWriter, r *http.Request) {
// 	provinces, err := h.rajaOngkirSvc.GetProvincesFromAPI()
// 	if err != nil {
// 		log.Printf("LocationAPIHandler: Error getting provinces from service: %v", err)
// 		http.Error(w, "Failed to fetch provinces: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{"provinces": provinces})
// }

// func (h *LocationAPIHandler) GetCitiesAPI(w http.ResponseWriter, r *http.Request) {
// 	provinceID := r.URL.Query().Get("province_id")

// 	cities, err := h.rajaOngkirSvc.GetCitiesFromAPI(provinceID)
// 	if err != nil {
// 		log.Printf("LocationAPIHandler: Error getting cities for province %s from service: %v", provinceID, err)
// 		http.Error(w, "Failed to fetch cities: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{"cities": cities})
// }

// func (h *LocationAPIHandler) CalculateShippingCost(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var req struct {
// 		Origin      string `json:"origin"`
// 		Destination string `json:"destination"`
// 		Weight      int    `json:"weight"`
// 		Courier     string `json:"courier"`
// 	}

// 	err := json.NewDecoder(r.Body).Decode(&req)
// 	if err != nil {
// 		log.Printf("CalculateShippingCost: Gagal decode request body: %v", err)
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	if req.Origin == "" || req.Destination == "" || req.Weight <= 0 || req.Courier == "" {
// 		log.Printf("CalculateShippingCost: Data input tidak lengkap atau tidak valid. Origin: %s, Destination: %s, Weight: %d, Courier: %s", req.Origin, req.Destination, req.Weight, req.Courier)
// 		http.Error(w, "Origin, destination, weight, and courier are required and must be valid.", http.StatusBadRequest)
// 		return
// 	}

// 	shippingCosts, err := h.rajaOngkirSvc.CalculateCost(req.Origin, req.Destination, req.Weight, req.Courier)
// 	if err != nil {
// 		log.Printf("CalculateShippingCost: Gagal mendapatkan biaya pengiriman dari RajaOngkir API: %v", err)

// 		http.Error(w, fmt.Sprintf("Gagal menghitung ongkir: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)

// 	json.NewEncoder(w).Encode(shippingCosts)
// 	log.Printf("CalculateShippingCost: Berhasil mengirimkan biaya pengiriman untuk Origin=%s, Destination=%s, Courier=%s", req.Origin, req.Destination, req.Courier)
// }
