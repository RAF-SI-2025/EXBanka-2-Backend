// Package repository — ExchangeRate-API HTTP provider for live currency mid-rates.
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"banka-backend/services/bank-service/internal/domain"
)

const cacheTTL = 15 * time.Minute

// exchangeRateAPIResponse is the JSON response shape from ExchangeRate-API v6.
//
// Example endpoint: GET https://v6.exchangerate-api.com/v6/{KEY}/latest/USD
type exchangeRateAPIResponse struct {
	Result          string             `json:"result"`           // "success" | "error"
	ConversionRates map[string]float64 `json:"conversion_rates"` // target code → rate from base
}

// ExchangeRateProvider fetches live mid rates from ExchangeRate-API v6.
// It uses USD as the base currency (available on all plan tiers) and derives
// RSD-per-unit values for every supported currency.
//
// Kursevi se keširaju u memoriji na cacheTTL (15 min) kako bi se smanjio
// broj poziva eksternog API-ja. RWMutex omogućava da više goroutina čita
// istovremeno, dok samo jedna može osvežiti cache.
type ExchangeRateProvider struct {
	apiKey  string
	baseURL string // e.g. "https://v6.exchangerate-api.com/v6"
	client  *http.Client

	mu          sync.RWMutex
	cachedRates map[string]float64
	cachedAt    time.Time
}

// NewExchangeRateProvider creates a new provider.
// baseURL should not have a trailing slash.
func NewExchangeRateProvider(apiKey, baseURL string) *ExchangeRateProvider {
	return &ExchangeRateProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// GetMidRates vraća srednje kurseve za podržane valute, izražene u RSD.
// Rezultat se kešira na cacheTTL; svežu vrednost povlači tek kada cache istekne.
//
// Konverzija: midRate(X) = conversionRates["RSD"] / conversionRates["X"]
//
// Returns domain.ErrExchangeProviderUnavailable on any network/parse error.
func (p *ExchangeRateProvider) GetMidRates(ctx context.Context) (map[string]float64, error) {
	// Brza provera cache-a — samo read lock.
	p.mu.RLock()
	if p.cachedRates != nil && time.Since(p.cachedAt) < cacheTTL {
		rates := p.cachedRates
		p.mu.RUnlock()
		return rates, nil
	}
	p.mu.RUnlock()

	// Cache je istekao ili prazan — pozovi API pa osveži.
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check: druga goroutina je možda već osvežila cache dok smo čekali na Lock.
	if p.cachedRates != nil && time.Since(p.cachedAt) < cacheTTL {
		return p.cachedRates, nil
	}

	url := fmt.Sprintf("%s/%s/latest/USD", p.baseURL, p.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, domain.ErrExchangeProviderUnavailable
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, domain.ErrExchangeProviderUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, domain.ErrExchangeProviderUnavailable
	}

	var apiResp exchangeRateAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, domain.ErrExchangeProviderUnavailable
	}
	if apiResp.Result != "success" {
		return nil, domain.ErrExchangeProviderUnavailable
	}

	// USD→RSD gives us the RSD price of 1 USD.
	rsdPerUSD, ok := apiResp.ConversionRates["RSD"]
	if !ok || rsdPerUSD == 0 {
		return nil, domain.ErrExchangeProviderUnavailable
	}

	// For each supported currency, derive: 1 X = (USD→RSD) / (USD→X) RSD.
	result := make(map[string]float64, len(domain.SupportedExchangeCodes))
	for _, code := range domain.SupportedExchangeCodes {
		usdToX, exists := apiResp.ConversionRates[code]
		if !exists || usdToX == 0 {
			continue
		}
		result[code] = rsdPerUSD / usdToX
	}

	p.cachedRates = result
	p.cachedAt = time.Now()

	return result, nil
}
