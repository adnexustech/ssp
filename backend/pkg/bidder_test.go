package ssp

import (
	"testing"
	"time"
)

func TestBidRequestBuilder(t *testing.T) {
	builder := NewBidRequestBuilder("test-ssp")

	publisher := &Publisher{
		ID:       "pub-1",
		Name:     "Test Publisher",
		Domain:   "testpub.com",
		RevShare: 0.70,
	}

	site := &Site{
		ID:          "site-1",
		PublisherID: "pub-1",
		Name:        "Test Site",
		Domain:      "testsite.com",
		Cat:         []string{"IAB1"},
	}

	placement := &Placement{
		ID:          "placement-1",
		SiteID:      "site-1",
		Name:        "Test Placement",
		AdType:      "banner",
		Width:       300,
		Height:      250,
		MinBidFloor: 0.50,
	}

	adReq := &AdRequest{
		PlacementID: "placement-1",
		URL:         "https://testsite.com/page",
		Referer:     "https://google.com",
		UserAgent:   "Mozilla/5.0...",
		IP:          "192.168.1.1",
		Width:       300,
		Height:      250,
	}

	bidReq, err := builder.BuildBidRequest(adReq, placement, site, publisher)
	if err != nil {
		t.Fatalf("Failed to build bid request: %v", err)
	}

	if bidReq == nil {
		t.Fatal("Bid request is nil")
	}

	if bidReq.ID == "" {
		t.Error("Bid request ID is empty")
	}

	if len(bidReq.Imp) != 1 {
		t.Errorf("Expected 1 impression, got %d", len(bidReq.Imp))
	}

	imp := bidReq.Imp[0]
	if imp.BidFloor != 0.50 {
		t.Errorf("Expected bid floor 0.50, got %f", imp.BidFloor)
	}

	if imp.Banner == nil {
		t.Fatal("Banner is nil")
	}

	if imp.Banner.W != 300 || imp.Banner.H != 250 {
		t.Errorf("Expected banner size 300x250, got %dx%d", imp.Banner.W, imp.Banner.H)
	}

	if bidReq.Site == nil {
		t.Fatal("Site is nil")
	}

	if bidReq.Site.Domain != "testsite.com" {
		t.Errorf("Expected site domain testsite.com, got %s", bidReq.Site.Domain)
	}

	if bidReq.Device == nil {
		t.Fatal("Device is nil")
	}

	if bidReq.Device.IP != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", bidReq.Device.IP)
	}
}

func TestBidRequestBuilderVideo(t *testing.T) {
	builder := NewBidRequestBuilder("test-ssp")

	publisher := &Publisher{
		ID:     "pub-1",
		Name:   "Test Publisher",
		Domain: "testpub.com",
	}

	site := &Site{
		ID:          "site-1",
		PublisherID: "pub-1",
		Name:        "Test Site",
		Domain:      "testsite.com",
	}

	videoSettings := &VideoSettings{
		Mimes:       []string{"video/mp4"},
		MinDuration: 5,
		MaxDuration: 30,
		Protocols:   []int{2, 3},
		Linearity:   1,
	}

	placement := &Placement{
		ID:          "placement-video-1",
		SiteID:      "site-1",
		Name:        "Video Placement",
		AdType:      "video",
		Width:       640,
		Height:      480,
		MinBidFloor: 1.00,
		Video:       videoSettings,
	}

	adReq := &AdRequest{
		PlacementID: "placement-video-1",
		URL:         "https://testsite.com/video",
		UserAgent:   "Mozilla/5.0...",
		IP:          "192.168.1.1",
	}

	bidReq, err := builder.BuildBidRequest(adReq, placement, site, publisher)
	if err != nil {
		t.Fatalf("Failed to build video bid request: %v", err)
	}

	if len(bidReq.Imp) != 1 {
		t.Fatalf("Expected 1 impression, got %d", len(bidReq.Imp))
	}

	imp := bidReq.Imp[0]
	if imp.Video == nil {
		t.Fatal("Video is nil")
	}

	if imp.Video.W != 640 || imp.Video.H != 480 {
		t.Errorf("Expected video size 640x480, got %dx%d", imp.Video.W, imp.Video.H)
	}

	if imp.Video.MaxDuration != 30 {
		t.Errorf("Expected max duration 30, got %d", imp.Video.MaxDuration)
	}

	if len(imp.Video.Mimes) != 1 || imp.Video.Mimes[0] != "video/mp4" {
		t.Errorf("Expected mime type video/mp4, got %v", imp.Video.Mimes)
	}
}

func TestAuctionEngine(t *testing.T) {
	engine := NewAuctionEngine(0.10)

	partner1 := &DemandPartner{
		ID:   "partner-1",
		Name: "Partner 1",
	}

	partner2 := &DemandPartner{
		ID:   "partner-2",
		Name: "Partner 2",
	}

	placement := &Placement{
		ID:          "placement-1",
		MinBidFloor: 0.10,
	}

	// Create bid responses
	response1 := &BidResponse{
		ID: "resp-1",
		SeatBid: []SeatBid{
			{
				Bid: []Bid{
					{
						ID:    "bid-1",
						ImpID: "imp-1",
						Price: 1.50,
						ADM:   "<ad>markup1</ad>",
					},
				},
			},
		},
	}

	response2 := &BidResponse{
		ID: "resp-2",
		SeatBid: []SeatBid{
			{
				Bid: []Bid{
					{
						ID:    "bid-2",
						ImpID: "imp-1",
						Price: 2.00,
					},
				},
			},
		},
	}

	responses := map[*DemandPartner]*BidResponse{
		partner1: response1,
		partner2: response2,
	}

	result, err := engine.RunAuction(responses, placement)
	if err != nil {
		t.Fatalf("Auction failed: %v", err)
	}

	if result == nil {
		t.Fatal("Auction result is nil")
	}

	// Winner should be highest bidder (partner2 with 2.00)
	if result.WinningBid.Price != 2.00 {
		t.Errorf("Expected winning price 2.00, got %f", result.WinningBid.Price)
	}

	if result.WinningPartner.ID != "partner-2" {
		t.Errorf("Expected winner partner-2, got %s", result.WinningPartner.ID)
	}

	// Cleared price should be second highest (1.50)
	if result.ClearedPrice != 1.50 {
		t.Errorf("Expected cleared price 1.50, got %f", result.ClearedPrice)
	}

	if result.AuctionType != 2 {
		t.Errorf("Expected auction type 2 (second price), got %d", result.AuctionType)
	}
}

func TestAuctionEngineNoBids(t *testing.T) {
	engine := NewAuctionEngine(0.10)

	placement := &Placement{
		ID:          "placement-1",
		MinBidFloor: 0.10,
	}

	responses := map[*DemandPartner]*BidResponse{}

	result, err := engine.RunAuction(responses, placement)
	if err != nil {
		t.Fatalf("Auction failed: %v", err)
	}

	if result != nil {
		t.Error("Expected nil result for no bids")
	}
}

func TestAuctionEngineBelowFloor(t *testing.T) {
	engine := NewAuctionEngine(0.10)

	partner := &DemandPartner{
		ID:   "partner-1",
		Name: "Partner 1",
	}

	placement := &Placement{
		ID:          "placement-1",
		MinBidFloor: 1.00,
	}

	response := &BidResponse{
		ID: "resp-1",
		SeatBid: []SeatBid{
			{
				Bid: []Bid{
					{
						ID:    "bid-1",
						ImpID: "imp-1",
						Price: 0.50, // Below floor
					},
				},
			},
		},
	}

	responses := map[*DemandPartner]*BidResponse{
		partner: response,
	}

	result, err := engine.RunAuction(responses, placement)
	if err != nil {
		t.Fatalf("Auction failed: %v", err)
	}

	if result != nil {
		t.Error("Expected nil result for bids below floor")
	}
}

func TestBidder(t *testing.T) {
	bidder := NewBidder("test-ssp", 100*time.Millisecond)

	if bidder == nil {
		t.Fatal("Bidder is nil")
	}

	if bidder.sspID != "test-ssp" {
		t.Errorf("Expected SSP ID test-ssp, got %s", bidder.sspID)
	}

	if bidder.timeout != 100*time.Millisecond {
		t.Errorf("Expected timeout 100ms, got %v", bidder.timeout)
	}
}

func TestPartnerManager(t *testing.T) {
	pm := NewPartnerManager()

	partner1 := &SupplyPartner{
		ID:       "partner-1",
		Name:     "Test Partner 1",
		Type:     "openrtb",
		Endpoint: "http://partner1.com/rtb",
		Active:   true,
		QPS:      100,
		RevShare: 0.30,
	}

	partner2 := &SupplyPartner{
		ID:       "partner-2",
		Name:     "Test Partner 2",
		Type:     "exads",
		Endpoint: "http://partner2.com/rtb",
		Active:   false,
		QPS:      200,
		RevShare: 0.25,
	}

	pm.AddPartner(partner1)
	pm.AddPartner(partner2)

	// Test GetPartner
	p, ok := pm.GetPartner("partner-1")
	if !ok {
		t.Error("Partner 1 not found")
	}
	if p.Name != "Test Partner 1" {
		t.Errorf("Expected partner name Test Partner 1, got %s", p.Name)
	}

	// Test GetActivePartners
	active := pm.GetActivePartners()
	if len(active) != 1 {
		t.Errorf("Expected 1 active partner, got %d", len(active))
	}

	if active[0].ID != "partner-1" {
		t.Errorf("Expected active partner-1, got %s", active[0].ID)
	}
}

func BenchmarkBidRequestBuilder(b *testing.B) {
	builder := NewBidRequestBuilder("bench-ssp")

	publisher := &Publisher{
		ID:     "pub-1",
		Name:   "Bench Publisher",
		Domain: "bench.com",
	}

	site := &Site{
		ID:          "site-1",
		PublisherID: "pub-1",
		Name:        "Bench Site",
		Domain:      "bench.com",
	}

	placement := &Placement{
		ID:          "placement-1",
		SiteID:      "site-1",
		AdType:      "banner",
		Width:       300,
		Height:      250,
		MinBidFloor: 0.50,
	}

	adReq := &AdRequest{
		PlacementID: "placement-1",
		URL:         "https://bench.com/page",
		UserAgent:   "Mozilla/5.0...",
		IP:          "192.168.1.1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = builder.BuildBidRequest(adReq, placement, site, publisher)
	}
}

func BenchmarkAuctionEngine(b *testing.B) {
	engine := NewAuctionEngine(0.10)

	partner1 := &DemandPartner{ID: "p1", Name: "Partner 1"}
	partner2 := &DemandPartner{ID: "p2", Name: "Partner 2"}
	partner3 := &DemandPartner{ID: "p3", Name: "Partner 3"}

	placement := &Placement{
		ID:          "placement-1",
		MinBidFloor: 0.10,
	}

	responses := map[*DemandPartner]*BidResponse{
		partner1: {
			SeatBid: []SeatBid{
				{Bid: []Bid{{ID: "b1", Price: 1.50}}},
			},
		},
		partner2: {
			SeatBid: []SeatBid{
				{Bid: []Bid{{ID: "b2", Price: 2.00}}},
			},
		},
		partner3: {
			SeatBid: []SeatBid{
				{Bid: []Bid{{ID: "b3", Price: 1.75}}},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.RunAuction(responses, placement)
	}
}
