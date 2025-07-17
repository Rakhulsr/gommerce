package services

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// 	"net/url"
// 	"strconv"
// 	"time"

// 	"github.com/Rakhulsr/go-ecommerce/app/configs"
// 	"github.com/Rakhulsr/go-ecommerce/app/models/other"
// )

// var (
// 	rajaOngkirBaseURL = configs.LoadENV.API_ONGKIR_BASE_URL
// 	rajaOngkirAPIKey  = configs.LoadENV.API_ONGKIR_KEY
// )

// type RajaOngkirClient interface {
// 	CalculateCost(origin, destination string, weight int, courier string) (*other.CostResponse, error)
// 	GetProvincesFromAPI() ([]other.Province, error)
// 	GetCitiesFromAPI(provinceID string) ([]other.City, error)
// 	GetProvinceByID(provinceID string) (*other.Province, error) // NEW: Tambahkan metode ini
// 	GetCityByID(cityID string) (*other.City, error)
// }
// type RajaOngkirService struct {
// 	client  *http.Client
// 	apiKey  string
// 	baseURL string
// }

// func NewRajaOngkirService() *RajaOngkirService {

// 	return &RajaOngkirService{
// 		client: &http.Client{
// 			Timeout: 30 * time.Second,
// 		},
// 		apiKey:  rajaOngkirAPIKey,
// 		baseURL: rajaOngkirBaseURL,
// 	}
// }

// func (s *RajaOngkirService) GetProvincesFromAPI() ([]other.Province, error) {
// 	req, err := http.NewRequest("GET", s.baseURL+"/province", nil)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error creating request for provinces: %v", err)
// 		return nil, fmt.Errorf("failed to create request: %w", err)
// 	}

// 	req.Header.Add("key", s.apiKey)

// 	resp, err := s.client.Do(req)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error performing request to RajaOngkir province API: %v", err)
// 		return nil, fmt.Errorf("failed to perform request to RajaOngkir: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error reading response body from RajaOngkir province API: %v", err)
// 		return nil, fmt.Errorf("failed to read response body: %w", err)
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		log.Printf("RajaOngkirService: RajaOngkir province API returned non-OK status: %d, Body: %s", resp.StatusCode, string(body))
// 		return nil, fmt.Errorf("RajaOngkir API error: status %d, body: %s", resp.StatusCode, string(body))
// 	}

// 	var provinceResponse other.ProvinceResponse
// 	err = json.Unmarshal(body, &provinceResponse)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error unmarshalling province API response: %v, Raw Body: %s", err, string(body))
// 		return nil, fmt.Errorf("failed to parse province API response: %w", err)
// 	}

// 	// KOREKSI: Akses RajaOngkir.Results
// 	if provinceResponse.RajaOngkir.Status.Code != 200 {
// 		return nil, fmt.Errorf("RajaOngkir API returned error: %s (Code: %d)", provinceResponse.RajaOngkir.Status.Description, provinceResponse.RajaOngkir.Status.Code)
// 	}

// 	log.Printf("RajaOngkirService: Successfully fetched %d provinces from RajaOngkir API.", len(provinceResponse.RajaOngkir.Results))
// 	return provinceResponse.RajaOngkir.Results, nil
// }

// func (s *RajaOngkirService) GetCitiesFromAPI(provinceID string) ([]other.City, error) {
// 	url := rajaOngkirBaseURL + "/city"
// 	if provinceID != "" {
// 		url = fmt.Sprintf("%s?province=%s", url, provinceID)
// 	}

// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		log.Printf("LocationService: Error creating request for cities: %v", err)
// 		return nil, fmt.Errorf("failed to create request: %w", err)
// 	}

// 	req.Header.Add("key", rajaOngkirAPIKey)

// 	resp, err := s.client.Do(req)
// 	if err != nil {
// 		log.Printf("LocationService: Error performing request to RajaOngkir city API: %v", err)
// 		return nil, fmt.Errorf("failed to perform request to RajaOngkir: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Printf("LocationService: Error reading response body from RajaOngkir city API: %v", err)
// 		return nil, fmt.Errorf("failed to read response body: %w", err)
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		log.Printf("LocationService: RajaOngkir city API returned non-OK status: %d, Body: %s", resp.StatusCode, string(body))
// 		return nil, fmt.Errorf("RajaOngkir API error: status %d, body: %s", resp.StatusCode, string(body))
// 	}

// 	type CityResponse struct {
// 		RajaOngkir struct {
// 			Results []other.City `json:"results"`
// 		} `json:"rajaongkir"`
// 	}
// 	var cityResponse CityResponse
// 	err = json.Unmarshal(body, &cityResponse)
// 	if err != nil {
// 		log.Printf("LocationService: Error unmarshalling city API response: %v, Raw Body: %s", err, string(body))
// 		return nil, fmt.Errorf("failed to parse city API response: %w", err)
// 	}

// 	log.Printf("LocationService: Successfully fetched %d cities from RajaOngkir API.", len(cityResponse.RajaOngkir.Results))
// 	return cityResponse.RajaOngkir.Results, nil
// }

// func (s *RajaOngkirService) CalculateCost(origin, destination string, weight int, courier string) (*other.CostResponse, error) {
// 	if weight <= 0 {
// 		weight = 1
// 	}

// 	formData := url.Values{}
// 	formData.Set("origin", origin)
// 	formData.Set("destination", destination)
// 	formData.Set("weight", strconv.Itoa(weight))
// 	formData.Set("courier", courier)

// 	req, err := http.NewRequest("POST", s.baseURL+"/cost", bytes.NewBufferString(formData.Encode()))
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error creating request for cost calculation: %v", err)
// 		return nil, fmt.Errorf("failed to create cost request: %w", err)
// 	}

// 	req.Header.Add("key", s.apiKey)
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

// 	log.Printf("RajaOngkirService: Calling RajaOngkir API for cost. Params: Origin=%s, Destination=%s, Weight=%d, Courier=%s", origin, destination, weight, courier)

// 	resp, err := s.client.Do(req)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error performing request to RajaOngkir cost API: %v", err)
// 		return nil, fmt.Errorf("failed to perform cost request to RajaOngkir: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error reading response body from RajaOngkir cost API: %v", err)
// 		return nil, fmt.Errorf("failed to read cost response body: %w", err)
// 	}

// 	log.Printf("RajaOngkirService: Raw RajaOngkir cost API response - Status: %s, Body: %s", resp.Status, string(body))

// 	if resp.StatusCode != http.StatusOK {
// 		var errorResponse struct {
// 			RajaOngkir struct {
// 				Status struct {
// 					Code        int    `json:"code"`
// 					Description string `json:"description"`
// 				} `json:"status"`
// 			} `json:"rajaongkir"`
// 		}
// 		json.Unmarshal(body, &errorResponse)
// 		errMsg := fmt.Sprintf("RajaOngkir cost API returned non-OK status: %d. Description: %s. Body: %s",
// 			resp.StatusCode, errorResponse.RajaOngkir.Status.Description, string(body))
// 		log.Printf("RajaOngkirService: %s", errMsg)
// 		return nil, fmt.Errorf(errMsg)
// 	}

// 	var costResponse other.CostResponse
// 	err = json.Unmarshal(body, &costResponse)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error unmarshalling cost API response: %v, Raw Body: %s", err, string(body))
// 		return nil, fmt.Errorf("failed to parse cost API response: %w", err)
// 	}

// 	if costResponse.RajaOngkir.Status.Code != 200 {
// 		errMsg := fmt.Sprintf("RajaOngkir internal status error for cost: Code %d, Description: %s",
// 			costResponse.RajaOngkir.Status.Code, costResponse.RajaOngkir.Status.Description)
// 		log.Printf("RajaOngkirService: %s", errMsg)
// 		return nil, fmt.Errorf(errMsg)
// 	}

// 	log.Printf("RajaOngkirService: Successfully fetched cost for courier %s, from %s to %s. Found %d results.",
// 		courier, origin, destination, len(costResponse.RajaOngkir.Results))

// 	return &costResponse, nil
// }

// func (s *RajaOngkirService) GetProvinceByID(provinceID string) (*other.Province, error) {
// 	url := fmt.Sprintf("%s/province?id=%s", s.baseURL, provinceID)

// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error creating request for single province: %v", err)
// 		return nil, fmt.Errorf("failed to create request: %w", err)
// 	}

// 	req.Header.Add("key", s.apiKey)

// 	resp, err := s.client.Do(req)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error performing request to RajaOngkir single province API: %v", err)
// 		return nil, fmt.Errorf("failed to perform request to RajaOngkir: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error reading response body from RajaOngkir single province API: %v", err)
// 		return nil, fmt.Errorf("failed to read response body: %w", err)
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		log.Printf("RajaOngkirService: RajaOngkir single province API returned non-OK status: %d, Body: %s", resp.StatusCode, string(body))
// 		return nil, fmt.Errorf("RajaOngkir API error: status %d, body: %s", resp.StatusCode, string(body))
// 	}

// 	var singleProvinceResponse other.SingleProvinceResponse // KOREKSI: Gunakan struct baru
// 	err = json.Unmarshal(body, &singleProvinceResponse)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error unmarshalling single province API response: %v, Raw Body: %s", err, string(body))
// 		return nil, fmt.Errorf("failed to parse single province API response: %w", err)
// 	}

// 	if singleProvinceResponse.RajaOngkir.Status.Code != 200 {
// 		return nil, fmt.Errorf("RajaOngkir API returned error for single province: %s (Code: %d)", singleProvinceResponse.RajaOngkir.Status.Description, singleProvinceResponse.RajaOngkir.Status.Code)
// 	}

// 	// KOREKSI: Kembalikan objek Province tunggal
// 	return &singleProvinceResponse.RajaOngkir.Results, nil
// }

// // KOREKSI: Gunakan SingleCityResponse
// func (s *RajaOngkirService) GetCityByID(cityID string) (*other.City, error) {
// 	url := fmt.Sprintf("%s/city?id=%s", s.baseURL, cityID)

// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error creating request for single city: %v", err)
// 		return nil, fmt.Errorf("failed to create request: %w", err)
// 	}

// 	req.Header.Add("key", s.apiKey)

// 	resp, err := s.client.Do(req)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error performing request to RajaOngkir single city API: %v", err)
// 		return nil, fmt.Errorf("failed to perform request to RajaOngkir: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error reading response body from RajaOngkir single city API: %v", err)
// 		return nil, fmt.Errorf("failed to read response body: %w", err)
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		log.Printf("RajaOngkirService: RajaOngkir single city API returned non-OK status: %d, Body: %s", resp.StatusCode, string(body))
// 		return nil, fmt.Errorf("RajaOngkir API error: status %d, body: %s", resp.StatusCode, string(body))
// 	}

// 	var singleCityResponse other.SingleCityResponse // KOREKSI: Gunakan struct baru
// 	err = json.Unmarshal(body, &singleCityResponse)
// 	if err != nil {
// 		log.Printf("RajaOngkirService: Error unmarshalling single city API response: %v, Raw Body: %s", err, string(body))
// 		return nil, fmt.Errorf("failed to parse single city API response: %w", err)
// 	}

// 	if singleCityResponse.RajaOngkir.Status.Code != 200 {
// 		return nil, fmt.Errorf("RajaOngkir API returned error for single city: %s (Code: %d)", singleCityResponse.RajaOngkir.Status.Description, singleCityResponse.RajaOngkir.Status.Code)
// 	}

// 	// KOREKSI: Kembalikan objek City tunggal
// 	return &singleCityResponse.RajaOngkir.Results, nil
// }
