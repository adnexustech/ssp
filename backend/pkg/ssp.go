package ssp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/prebid/openrtb/v20/openrtb2"
)

// SSP represents the Supply-Side Platform core
type SSP struct {
	partnerManager *PartnerManager
	auctionEngine  *AuctionEngine
	bidder         *Bidder
	logger         *slog.Logger
}

// NewSSP creates a new SSP instance
func NewSSP(
	partnerManager *PartnerManager,
	auctionEngine *AuctionEngine,
	bidder *Bidder,
	logger *slog.Logger,
) *SSP {
	return &SSP{
		partnerManager: partnerManager,
		auctionEngine:  auctionEngine,
		bidder:         bidder,
		logger:         logger,
	}
}

// processRequest handles an OpenRTB bid request by sending it to partners and running an auction
func (s *SSP) processRequest(ctx context.Context, bidRequest *openrtb2.BidRequest) (*openrtb2.BidResponse, error) {
	// Get active partners
	partners := s.partnerManager.GetActivePartners()
	if len(partners) == 0 {
		return nil, fmt.Errorf("no active partners available")
	}

	// Send bid requests to all partners in parallel
	type partnerResult struct {
		partner  *SupplyPartner
		response *openrtb2.BidResponse
		err      error
	}

	resultCh := make(chan partnerResult, len(partners))
	var wg sync.WaitGroup

	for _, partner := range partners {
		wg.Add(1)
		go func(p *SupplyPartner) {
			defer wg.Done()

			// Set timeout for this partner
			partnerCtx, cancel := context.WithTimeout(ctx, p.Timeout)
			defer cancel()

			// Send bid request based on partner type
			var response *openrtb2.BidResponse
			var err error

			switch p.Type {
			case "adnexus":
				// Create BidsCube client and send request
				adnexusPartner := NewBidsCubePartner(p.Endpoint, p.APIKey, p.RevShare)
				response, err = adnexusPartner.SendBidRequest(partnerCtx, bidRequest)
			case "dsp":
				// Direct OpenRTB request to DSP
				response, err = s.sendOpenRTBRequest(partnerCtx, p, bidRequest)
			default:
				// Generic OpenRTB request
				response, err = s.sendOpenRTBRequest(partnerCtx, p, bidRequest)
			}

			resultCh <- partnerResult{
				partner:  p,
				response: response,
				err:      err,
			}
		}(partner)
	}

	// Wait for all partners to respond or timeout
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect responses
	var bidResponses []*openrtb2.BidResponse
	partnerResponses := make(map[string]*openrtb2.BidResponse)

	for result := range resultCh {
		if result.err != nil {
			s.logger.Warn("Partner bid request failed",
				"partner", result.partner.Name,
				"error", result.err)
			continue
		}

		if result.response != nil && len(result.response.SeatBid) > 0 {
			bidResponses = append(bidResponses, result.response)
			partnerResponses[result.partner.ID] = result.response
		}
	}

	// If no valid responses, return nil
	if len(bidResponses) == 0 {
		return nil, fmt.Errorf("no valid bid responses received")
	}

	// Run auction to determine winner
	winningResponse := s.selectWinner(bidResponses)
	if winningResponse == nil {
		return nil, fmt.Errorf("no winning bid")
	}

	return winningResponse, nil
}

// sendOpenRTBRequest sends an OpenRTB request to a partner endpoint
func (s *SSP) sendOpenRTBRequest(ctx context.Context, partner *SupplyPartner, bidRequest *openrtb2.BidRequest) (*openrtb2.BidResponse, error) {
	// Marshal bid request
	requestBody, err := json.Marshal(bidRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bid request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", partner.Endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-OpenRTB-Version", "2.5")

	// Add API key if provided
	if partner.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+partner.APIKey)
	}

	// Send request
	client := &http.Client{
		Timeout: partner.Timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil // No bid
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var bidResponse openrtb2.BidResponse
	if err := json.NewDecoder(resp.Body).Decode(&bidResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &bidResponse, nil
}

// selectWinner selects the winning bid from multiple responses
func (s *SSP) selectWinner(responses []*openrtb2.BidResponse) *openrtb2.BidResponse {
	var winningResponse *openrtb2.BidResponse
	var highestPrice float64

	for _, response := range responses {
		for _, seatBid := range response.SeatBid {
			for _, bid := range seatBid.Bid {
				if bid.Price > highestPrice {
					highestPrice = bid.Price
					winningResponse = response
				}
			}
		}
	}

	return winningResponse
}