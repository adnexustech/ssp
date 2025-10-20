package ssp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	openrtb2 "github.com/prebid/openrtb/v20/openrtb2"
)

// BidsCubePartner handles integration with BidsCube White Label
type BidsCubePartner struct {
	endpoint string
	apiKey   string
	client   *http.Client
	revShare float64 // Revenue share for SSP
}

// NewBidsCubePartner creates a new BidsCube partner integration
func NewBidsCubePartner(endpoint, apiKey string, revShare float64) *BidsCubePartner {
	return &BidsCubePartner{
		endpoint: endpoint,
		apiKey:   apiKey,
		client: &http.Client{
			Timeout: 100 * time.Millisecond, // OpenRTB timeout
		},
		revShare: revShare,
	}
}

// SendBidRequest sends OpenRTB request to BidsCube
func (b *BidsCubePartner) SendBidRequest(ctx context.Context, req *openrtb2.BidRequest) (*openrtb2.BidResponse, error) {
	// Add BidsCube specific extensions
	bidsCubeReq := b.transformRequest(req)
	
	// Marshal request
	body, err := json.Marshal(bidsCubeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", b.endpoint+"/openrtb2/auction", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", b.apiKey)
	httpReq.Header.Set("X-OpenRTB-Version", "2.5")
	
	// Send request
	resp, err := b.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Handle no bid (204)
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	
	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	
	// Parse response
	var bidResp openrtb2.BidResponse
	if err := json.NewDecoder(resp.Body).Decode(&bidResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Apply revenue share to bids
	b.applyRevShare(&bidResp)
	
	return &bidResp, nil
}

// transformRequest adds BidsCube specific fields
func (b *BidsCubePartner) transformRequest(req *openrtb2.BidRequest) *openrtb2.BidRequest {
	// Clone request to avoid modifying original
	transformed := *req
	
	// Add source if not present
	if transformed.Source == nil {
		transformed.Source = &openrtb2.Source{}
	}
	
	// Add SSP information
	transformed.Source.FD = openrtb2.Int8Ptr(1) // First-party data
	transformed.Source.TID = uuid.New().String() // Transaction ID
	
	// Add extension data for BidsCube
	if transformed.Ext == nil {
		transformed.Ext = json.RawMessage(`{}`)
	}
	
	// Add SSP identifier
	extData := map[string]interface{}{
		"ssp": "adnexus-ssp",
		"integration": "adnexus-wl",
		"version": "1.0.0",
	}
	
	if extBytes, err := json.Marshal(extData); err == nil {
		transformed.Ext = extBytes
	}
	
	return &transformed
}

// applyRevShare adjusts bid prices based on revenue share
func (b *BidsCubePartner) applyRevShare(resp *openrtb2.BidResponse) {
	if resp == nil || len(resp.SeatBid) == 0 {
		return
	}
	
	for i := range resp.SeatBid {
		for j := range resp.SeatBid[i].Bid {
			// Apply SSP revenue share (we keep revShare%, publisher gets rest)
			originalPrice := resp.SeatBid[i].Bid[j].Price
			resp.SeatBid[i].Bid[j].Price = originalPrice * (1 - b.revShare)
		}
	}
}

// GetMetrics returns partner metrics
func (b *BidsCubePartner) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"endpoint": b.endpoint,
		"rev_share": b.revShare,
		"timeout_ms": b.client.Timeout.Milliseconds(),
	}
}