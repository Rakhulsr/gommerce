package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Rakhulsr/go-ecommerce/app/services"
	"github.com/unrolled/render"
)

type LocationAPIHandler struct {
	rajaOngkirSvc services.RajaOngkirClient
	render        *render.Render
}

func NewLocationAPIHandler(rajaOngkirSvc services.RajaOngkirClient, render *render.Render) *LocationAPIHandler {
	return &LocationAPIHandler{
		rajaOngkirSvc: rajaOngkirSvc,
		render:        render,
	}
}

func (h *LocationAPIHandler) GetProvincesAPI(w http.ResponseWriter, r *http.Request) {
	provinces, err := h.rajaOngkirSvc.GetProvincesFromAPI()
	if err != nil {
		log.Printf("LocationAPIHandler: Error getting provinces from service: %v", err)
		http.Error(w, "Failed to fetch provinces: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"provinces": provinces})
}

func (h *LocationAPIHandler) GetCitiesAPI(w http.ResponseWriter, r *http.Request) {
	provinceID := r.URL.Query().Get("province_id")

	cities, err := h.rajaOngkirSvc.GetCitiesFromAPI(provinceID)
	if err != nil {
		log.Printf("LocationAPIHandler: Error getting cities for province %s from service: %v", provinceID, err)
		http.Error(w, "Failed to fetch cities: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"cities": cities})
}

func (h *LocationAPIHandler) CalculateShippingCostAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Origin      string `json:"origin"`
		Destination string `json:"destination"`
		Weight      int    `json:"weight"`
		Courier     string `json:"courier"`
	}

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Printf("LocationAPIHandler: Error decoding calculate shipping cost request body: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if reqBody.Origin == "" || reqBody.Destination == "" || reqBody.Weight <= 0 || reqBody.Courier == "" {
		http.Error(w, "Missing or invalid parameters (origin, destination, weight, courier are required and weight must be positive)", http.StatusBadRequest)
		return
	}

	costs, err := h.rajaOngkirSvc.CalculateCost(reqBody.Origin, reqBody.Destination, reqBody.Weight, reqBody.Courier)
	if err != nil {
		log.Printf("LocationAPIHandler: Error calculating shipping cost via RajaOngkirService: %v", err)
		http.Error(w, "Failed to calculate shipping cost: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"costs": costs})
}
