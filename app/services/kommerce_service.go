package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time" // Pastikan time diimpor

	"github.com/Rakhulsr/go-ecommerce/app/models/other"
)

// Global variables for in-memory caching (tetap ada sesuai strategi terakhir)
var (
	allDomesticDestinations      []other.KomerceDomesticDestination
	domesticDestinationsMutex    sync.RWMutex
	lastDomesticDestinationsSync time.Time
)

type KomerceRajaOngkirClient interface {
	CalculateCost(ctx context.Context, originID, destinationID int, weight int, courier string) ([]other.KomerceCostDetail, error)
	SearchDomesticDestinations(ctx context.Context, query string, limit, offset int) ([]other.KomerceDomesticDestination, error)
}

type komerceRajaOngkirService struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

func NewKomerceRajaOngkirClient(apiKey string) KomerceRajaOngkirClient {
	return &komerceRajaOngkirService{
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: "https://rajaongkir.komerce.id/api",
	}
}

// doRequest helper function
func (s *komerceRajaOngkirService) doRequest(ctx context.Context, method, fullPath string, bodyReader *bytes.Buffer, contentType string) ([]byte, error) {

	fullURL := s.baseURL + fullPath
	log.Printf("doRequest: Membuat request %s ke URL: %s dengan Content-Type: %s", method, fullURL, contentType)

	// Pastikan bodyReader tidak nil, jika nil, gunakan bytes.NewBuffer(nil) atau http.NoBody
	var reqBodyReader *bytes.Buffer
	if bodyReader == nil {
		reqBodyReader = bytes.NewBuffer(nil) // Gunakan buffer kosong jika bodyReader nil
	} else {
		reqBodyReader = bodyReader
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBodyReader)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal melakukan request HTTP: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Komerce mengembalikan status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// CalculateCost (tidak ada perubahan)
func (s *komerceRajaOngkirService) CalculateCost(ctx context.Context, originID, destinationID int, weight int, courier string) ([]other.KomerceCostDetail, error) {
	formData := url.Values{}
	formData.Add("origin", fmt.Sprintf("%d", originID))
	formData.Add("destination", fmt.Sprintf("%d", destinationID))
	formData.Add("weight", fmt.Sprintf("%d", weight))
	formData.Add("courier", courier)
	formData.Add("price", "lowest")

	requestBodyReader := bytes.NewBufferString(formData.Encode())
	contentType := "application/x-www-form-urlencoded"

	log.Printf("CalculateCost: Mengirim form-urlencoded body ke Komerce API: %s", formData.Encode())

	body, err := s.doRequest(ctx, "POST", "/v1/calculate/domestic-cost", requestBodyReader, contentType)
	if err != nil {
		return nil, err
	}

	var apiResponse other.KomerceShippingCostResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("gagal mengurai respons JSON biaya pengiriman: %w", err)
	}

	if apiResponse.Meta.Code != 200 || apiResponse.Meta.Status != "success" {
		return nil, fmt.Errorf("API Komerce mengembalikan status error: %d - %s", apiResponse.Meta.Code, apiResponse.Meta.Message)
	}

	return apiResponse.Data, nil
}

// SearchDomesticDestinations (PERBAIKAN PANGGILAN doRequest)
func (s *komerceRajaOngkirService) SearchDomesticDestinations(ctx context.Context, query string, limit, offset int) ([]other.KomerceDomesticDestination, error) {
	params := url.Values{}
	params.Add("search", query)
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))

	fullPath := fmt.Sprintf("/v1/destination/domestic-destination?%s", params.Encode())

	// PERBAIKAN DI SINI: Berikan bytes.NewBuffer(nil) sebagai bodyReader untuk GET request
	body, err := s.doRequest(ctx, "GET", fullPath, bytes.NewBuffer(nil), "application/json")
	if err != nil {
		return nil, err
	}

	var apiResponse other.KomerceDomesticDestinationResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("gagal mengurai respons JSON destinasi domestik: %w", err)
	}

	if apiResponse.Meta.Code != 200 || apiResponse.Meta.Status != "success" {
		return nil, fmt.Errorf("API Komerce mengembalikan status error: %d - %s", apiResponse.Meta.Code, apiResponse.Meta.Message)
	}

	return apiResponse.Data, nil
}
