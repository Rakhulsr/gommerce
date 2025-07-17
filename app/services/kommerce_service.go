package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Rakhulsr/go-ecommerce/app/models/other"
)

type KomerceRajaOngkirClient interface {
	GetDomesticShippingCost(ctx context.Context, originID, destinationID string, weight int, courier string) ([]other.KomerceCostResult, error)
}

type komerceRajaOngkirService struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

func NewKomerceRajaOngkirService() KomerceRajaOngkirClient {
	const hardcodedBaseURL = "https://rajaongkir.komerce.id/api"
	const hardcodedAPIKey = "I1lMHC0ib9559cde0a34dd9dRs9LxUYa"

	return &komerceRajaOngkirService{
		client:  &http.Client{Timeout: 10 * time.Second},
		apiKey:  hardcodedAPIKey,
		baseURL: hardcodedBaseURL,
	}
}

func (s *komerceRajaOngkirService) doRequest(ctx context.Context, method, path string, payload interface{}) ([]byte, error) {
	requestURL := fmt.Sprintf("%s%s", s.baseURL, path)
	var reqBody []byte
	var err error

	if payload != nil {
		reqBody, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("gagal membuat payload JSON: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request: %w", err)
	}

	req.Header.Add("key", s.apiKey)
	if method == "POST" || payload != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	log.Printf("KomerceRajaOngkirService.doRequest: Requesting URL: %s", requestURL)
	if payload != nil {
		log.Printf("KomerceRajaOngkirService.doRequest: Request Payload: %s", string(reqBody))
	}
	apiKeyLog := "API Key not set or too short"
	if len(s.apiKey) >= 5 {
		apiKeyLog = s.apiKey[:5]
	}
	log.Printf("KomerceRajaOngkirService.doRequest: API Key: %s (first 5 chars)", apiKeyLog)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal melakukan request ke Komerce API: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons body: %w", err)
	}

	log.Printf("KomerceRajaOngkirService.doRequest: Response Status: %d, Body: %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Komerce API mengembalikan status error: %d - %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (s *komerceRajaOngkirService) GetDomesticShippingCost(ctx context.Context, originID, destinationID string, weight int, courier string) ([]other.KomerceCostResult, error) {
	if s.baseURL == "" {
		return nil, fmt.Errorf("Komerce API Base URL belum diatur di service (s.baseURL kosong)")
	}
	if s.apiKey == "" {
		return nil, fmt.Errorf("Komerce API Key belum diatur di service (s.apiKey kosong)")
	}

	payload := map[string]interface{}{
		"origin":      originID,
		"destination": destinationID,
		"weight":      weight,
		"courier":     courier,
	}

	body, err := s.doRequest(ctx, "POST", "/v1/cost/domestic-cost", payload)
	if err != nil {
		return nil, err
	}

	var apiResponse other.KomerceShippingCostResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		log.Printf("KomerceRajaOngkirService.GetDomesticShippingCost: Gagal unmarshal respons JSON: %v, Body: %s", err, string(body))
		return nil, fmt.Errorf("gagal mengurai respons JSON: %w", err)
	}

	if apiResponse.Meta.Code != 200 {
		return nil, fmt.Errorf("Komerce API mengembalikan status non-200: %d - %s", apiResponse.Meta.Code, apiResponse.Meta.Message)
	}

	if len(apiResponse.Data) == 0 || len(apiResponse.Data[0].Costs) == 0 {
		return []other.KomerceCostResult{}, nil
	}

	return apiResponse.Data[0].Costs, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
