package ssp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EXADSClient handles integration with EXADS RTB exchange
type EXADSClient struct {
	client   *http.Client
	endpoint string
	apiKey   string
	timeout  time.Duration
}

// NewEXADSClient creates a new EXADS client
func NewEXADSClient(endpoint, apiKey string, timeout time.Duration) *EXADSClient {
	return &EXADSClient{
		client: &http.Client{
			Timeout: timeout,
		},
		endpoint: endpoint,
		apiKey:   apiKey,
		timeout:  timeout,
	}
}

// SendBidRequest sends an OpenRTB 2.5 bid request to EXADS
func (ec *EXADSClient) SendBidRequest(ctx context.Context, bidReq *BidRequest) (*BidResponse, error) {
	// Marshal bid request
	reqBody, err := json.Marshal(bidReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bid request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", ec.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-OpenRTB-Version", "2.5")

	// Add EXADS-specific headers
	if ec.apiKey != "" {
		req.Header.Set("X-EXADS-API-Key", ec.apiKey)
	}

	// Send request
	resp, err := ec.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == http.StatusNoContent {
		// No bid
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var bidResp BidResponse
	if err := json.Unmarshal(respBody, &bidResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bid response: %w", err)
	}

	return &bidResp, nil
}

// SendWinNotice sends a win notice to EXADS
func (ec *EXADSClient) SendWinNotice(ctx context.Context, nurl string, price float64) error {
	// Replace ${AUCTION_PRICE} macro
	finalURL := nurl
	// Note: In production, you'd replace macros like ${AUCTION_PRICE}, ${AUCTION_CURRENCY}, etc.

	req, err := http.NewRequestWithContext(ctx, "GET", finalURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create win notice request: %w", err)
	}

	resp, err := ec.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send win notice: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("win notice failed with status: %d", resp.StatusCode)
	}

	return nil
}

// SupplyPartner represents an external supply partner configuration
type SupplyPartner struct {
	ID       string
	Name     string
	Type     string // "exads", "openrtb", "custom"
	Endpoint string
	APIKey   string
	Timeout  time.Duration
	Active   bool
	QPS      int     // Queries per second limit
	RevShare float64 // Partner revenue share (0.0-1.0)
}

// PartnerManager manages multiple supply partners
type PartnerManager struct {
	partners    map[string]*SupplyPartner
	exadsClient *EXADSClient
}

// NewPartnerManager creates a new partner manager
func NewPartnerManager() *PartnerManager {
	return &PartnerManager{
		partners: make(map[string]*SupplyPartner),
	}
}

// AddPartner adds a supply partner
func (pm *PartnerManager) AddPartner(partner *SupplyPartner) {
	pm.partners[partner.ID] = partner

	// Initialize EXADS client if needed
	if partner.Type == "exads" && pm.exadsClient == nil {
		pm.exadsClient = NewEXADSClient(partner.Endpoint, partner.APIKey, partner.Timeout)
	}
}

// GetPartner retrieves a partner by ID
func (pm *PartnerManager) GetPartner(id string) (*SupplyPartner, bool) {
	partner, ok := pm.partners[id]
	return partner, ok
}

// GetActivePartners returns all active partners
func (pm *PartnerManager) GetActivePartners() []*SupplyPartner {
	active := []*SupplyPartner{}
	for _, p := range pm.partners {
		if p.Active {
			active = append(active, p)
		}
	}
	return active
}

// SendToPartner sends a bid request to appropriate partner
func (pm *PartnerManager) SendToPartner(ctx context.Context, partner *SupplyPartner, bidReq *BidRequest) (*BidResponse, error) {
	switch partner.Type {
	case "exads":
		if pm.exadsClient == nil {
			pm.exadsClient = NewEXADSClient(partner.Endpoint, partner.APIKey, partner.Timeout)
		}
		return pm.exadsClient.SendBidRequest(ctx, bidReq)
	case "openrtb":
		// Generic OpenRTB client
		client := &http.Client{Timeout: partner.Timeout}
		reqBody, err := json.Marshal(bidReq)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, "POST", partner.Endpoint, bytes.NewReader(reqBody))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-OpenRTB-Version", "2.5")

		if partner.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+partner.APIKey)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNoContent {
			return nil, nil
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}

		var bidResp BidResponse
		if err := json.NewDecoder(resp.Body).Decode(&bidResp); err != nil {
			return nil, err
		}

		return &bidResp, nil
	default:
		return nil, fmt.Errorf("unsupported partner type: %s", partner.Type)
	}
}
