package ssp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prebid/openrtb/v20/openrtb2"
)

// PublicaSSAIRequest represents the Server-Side Ad Insertion request from Publica
type PublicaSSAIRequest struct {
	PublisherID    string                 `json:"publisher_id"`
	SiteID         string                 `json:"site_id"`
	ContentID      string                 `json:"content_id"`
	DeviceID       string                 `json:"device_id"`
	IP             string                 `json:"ip"`
	UserAgent      string                 `json:"ua"`
	FloorPrice     float64                `json:"floor_price"`
	DealID         string                 `json:"deal_id,omitempty"`
	Params         map[string]interface{} `json:"params,omitempty"`
	ContentGenre   string                 `json:"content_genre,omitempty"`
	ContentRating  string                 `json:"content_rating,omitempty"`
	ContentLanguage string                `json:"content_language,omitempty"`
	Duration       int                    `json:"duration,omitempty"`
}

// PublicaSSAIResponse represents the response to Publica for ad decisioning
type PublicaSSAIResponse struct {
	VASTURL       string   `json:"vast_url"`
	Duration      int      `json:"duration"`
	AdBreakID     string   `json:"ad_break_id"`
	Ads           []Ad     `json:"ads"`
	TrackingURLs  Tracking `json:"tracking_urls"`
	CacheBuster   string   `json:"cache_buster"`
}

type Ad struct {
	ID          string  `json:"id"`
	Duration    int     `json:"duration"`
	MediaURL    string  `json:"media_url"`
	ClickURL    string  `json:"click_url"`
	Title       string  `json:"title"`
	Advertiser  string  `json:"advertiser"`
	CPM         float64 `json:"cpm"`
}

type Tracking struct {
	Impression []string `json:"impression"`
	Click      []string `json:"click"`
	Complete   []string `json:"complete"`
	Quartile1  []string `json:"quartile_1"`
	Quartile2  []string `json:"quartile_2"`
	Quartile3  []string `json:"quartile_3"`
}

// PublicaHandler handles Publica-specific endpoints
type PublicaHandler struct {
	ssp *SSP
}

// NewPublicaHandler creates a new Publica handler
func NewPublicaHandler(ssp *SSP) *PublicaHandler {
	return &PublicaHandler{ssp: ssp}
}

// HandleSSAI handles Server-Side Ad Insertion requests from Publica
func (h *PublicaHandler) HandleSSAI(c *gin.Context) {
	var req PublicaSSAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert Publica request to OpenRTB bid request
	bidRequest := h.convertToOpenRTB(&req)

	// Process the bid request through the SSP
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Get bids from partners (including BidsCube if it's a P1 publisher)
	bidResponse, err := h.ssp.processRequest(ctx, bidRequest)
	if err != nil {
		c.JSON(http.StatusNoContent, gin.H{"message": "No ads available"})
		return
	}

	// Convert OpenRTB response to Publica format
	publicaResponse := h.convertToPublicaResponse(bidResponse, &req)

	c.JSON(http.StatusOK, publicaResponse)
}

// HandleVAST generates VAST response for Publica
func (h *PublicaHandler) HandleVAST(c *gin.Context) {
	// Parse query parameters
	pubID := c.Query("pub")
	siteID := c.Query("site")
	contentID := c.Query("content_id")
	deviceID := c.Query("device_id")
	ip := c.Query("ip")
	ua := c.Query("ua")
	floorStr := c.Query("floor")
	dealID := c.Query("deal")

	floor, _ := strconv.ParseFloat(floorStr, 64)
	if floor == 0 {
		floor = 0.50 // Default floor
	}

	// Create OpenRTB request
	bidRequest := &openrtb2.BidRequest{
		ID: uuid.New().String(),
		Imp: []openrtb2.Imp{
			{
				ID: "1",
				Video: &openrtb2.Video{
					W:           ptrInt64(1920),
					H:           ptrInt64(1080),
					MinDuration: 5,
					MaxDuration: 30,
					MIMEs:       []string{"video/mp4", "video/webm"},
				},
				BidFloor: floor,
			},
		},
		Site: &openrtb2.Site{
			ID:        siteID,
			Publisher: &openrtb2.Publisher{ID: pubID},
		},
		Device: &openrtb2.Device{
			IP:  ip,
			UA:  ua,
			IFA: deviceID,
		},
	}

	// Add PMP deal if specified
	if dealID != "" {
		bidRequest.Imp[0].PMP = &openrtb2.PMP{
			PrivateAuction: 1,
			Deals: []openrtb2.Deal{
				{ID: dealID},
			},
		}
	}

	// Process the request
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	bidResponse, err := h.ssp.processRequest(ctx, bidRequest)
	if err != nil || len(bidResponse.SeatBid) == 0 {
		// Return empty VAST
		c.Data(http.StatusOK, "application/xml", []byte(emptyVAST()))
		return
	}

	// Generate VAST from winning bid
	vast := h.generateVAST(bidResponse.SeatBid[0].Bid[0], contentID)
	c.Data(http.StatusOK, "application/xml", []byte(vast))
}

// convertToOpenRTB converts Publica SSAI request to OpenRTB format
func (h *PublicaHandler) convertToOpenRTB(req *PublicaSSAIRequest) *openrtb2.BidRequest {
	// Parse size if provided
	var width, height int64 = 1920, 1080 // Default to Full HD
	if size, ok := req.Params["size"].(string); ok {
		parts := strings.Split(size, "x")
		if len(parts) == 2 {
			width, _ = strconv.ParseInt(parts[0], 10, 64)
			height, _ = strconv.ParseInt(parts[1], 10, 64)
		}
	}

	bidRequest := &openrtb2.BidRequest{
		ID: uuid.New().String(),
		Imp: []openrtb2.Imp{
			{
				ID: "1",
				Video: &openrtb2.Video{
					W:           &width,
					H:           &height,
					MinDuration: 5,
					MaxDuration: 30,
					MIMEs:       []string{"video/mp4", "video/webm"},
				},
				BidFloor: req.FloorPrice,
			},
		},
		Site: &openrtb2.Site{
			ID:        req.SiteID,
			Publisher: &openrtb2.Publisher{ID: req.PublisherID},
			Content: &openrtb2.Content{
				ID:       req.ContentID,
				Genre:    req.ContentGenre,
				Language: req.ContentLanguage,
			},
		},
		Device: &openrtb2.Device{
			IP:  req.IP,
			UA:  req.UserAgent,
			IFA: req.DeviceID,
		},
	}

	// Add PMP deal if specified
	if req.DealID != "" {
		bidRequest.Imp[0].PMP = &openrtb2.PMP{
			PrivateAuction: 1,
			Deals: []openrtb2.Deal{
				{
					ID:       req.DealID,
					BidFloor: req.FloorPrice,
				},
			},
		}
	}

	return bidRequest
}

// convertToPublicaResponse converts OpenRTB response to Publica format
func (h *PublicaHandler) convertToPublicaResponse(resp *openrtb2.BidResponse, req *PublicaSSAIRequest) *PublicaSSAIResponse {
	baseURL := "https://ssp.ad.nexus" // This should come from config
	cacheBuster := strconv.FormatInt(time.Now().Unix(), 10)

	publicaResp := &PublicaSSAIResponse{
		AdBreakID:   uuid.New().String(),
		Ads:         []Ad{},
		CacheBuster: cacheBuster,
		TrackingURLs: Tracking{
			Impression: []string{},
			Click:      []string{},
			Complete:   []string{},
			Quartile1:  []string{},
			Quartile2:  []string{},
			Quartile3:  []string{},
		},
	}

	// Process winning bids
	for _, seatBid := range resp.SeatBid {
		for _, bid := range seatBid.Bid {
			ad := Ad{
				ID:         bid.ID,
				Duration:   30, // Default duration, should be parsed from VAST
				MediaURL:   bid.AdM,
				ClickURL:   fmt.Sprintf("%s/publica/click?bid=%s&cb=%s", baseURL, bid.ID, cacheBuster),
				Title:      "Video Ad",
				Advertiser: seatBid.Seat,
				CPM:        bid.Price,
			}

			publicaResp.Ads = append(publicaResp.Ads, ad)

			// Add tracking URLs
			publicaResp.TrackingURLs.Impression = append(publicaResp.TrackingURLs.Impression,
				fmt.Sprintf("%s/publica/pixel/impression?bid=%s&cb=%s", baseURL, bid.ID, cacheBuster))
			publicaResp.TrackingURLs.Complete = append(publicaResp.TrackingURLs.Complete,
				fmt.Sprintf("%s/publica/pixel/complete?bid=%s&cb=%s", baseURL, bid.ID, cacheBuster))
		}
	}

	// Set total duration
	totalDuration := 0
	for _, ad := range publicaResp.Ads {
		totalDuration += ad.Duration
	}
	publicaResp.Duration = totalDuration

	// Set VAST URL for fallback
	params := fmt.Sprintf("pub=%s&site=%s&content_id=%s&cb=%s",
		req.PublisherID, req.SiteID, req.ContentID, cacheBuster)
	if req.DealID != "" {
		params += "&deal=" + req.DealID
	}
	publicaResp.VASTURL = fmt.Sprintf("%s/publica/vast?%s", baseURL, params)

	return publicaResp
}

// generateVAST generates a VAST response from a bid
func (h *PublicaHandler) generateVAST(bid openrtb2.Bid, contentID string) string {
	// If the bid already contains VAST, return it
	if strings.Contains(bid.AdM, "<VAST") {
		return bid.AdM
	}

	// Otherwise generate VAST wrapper
	vast := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<VAST version="3.0">
  <Ad id="%s">
    <InLine>
      <AdSystem>AdNexus SSP</AdSystem>
      <AdTitle>Video Ad</AdTitle>
      <Impression><![CDATA[%s]]></Impression>
      <Creatives>
        <Creative>
          <Linear>
            <Duration>00:00:30</Duration>
            <MediaFiles>
              <MediaFile delivery="progressive" type="video/mp4" width="1920" height="1080">
                <![CDATA[%s]]>
              </MediaFile>
            </MediaFiles>
            <VideoClicks>
              <ClickThrough><![CDATA[https://ssp.ad.nexus/click?bid=%s]]></ClickThrough>
            </VideoClicks>
            <TrackingEvents>
              <Tracking event="start"><![CDATA[https://ssp.ad.nexus/pixel/start?bid=%s]]></Tracking>
              <Tracking event="firstQuartile"><![CDATA[https://ssp.ad.nexus/pixel/q1?bid=%s]]></Tracking>
              <Tracking event="midpoint"><![CDATA[https://ssp.ad.nexus/pixel/q2?bid=%s]]></Tracking>
              <Tracking event="thirdQuartile"><![CDATA[https://ssp.ad.nexus/pixel/q3?bid=%s]]></Tracking>
              <Tracking event="complete"><![CDATA[https://ssp.ad.nexus/pixel/complete?bid=%s]]></Tracking>
            </TrackingEvents>
          </Linear>
        </Creative>
      </Creatives>
    </InLine>
  </Ad>
</VAST>`, bid.ID, bid.NURL, bid.AdM, bid.ID, bid.ID, bid.ID, bid.ID, bid.ID, bid.ID)

	return vast
}

// emptyVAST returns an empty VAST response
func emptyVAST() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<VAST version="3.0">
  <Ad id="empty">
    <InLine>
      <AdSystem>AdNexus SSP</AdSystem>
      <AdTitle>No Ad Available</AdTitle>
      <Creatives></Creatives>
    </InLine>
  </Ad>
</VAST>`
}

// Helper functions
func ptrInt64(i int64) *int64 {
	return &i
}

func ptrInt8(i int8) *int8 {
	return &i
}