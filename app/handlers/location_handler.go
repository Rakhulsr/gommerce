package handlers

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"strconv"

// 	"github.com/Rakhulsr/go-ecommerce/app/services"
// 	"github.com/unrolled/render"
// )

// type KomerceLocationAPIHandler struct {
// 	komerceLocationSvc services.KomerceRajaOngkirClient
// 	render             *render.Render
// }

// func NewKomerceLocationAPIHandler(
// 	komerceLocationSvc services.KomerceRajaOngkirClient,
// 	render *render.Render,
// ) *KomerceLocationAPIHandler {
// 	return &KomerceLocationAPIHandler{
// 		komerceLocationSvc: komerceLocationSvc,
// 		render:             render,
// 	}
// }

// func (h *KomerceLocationAPIHandler) SearchDomesticDestinationAPI(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	keyword := r.URL.Query().Get("keyword")

// 	if keyword == "" {
// 		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
// 			"status":  "error",
// 			"message": "Keyword pencarian tidak boleh kosong.",
// 			"data":    []interface{}{},
// 		})
// 		return
// 	}

// 	destinations, err := h.komerceLocationSvc.SearchDomesticDestination(ctx, keyword)
// 	if err != nil {
// 		log.Printf("SearchDomesticDestinationAPI: Gagal mencari destinasi dari Komerce API: %v", err)
// 		h.render.JSON(w, http.StatusInternalServerError, map[string]interface{}{
// 			"status":  "error",
// 			"message": fmt.Sprintf("Gagal mencari destinasi: %v", err),
// 			"data":    []interface{}{},
// 		})
// 		return
// 	}

// 	h.render.JSON(w, http.StatusOK, map[string]interface{}{
// 		"status":  "success",
// 		"message": "Destinasi berhasil ditemukan.",
// 		"data":    destinations,
// 	})
// }

// func (h *KomerceLocationAPIHandler) GetDomesticDestinationByIDAPI(w http.ResponseWriter, r *http.Request) {
// 	// Fungsi ini tidak lagi memanggil GetDomesticDestinationByID dari service.
// 	// Endpoint ini mungkin tidak diperlukan lagi jika detail lokasi tidak diambil berdasarkan ID.
// 	// Namun, jika ada bagian lain dari aplikasi yang masih memanggilnya,
// 	// kita bisa mengembalikan respons error atau data kosong.
// 	// Untuk saat ini, saya akan mengembalikan 400 Bad Request karena Location ID tidak dapat diverifikasi secara langsung.
// 	log.Println("GetDomesticDestinationByIDAPI: Endpoint ini tidak lagi didukung untuk pencarian detail berdasarkan ID.")
// 	h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
// 		"status":  "error",
// 		"message": "Pencarian detail destinasi berdasarkan ID tidak lagi didukung secara langsung.",
// 		"data":    nil,
// 	})
// 	return
// }

// func (h *KomerceLocationAPIHandler) CalculateShippingCostKomerce(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	if err := r.ParseForm(); err != nil {
// 		log.Printf("CalculateShippingCostKomerce: Error parsing form: %v", err)
// 		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
// 			"status":  "error",
// 			"message": "Gagal membaca data form.",
// 		})
// 		return
// 	}

// 	originID := r.FormValue("origin_id")
// 	destinationID := r.FormValue("destination_id")
// 	weightStr := r.FormValue("weight")
// 	courier := r.FormValue("courier")

// 	if originID == "" || destinationID == "" || weightStr == "" || courier == "" {
// 		log.Printf("CalculateShippingCostKomerce: Data tidak lengkap. OriginID: %s, DestinationID: %s, Weight: %s, Courier: %s", originID, destinationID, weightStr, courier)
// 		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
// 			"status":  "error",
// 			"message": "Data pengiriman tidak lengkap. Mohon isi semua field yang wajib.",
// 		})
// 		return
// 	}

// 	weight, err := strconv.Atoi(weightStr)
// 	if err != nil || weight <= 0 {
// 		log.Printf("CalculateShippingCostKomerce: Berat tidak valid: %s, error: %v", weightStr, err)
// 		h.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
// 			"status":  "error",
// 			"message": "Berat tidak valid, harus berupa angka positif.",
// 		})
// 		return
// 	}

// 	log.Printf("CalculateShippingCostKomerce: Menghitung biaya pengiriman dari %s ke %s untuk berat %d dengan kurir %s", originID, destinationID, weight, courier)

// 	costs, err := h.komerceLocationSvc.GetDomesticShippingCost(ctx, originID, destinationID, weight, courier)
// 	if err != nil {
// 		log.Printf("CalculateShippingCostKomerce: Gagal menghitung biaya pengiriman dari Komerce API: %v", err)
// 		h.render.JSON(w, http.StatusInternalServerError, map[string]interface{}{
// 			"status":  "error",
// 			"message": fmt.Sprintf("Gagal menghitung biaya pengiriman: %v", err),
// 		})
// 		return
// 	}

// 	if len(costs) == 0 {
// 		log.Printf("CalculateShippingCostKomerce: Tidak ada biaya pengiriman ditemukan untuk rute ini.")
// 		h.render.JSON(w, http.StatusOK, map[string]interface{}{
// 			"status":  "success",
// 			"message": "Tidak ada biaya pengiriman ditemukan untuk rute ini.",
// 			"data":    []interface{}{},
// 		})
// 		return
// 	}

// 	log.Printf("CalculateShippingCostKomerce: Berhasil menghitung %d biaya pengiriman.", len(costs))
// 	h.render.JSON(w, http.StatusOK, map[string]interface{}{
// 		"status":  "success",
// 		"message": "Biaya pengiriman berhasil ditemukan.",
// 		"data":    costs,
// 	})
// }

// func (h *KomerceLocationAPIHandler) GetCitiesAPI(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	provinceName := r.URL.Query().Get("province_name")
// 	cities, err := h.komerceLocationSvc.GetCitiesByProvince(ctx, provinceName)
// 	if err != nil {
// 		log.Printf("GetCitiesAPI: Gagal mengambil kota untuk provinsi '%s': %v", provinceName, err)
// 		h.render.JSON(w, http.StatusInternalServerError, map[string]interface{}{
// 			"status":  "error",
// 			"message": fmt.Sprintf("Gagal mengambil daftar kota: %v", err),
// 			"data":    []interface{}{},
// 		})
// 		return
// 	}
// 	h.render.JSON(w, http.StatusOK, map[string]interface{}{
// 		"status":  "success",
// 		"message": "Daftar kota berhasil diambil.",
// 		"data":    cities,
// 	})
// }

// func (h *KomerceLocationAPIHandler) GetDistrictsAPI(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	cityName := r.URL.Query().Get("city_name")
// 	provinceName := r.URL.Query().Get("province_name")
// 	districts, err := h.komerceLocationSvc.GetDistrictsByCity(ctx, cityName, provinceName)
// 	if err != nil {
// 		log.Printf("GetDistrictsAPI: Gagal mengambil kecamatan untuk kota '%s', provinsi '%s': %v", cityName, provinceName, err)
// 		h.render.JSON(w, http.StatusInternalServerError, map[string]interface{}{
// 			"status":  "error",
// 			"message": fmt.Sprintf("Gagal mengambil daftar kecamatan: %v", err),
// 			"data":    []interface{}{},
// 		})
// 		return
// 	}
// 	h.render.JSON(w, http.StatusOK, map[string]interface{}{
// 		"status":  "success",
// 		"message": "Daftar kecamatan berhasil diambil.",
// 		"data":    districts,
// 	})
// }

// func (h *KomerceLocationAPIHandler) GetSubdistrictsAPI(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	districtName := r.URL.Query().Get("district_name")
// 	cityName := r.URL.Query().Get("city_name")
// 	provinceName := r.URL.Query().Get("province_name")
// 	subdistricts, err := h.komerceLocationSvc.GetSubdistrictsByDistrict(ctx, districtName, cityName, provinceName)
// 	if err != nil {
// 		log.Printf("GetSubdistrictsAPI: Gagal mengambil kelurahan untuk kecamatan '%s', kota '%s', provinsi '%s': %v", districtName, cityName, provinceName, err)
// 		h.render.JSON(w, http.StatusInternalServerError, map[string]interface{}{
// 			"status":  "error",
// 			"message": fmt.Sprintf("Gagal mengambil daftar kelurahan: %v", err),
// 			"data":    []interface{}{},
// 		})
// 		return
// 	}
// 	h.render.JSON(w, http.StatusOK, map[string]interface{}{
// 		"status":  "success",
// 		"message": "Daftar kelurahan berhasil diambil.",
// 		"data":    subdistricts,
// 	})
// }
