package ssp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Bidder handles OpenRTB 2.5 bid requests to demand partners
type Bidder struct {
	client  *http.Client
	sspID   string
	timeout time.Duration
}

// NewBidder creates a new bidder instance
func NewBidder(sspID string, timeout time.Duration) *Bidder {
	return &Bidder{
		client: &http.Client{
			Timeout: timeout,
		},
		sspID:   sspID,
		timeout: timeout,
	}
}

// DemandPartner represents a demand-side partner (DSP, ADX, exchange)
type DemandPartner struct {
	ID       string
	Name     string
	Endpoint string
	Timeout  time.Duration
	Active   bool
	QPS      int     // Queries per second limit
	RevShare float64 // SSP revenue share (0.0-1.0)
}

// SendBidRequest sends a bid request to a demand partner
func (b *Bidder) SendBidRequest(ctx context.Context, bidReq *BidRequest, partner *DemandPartner) (*BidResponse, error) {
	// Marshal bid request
	reqBody, err := json.Marshal(bidReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bid request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", partner.Endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-OpenRTB-Version", "2.5")

	// Send request
	resp, err := b.client.Do(req)
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
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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

// BidRequestBuilder builds OpenRTB 2.5 bid requests from placements
type BidRequestBuilder struct {
	sspID string
}

// NewBidRequestBuilder creates a new bid request builder
func NewBidRequestBuilder(sspID string) *BidRequestBuilder {
	return &BidRequestBuilder{
		sspID: sspID,
	}
}

// AdRequest represents an incoming ad request from publisher
type AdRequest struct {
	PlacementID string
	URL         string
	Referer     string
	UserAgent   string
	IP          string
	Width       int
	Height      int
	// Additional params
	Params map[string]interface{}
}

// BuildBidRequest builds an OpenRTB 2.5 bid request from ad request and placement
func (b *BidRequestBuilder) BuildBidRequest(adReq *AdRequest, placement *Placement, site *Site, pub *Publisher) (*BidRequest, error) {
	// Generate request ID
	reqID := uuid.New().String()

	// Build impression
	imp := Impression{
		ID:          uuid.New().String(),
		TagID:       placement.ID,
		BidFloor:    placement.MinBidFloor,
		BidFloorCur: "USD",
		Secure:      1, // Assume HTTPS
	}

	// Add impression type based on ad type
	switch placement.AdType {
	case "banner":
		imp.Banner = b.buildBanner(placement)
	case "video":
		imp.Video = b.buildVideo(placement)
	case "native":
		imp.Native = &Native{
			Request: `{"ver":"1.2"}`, // Placeholder
			Ver:     "1.2",
		}
	default:
		return nil, fmt.Errorf("unsupported ad type: %s", placement.AdType)
	}

	// Build site object
	siteInfo := &SiteInfo{
		ID:     site.ID,
		Name:   site.Name,
		Domain: site.Domain,
		Page:   adReq.URL,
		Ref:    adReq.Referer,
		Cat:    site.Cat,
		Publisher: &Publisher2{
			ID:     pub.ID,
			Name:   pub.Name,
			Domain: pub.Domain,
		},
	}

	// Build device object
	device := &Device{
		UA: adReq.UserAgent,
		IP: adReq.IP,
	}

	// Build bid request
	bidReq := &BidRequest{
		ID:     reqID,
		Imp:    []Impression{imp},
		Site:   siteInfo,
		Device: device,
		At:     2,   // Second price auction
		Tmax:   120, // 120ms timeout
		Cur:    []string{"USD"},
		Source: &Source{
			FD:  1, // Final destination
			TID: uuid.New().String(),
		},
	}

	return bidReq, nil
}

func (b *BidRequestBuilder) buildBanner(placement *Placement) *Banner {
	banner := &Banner{
		ID:  placement.ID,
		Pos: 1, // Above the fold
	}

	// Add formats
	if len(placement.Formats) > 0 {
		banner.Format = placement.Formats
	} else if placement.Width > 0 && placement.Height > 0 {
		banner.W = placement.Width
		banner.H = placement.Height
		banner.Format = []Format{
			{W: placement.Width, H: placement.Height},
		}
	}

	return banner
}

func (b *BidRequestBuilder) buildVideo(placement *Placement) *Video {
	video := &Video{
		W:   placement.Width,
		H:   placement.Height,
		Pos: 1,
	}

	// Add video settings if available
	if placement.Video != nil {
		video.Mimes = placement.Video.Mimes
		video.MinDuration = placement.Video.MinDuration
		video.MaxDuration = placement.Video.MaxDuration
		video.Protocols = placement.Video.Protocols
		video.Linearity = placement.Video.Linearity
		video.StartDelay = placement.Video.StartDelay
		video.PlaybackMethod = placement.Video.PlaybackMethod
		video.API = placement.Video.API
	} else {
		// Default video settings
		video.Mimes = []string{"video/mp4", "video/webm"}
		video.MinDuration = 5
		video.MaxDuration = 30
		video.Protocols = []int{2, 3, 5, 6} // VAST 2.0, 3.0, VAST 4.0
		video.Linearity = 1                 // Linear
	}

	return video
}

// AuctionEngine handles auction logic for bid responses
type AuctionEngine struct {
	minBidFloor float64
}

// NewAuctionEngine creates a new auction engine
func NewAuctionEngine(minBidFloor float64) *AuctionEngine {
	return &AuctionEngine{
		minBidFloor: minBidFloor,
	}
}

// AuctionResult represents the result of an auction
type AuctionResult struct {
	WinningBid     *Bid
	WinningPartner *DemandPartner
	AllBids        []BidWithPartner
	AuctionType    int // 1=first price, 2=second price
	ClearedPrice   float64
}

// BidWithPartner combines a bid with its partner info
type BidWithPartner struct {
	Bid     *Bid
	Partner *DemandPartner
}

// RunAuction runs a second-price auction on bid responses
func (ae *AuctionEngine) RunAuction(responses map[*DemandPartner]*BidResponse, placement *Placement) (*AuctionResult, error) {
	allBids := []BidWithPartner{}

	// Collect all bids
	for partner, response := range responses {
		if response == nil {
			continue
		}

		for _, seatBid := range response.SeatBid {
			for _, bid := range seatBid.Bid {
				// Filter by bid floor
				if bid.Price >= placement.MinBidFloor {
					allBids = append(allBids, BidWithPartner{
						Bid:     &bid,
						Partner: partner,
					})
				}
			}
		}
	}

	// No bids
	if len(allBids) == 0 {
		return nil, nil
	}

	// Sort bids by price (descending)
	for i := 0; i < len(allBids)-1; i++ {
		for j := i + 1; j < len(allBids); j++ {
			if allBids[j].Bid.Price > allBids[i].Bid.Price {
				allBids[i], allBids[j] = allBids[j], allBids[i]
			}
		}
	}

	// Winner is highest bid
	winner := allBids[0]

	// Calculate cleared price (second price auction)
	clearedPrice := placement.MinBidFloor
	if len(allBids) > 1 {
		// Second highest bid
		clearedPrice = allBids[1].Bid.Price
	}

	// But never less than bid floor
	if clearedPrice < placement.MinBidFloor {
		clearedPrice = placement.MinBidFloor
	}

	return &AuctionResult{
		WinningBid:     winner.Bid,
		WinningPartner: winner.Partner,
		AllBids:        allBids,
		AuctionType:    2, // Second price
		ClearedPrice:   clearedPrice,
	}, nil
}
