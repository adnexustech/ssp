package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hanzo-labs/adnexus/pkg/ssp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type SSPService struct {
	ssp            *ssp.SSP
	store          *ssp.PostgresStore
	analyticsStore *ssp.AnalyticsStore
	bidder         *ssp.Bidder
	bidReqBuilder  *ssp.BidRequestBuilder
	auctionEngine  *ssp.AuctionEngine
	tagGenerator   *ssp.TagGenerator
	partnerManager *ssp.PartnerManager
	logger         *slog.Logger

	// Prometheus Metrics
	adRequestsTotal  prometheus.Counter
	auctionTotal     prometheus.Counter
	impressionsTotal prometheus.Counter
	auctionLatency   prometheus.Histogram
	publisherRevenue prometheus.Counter
}

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Initialize Prometheus metrics
	adRequestsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ssp_ad_requests_total",
		Help: "Total number of ad requests received",
	})
	auctionTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ssp_auctions_total",
		Help: "Total number of auctions conducted",
	})
	impressionsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ssp_impressions_total",
		Help: "Total number of impressions served",
	})
	auctionLatency := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "ssp_auction_latency_seconds",
		Help:    "Auction processing latency in seconds",
		Buckets: prometheus.DefBuckets,
	})
	publisherRevenue := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ssp_publisher_revenue_total",
		Help: "Total publisher revenue",
	})

	prometheus.MustRegister(adRequestsTotal, auctionTotal, impressionsTotal, auctionLatency, publisherRevenue)

	// Configuration from environment
	dbConnString := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/adnexus?sslmode=disable")
	clickhouseAddr := getEnv("CLICKHOUSE_ADDR", "localhost:9000")
	clickhouseEnabled := strings.ToLower(getEnv("CLICKHOUSE_ENABLED", "true")) == "true"
	sspID := getEnv("SSP_ID", "adnexus-ssp")
	sspEndpoint := getEnv("SSP_ENDPOINT", "http://localhost:8081")
	cdnURL := getEnv("CDN_URL", "https://cdn.adnexus.io")
	port := getEnv("PORT", "8081")
	metricsPort := getEnv("METRICS_PORT", "6061")
	exadsEndpoint := getEnv("EXADS_ENDPOINT", "")
	exadsAPIKey := getEnv("EXADS_API_KEY", "")

	// Initialize stores
	logger.Info("Initializing PostgreSQL store")
	postgresStore, err := ssp.NewPostgresStore(dbConnString)
	if err != nil {
		logger.Error("Failed to initialize PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer postgresStore.Close()

	var analyticsStore *ssp.AnalyticsStore
	if clickhouseEnabled {
		logger.Info("Initializing ClickHouse analytics")
		analyticsStore, err = ssp.NewAnalyticsStore(clickhouseAddr)
		if err != nil {
			logger.Warn("Failed to initialize ClickHouse, continuing without analytics", "error", err)
			analyticsStore = nil
		} else {
			defer analyticsStore.Close()
		}
	} else {
		logger.Info("ClickHouse disabled, skipping analytics initialization")
	}

	// Initialize components
	bidder := ssp.NewBidder(sspID, 120*time.Millisecond)
	bidReqBuilder := ssp.NewBidRequestBuilder(sspID)
	auctionEngine := ssp.NewAuctionEngine(0.01) // $0.01 minimum bid floor
	tagGenerator := ssp.NewTagGenerator(sspEndpoint, cdnURL)
	partnerManager := ssp.NewPartnerManager()

	// Add BidsCube as demand partner
	adnexusEndpoint := getEnv("ADNEXUS_ENDPOINT", "")
	adnexusAPIKey := getEnv("ADNEXUS_API_KEY", "")
	adnexusEnabled := strings.ToLower(getEnv("ADNEXUS_ENABLED", "false")) == "true"
	
	if adnexusEnabled && adnexusEndpoint != "" {
		logger.Info("Configuring BidsCube partner", "endpoint", adnexusEndpoint)
		adnexusRevShare, _ := strconv.ParseFloat(getEnv("ADNEXUS_REV_SHARE", "0.30"), 64)
		
		partnerManager.AddPartner(&ssp.SupplyPartner{
			ID:       "adnexus-wl",
			Name:     "BidsCube WL",
			Type:     "adnexus",
			Endpoint: adnexusEndpoint,
			APIKey:   adnexusAPIKey,
			Timeout:  100 * time.Millisecond,
			Active:   true,
			QPS:      1000,
			RevShare: adnexusRevShare,
		})
	}
	
	// Add DSP as primary demand partner
	dspEndpoint := getEnv("DSP_ENDPOINT", "https://dsp.ad.nexus/openrtb2/auction")
	if dspEndpoint != "" {
		logger.Info("Configuring DSP partner", "endpoint", dspEndpoint)
		partnerManager.AddPartner(&ssp.SupplyPartner{
			ID:       "adnexus-dsp",
			Name:     "AdNexus DSP",
			Type:     "dsp",
			Endpoint: dspEndpoint,
			Timeout:  100 * time.Millisecond,
			Active:   true,
			QPS:      5000,
			RevShare: 0.20, // SSP keeps 20%, publisher gets 80%
		})
	}
	
	// Legacy EXADS support (optional)
	if exadsEndpoint != "" {
		logger.Info("Configuring EXADS partner", "endpoint", exadsEndpoint)
		partnerManager.AddPartner(&ssp.SupplyPartner{
			ID:       "exads-1",
			Name:     "EXADS",
			Type:     "exads",
			Endpoint: exadsEndpoint,
			APIKey:   exadsAPIKey,
			Timeout:  100 * time.Millisecond,
			Active:   true,
			QPS:      1000,
			RevShare: 0.30, // SSP keeps 30%, publisher gets 70%
		})
	}

	// Create SSP instance
	sspInstance := ssp.NewSSP(partnerManager, auctionEngine, bidder, logger)

	// Create service
	service := &SSPService{
		ssp:              sspInstance,
		store:            postgresStore,
		analyticsStore:   analyticsStore,
		bidder:           bidder,
		bidReqBuilder:    bidReqBuilder,
		auctionEngine:    auctionEngine,
		tagGenerator:     tagGenerator,
		partnerManager:   partnerManager,
		logger:           logger,
		adRequestsTotal:  adRequestsTotal,
		auctionTotal:     auctionTotal,
		impressionsTotal: impressionsTotal,
		auctionLatency:   auctionLatency,
		publisherRevenue: publisherRevenue,
	}

	// Setup router
	router := setupRouter(service)

	// Start metrics server
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		logger.Info("Starting metrics server", "port", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, mux); err != nil {
			logger.Error("Metrics server failed", "error", err)
		}
	}()

	// Start server
	logger.Info("Starting SSP server", "port", port, "sspID", sspID)
	if err := router.Run(":" + port); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

func setupRouter(service *SSPService) *gin.Engine {
	router := gin.Default()

	// Health check - support both GET and HEAD
	healthHandler := func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "ssp"})
	}
	router.GET("/health", healthHandler)
	router.HEAD("/health", healthHandler)

	// P1 Publisher onboarding and management
	api := router.Group("/api")
	{
		// Publisher management
		api.POST("/publishers", service.handleCreatePublisher)
		api.GET("/publishers", service.handleListPublishers)
		api.GET("/publishers/:id", service.handleGetPublisher)
		api.PUT("/publishers/:id", service.handleUpdatePublisher)
		api.DELETE("/publishers/:id", service.handleDeletePublisher)

		// Site management
		api.POST("/sites", service.handleCreateSite)
		api.GET("/sites", service.handleListSites)
		api.GET("/sites/:id", service.handleGetSite)
		api.PUT("/sites/:id", service.handleUpdateSite)
		api.DELETE("/sites/:id", service.handleDeleteSite)

		// Placement management
		api.POST("/placements", service.handleCreatePlacement)
		api.GET("/placements", service.handleListPlacements)
		api.GET("/placements/:id", service.handleGetPlacement)
		api.PUT("/placements/:id", service.handleUpdatePlacement)
		api.DELETE("/placements/:id", service.handleDeletePlacement)

		// Tag generation
		api.GET("/tags/display/:placement_id", service.handleGenerateDisplayTag)
		api.GET("/tags/vast/:placement_id", service.handleGenerateVASTTag)
		api.GET("/tags/header-bidding/:placement_id", service.handleGenerateHeaderBiddingTag)

		// Analytics
		api.GET("/stats/publisher/:id", service.handleGetPublisherStats)
		api.GET("/stats/site/:id", service.handleGetSiteStats)
		api.GET("/stats/placement/:id", service.handleGetPlacementStats)
	}

	// Ad serving endpoints
	router.GET("/ad/request", service.handleAdRequest)
	router.POST("/ad/request", service.handleAdRequest)

	// VAST endpoint for video ads
	router.GET("/vast/:placement_id", service.handleVASTRequest)

	// OpenRTB 2.5 endpoint (receive from internal ADX)
	router.POST("/openrtb2/auction", service.handleOpenRTBAuction)

	// Impression tracking
	router.GET("/impression/:bid_id", service.handleImpressionTracking)

	// Click tracking
	router.GET("/click/:bid_id", service.handleClickTracking)

	// Publica SSAI endpoints for P1
	publicaHandler := ssp.NewPublicaHandler(service.ssp)
	publica := router.Group("/publica")
	{
		// Server-Side Ad Insertion endpoint
		publica.POST("/ssai", publicaHandler.HandleSSAI)
		// VAST endpoint with Publica macros
		publica.GET("/vast", publicaHandler.HandleVAST)
		// Tracking endpoints
		publica.GET("/pixel/impression", service.handleImpressionTracking)
		publica.GET("/pixel/complete", service.handleImpressionTracking)
		publica.GET("/pixel/start", service.handleImpressionTracking)
		publica.GET("/pixel/q1", service.handleImpressionTracking)
		publica.GET("/pixel/q2", service.handleImpressionTracking)
		publica.GET("/pixel/q3", service.handleImpressionTracking)
		publica.GET("/click", service.handleClickTracking)
	}

	return router
}

// Publisher onboarding handlers

func (s *SSPService) handleCreatePublisher(c *gin.Context) {
	var pub ssp.Publisher
	if err := c.ShouldBindJSON(&pub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid publisher data"})
		return
	}

	// Generate ID if not provided
	if pub.ID == "" {
		pub.ID = uuid.New().String()
	}

	pub.CreatedAt = time.Now()
	pub.UpdatedAt = time.Now()

	// Default revenue share: 70% to publisher, 30% to SSP
	if pub.RevShare == 0 {
		pub.RevShare = 0.70
	}

	if err := s.store.CreatePublisher(c.Request.Context(), &pub); err != nil {
		s.logger.Error("Failed to create publisher", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("Publisher created", "id", pub.ID, "name", pub.Name)
	c.JSON(http.StatusCreated, pub)
}

func (s *SSPService) handleListPublishers(c *gin.Context) {
	activeOnly := c.Query("active") == "true"

	publishers, err := s.store.ListPublishers(c.Request.Context(), activeOnly)
	if err != nil {
		s.logger.Error("Failed to list publishers", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, publishers)
}

func (s *SSPService) handleGetPublisher(c *gin.Context) {
	id := c.Param("id")

	pub, err := s.store.GetPublisher(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "publisher not found"})
			return
		}
		s.logger.Error("Failed to get publisher", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pub)
}

func (s *SSPService) handleUpdatePublisher(c *gin.Context) {
	id := c.Param("id")

	var pub ssp.Publisher
	if err := c.ShouldBindJSON(&pub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid publisher data"})
		return
	}

	pub.ID = id
	pub.UpdatedAt = time.Now()

	if err := s.store.UpdatePublisher(c.Request.Context(), &pub); err != nil {
		s.logger.Error("Failed to update publisher", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pub)
}

func (s *SSPService) handleDeletePublisher(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.DeletePublisher(c.Request.Context(), id); err != nil {
		s.logger.Error("Failed to delete publisher", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// Site handlers

func (s *SSPService) handleCreateSite(c *gin.Context) {
	var site ssp.Site
	if err := c.ShouldBindJSON(&site); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site data"})
		return
	}

	if site.ID == "" {
		site.ID = uuid.New().String()
	}

	site.CreatedAt = time.Now()
	site.UpdatedAt = time.Now()

	if err := s.store.CreateSite(c.Request.Context(), &site); err != nil {
		s.logger.Error("Failed to create site", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("Site created", "id", site.ID, "domain", site.Domain)
	c.JSON(http.StatusCreated, site)
}

func (s *SSPService) handleListSites(c *gin.Context) {
	publisherID := c.Query("publisher_id")
	if publisherID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "publisher_id required"})
		return
	}

	activeOnly := c.Query("active") == "true"

	sites, err := s.store.ListSites(c.Request.Context(), publisherID, activeOnly)
	if err != nil {
		s.logger.Error("Failed to list sites", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sites)
}

func (s *SSPService) handleGetSite(c *gin.Context) {
	id := c.Param("id")

	site, err := s.store.GetSite(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
			return
		}
		s.logger.Error("Failed to get site", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, site)
}

func (s *SSPService) handleUpdateSite(c *gin.Context) {
	id := c.Param("id")

	var site ssp.Site
	if err := c.ShouldBindJSON(&site); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site data"})
		return
	}

	site.ID = id
	site.UpdatedAt = time.Now()

	if err := s.store.UpdateSite(c.Request.Context(), &site); err != nil {
		s.logger.Error("Failed to update site", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, site)
}

func (s *SSPService) handleDeleteSite(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.DeleteSite(c.Request.Context(), id); err != nil {
		s.logger.Error("Failed to delete site", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// Placement handlers

func (s *SSPService) handleCreatePlacement(c *gin.Context) {
	var placement ssp.Placement
	if err := c.ShouldBindJSON(&placement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid placement data"})
		return
	}

	if placement.ID == "" {
		placement.ID = uuid.New().String()
	}

	placement.CreatedAt = time.Now()
	placement.UpdatedAt = time.Now()

	if err := s.store.CreatePlacement(c.Request.Context(), &placement); err != nil {
		s.logger.Error("Failed to create placement", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("Placement created", "id", placement.ID, "ad_type", placement.AdType)
	c.JSON(http.StatusCreated, placement)
}

func (s *SSPService) handleListPlacements(c *gin.Context) {
	siteID := c.Query("site_id")
	if siteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "site_id required"})
		return
	}

	activeOnly := c.Query("active") == "true"

	placements, err := s.store.ListPlacements(c.Request.Context(), siteID, activeOnly)
	if err != nil {
		s.logger.Error("Failed to list placements", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, placements)
}

func (s *SSPService) handleGetPlacement(c *gin.Context) {
	id := c.Param("id")

	placement, err := s.store.GetPlacement(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "placement not found"})
			return
		}
		s.logger.Error("Failed to get placement", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, placement)
}

func (s *SSPService) handleUpdatePlacement(c *gin.Context) {
	id := c.Param("id")

	var placement ssp.Placement
	if err := c.ShouldBindJSON(&placement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid placement data"})
		return
	}

	placement.ID = id
	placement.UpdatedAt = time.Now()

	if err := s.store.UpdatePlacement(c.Request.Context(), &placement); err != nil {
		s.logger.Error("Failed to update placement", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, placement)
}

func (s *SSPService) handleDeletePlacement(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.DeletePlacement(c.Request.Context(), id); err != nil {
		s.logger.Error("Failed to delete placement", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// Tag generation handlers

func (s *SSPService) handleGenerateDisplayTag(c *gin.Context) {
	placementID := c.Param("placement_id")

	placement, err := s.store.GetPlacement(c.Request.Context(), placementID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "placement not found"})
		return
	}

	tag, err := s.tagGenerator.GenerateDisplayTag(placement)
	if err != nil {
		s.logger.Error("Failed to generate display tag", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, tag)
}

func (s *SSPService) handleGenerateVASTTag(c *gin.Context) {
	placementID := c.Param("placement_id")

	placement, err := s.store.GetPlacement(c.Request.Context(), placementID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "placement not found"})
		return
	}

	tag, err := s.tagGenerator.GenerateVASTTag(placement)
	if err != nil {
		s.logger.Error("Failed to generate VAST tag", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, tag)
}

func (s *SSPService) handleGenerateHeaderBiddingTag(c *gin.Context) {
	placementID := c.Param("placement_id")

	placement, err := s.store.GetPlacement(c.Request.Context(), placementID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "placement not found"})
		return
	}

	tag, err := s.tagGenerator.GenerateHeaderBiddingTag(placement)
	if err != nil {
		s.logger.Error("Failed to generate header bidding tag", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, tag)
}

// Ad request handler

func (s *SSPService) handleAdRequest(c *gin.Context) {
	start := time.Now()
	s.adRequestsTotal.Inc()

	placementID := c.Query("placement_id")
	if placementID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "placement_id required"})
		return
	}

	// Load placement, site, and publisher
	placement, err := s.store.GetPlacement(c.Request.Context(), placementID)
	if err != nil {
		s.logger.Error("Placement not found", "placement_id", placementID)
		c.Status(http.StatusNoContent)
		return
	}

	site, err := s.store.GetSite(c.Request.Context(), placement.SiteID)
	if err != nil {
		s.logger.Error("Site not found", "site_id", placement.SiteID)
		c.Status(http.StatusNoContent)
		return
	}

	publisher, err := s.store.GetPublisher(c.Request.Context(), site.PublisherID)
	if err != nil {
		s.logger.Error("Publisher not found", "publisher_id", site.PublisherID)
		c.Status(http.StatusNoContent)
		return
	}

	// Build ad request
	adReq := &ssp.AdRequest{
		PlacementID: placementID,
		URL:         c.Request.Referer(),
		Referer:     c.Request.Header.Get("Referer"),
		UserAgent:   c.Request.UserAgent(),
		IP:          c.ClientIP(),
		Width:       placement.Width,
		Height:      placement.Height,
	}

	// Build OpenRTB bid request
	bidReq, err := s.bidReqBuilder.BuildBidRequest(adReq, placement, site, publisher)
	if err != nil {
		s.logger.Error("Failed to build bid request", "error", err)
		c.Status(http.StatusNoContent)
		return
	}

	// Log ad request
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		logEntry := &ssp.AdRequestLog{
			RequestID:   bidReq.ID,
			PlacementID: placementID,
			SiteID:      site.ID,
			PublisherID: publisher.ID,
			Timestamp:   time.Now(),
			URL:         adReq.URL,
			Referer:     adReq.Referer,
			UserAgent:   adReq.UserAgent,
			IP:          adReq.IP,
			Width:       placement.Width,
			Height:      placement.Height,
			AdType:      placement.AdType,
			BidFloor:    placement.MinBidFloor,
		}
		if err := s.analyticsStore.LogAdRequest(ctx, logEntry); err != nil {
			s.logger.Error("Failed to log ad request", "error", err)
		}
	}()

	// Send bid requests to partners
	responses := make(map[*ssp.DemandPartner]*ssp.BidResponse)
	partners := s.partnerManager.GetActivePartners()

	for _, partner := range partners {
		// Convert SupplyPartner to DemandPartner for compatibility
		dp := &ssp.DemandPartner{
			ID:       partner.ID,
			Name:     partner.Name,
			Endpoint: partner.Endpoint,
			Timeout:  partner.Timeout,
			Active:   partner.Active,
			QPS:      partner.QPS,
			RevShare: partner.RevShare,
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), partner.Timeout)
		resp, err := s.bidder.SendBidRequest(ctx, bidReq, dp)
		cancel()

		if err != nil {
			s.logger.Error("Partner bid request failed", "partner", partner.Name, "error", err)
			continue
		}

		if resp != nil {
			responses[dp] = resp
		}
	}

	// Run auction
	s.auctionTotal.Inc()
	result, err := s.auctionEngine.RunAuction(responses, placement)
	if err != nil || result == nil {
		s.logger.Debug("No winning bid", "request_id", bidReq.ID)
		c.Status(http.StatusNoContent)
		return
	}

	// Calculate publisher revenue (70% default)
	publisherRevenue := result.ClearedPrice * publisher.RevShare

	// Update metrics
	s.publisherRevenue.Add(publisherRevenue)
	duration := time.Since(start)
	s.auctionLatency.Observe(duration.Seconds())

	s.logger.Info("Auction complete",
		"request_id", bidReq.ID,
		"winner", result.WinningPartner.Name,
		"price", result.ClearedPrice,
		"pub_revenue", publisherRevenue,
		"duration_ms", duration.Milliseconds(),
	)

	// Return ad markup
	c.JSON(http.StatusOK, gin.H{
		"ad":      result.WinningBid.ADM,
		"bid_id":  result.WinningBid.ID,
		"price":   result.ClearedPrice,
		"adomain": result.WinningBid.ADomain,
	})
}

// OpenRTB auction handler (from internal ADX)

func (s *SSPService) handleOpenRTBAuction(c *gin.Context) {
	var bidReq ssp.BidRequest
	if err := c.ShouldBindJSON(&bidReq); err != nil {
		s.logger.Error("Invalid bid request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bid request"})
		return
	}

	// Process similar to handleAdRequest but return OpenRTB response
	// This would integrate with internal ADX

	c.Status(http.StatusNoContent) // Placeholder
}

// VAST request handler

func (s *SSPService) handleVASTRequest(c *gin.Context) {
	placementID := c.Param("placement_id")

	// Similar to handleAdRequest but returns VAST XML
	_, err := s.store.GetPlacement(c.Request.Context(), placementID)
	if err != nil {
		c.XML(http.StatusNotFound, gin.H{"error": "placement not found"})
		return
	}

	// Run auction...
	// Return VAST XML with winning creative

	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, "<?xml version=\"1.0\" encoding=\"UTF-8\"?><VAST version=\"3.0\"></VAST>")
}

// Impression tracking

func (s *SSPService) handleImpressionTracking(c *gin.Context) {
	bidID := c.Param("bid_id")
	s.impressionsTotal.Inc()

	// Log impression
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		log := &ssp.ImpressionLog{
			ImpressionID: uuid.New().String(),
			BidID:        bidID,
			Timestamp:    time.Now(),
		}

		if err := s.analyticsStore.LogImpression(ctx, log); err != nil {
			s.logger.Error("Failed to log impression", "error", err)
		}
	}()

	// Return 1x1 transparent pixel
	c.Data(http.StatusOK, "image/gif", []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00,
		0x01, 0x00, 0x80, 0x00, 0x00, 0xFF, 0xFF, 0xFF,
		0x00, 0x00, 0x00, 0x21, 0xF9, 0x04, 0x01, 0x00,
		0x00, 0x00, 0x00, 0x2C, 0x00, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
		0x01, 0x00, 0x3B,
	})
}

// Click tracking

func (s *SSPService) handleClickTracking(c *gin.Context) {
	bidID := c.Param("bid_id")

	// Log click
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		log := &ssp.ClickLog{
			ClickID:   uuid.New().String(),
			BidID:     bidID,
			Timestamp: time.Now(),
		}

		if err := s.analyticsStore.LogClick(ctx, log); err != nil {
			s.logger.Error("Failed to log click", "error", err)
		}
	}()

	c.Status(http.StatusOK)
}

// Analytics handlers

func (s *SSPService) handleGetPublisherStats(c *gin.Context) {
	id := c.Param("id")

	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()

	if startStr := c.Query("start"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = parsed
		}
	}

	if endStr := c.Query("end"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = parsed
		}
	}

	stats, err := s.analyticsStore.GetPublisherStats(c.Request.Context(), id, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to get publisher stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (s *SSPService) handleGetSiteStats(c *gin.Context) {
	id := c.Param("id")

	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()

	if startStr := c.Query("start"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = parsed
		}
	}

	if endStr := c.Query("end"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = parsed
		}
	}

	stats, err := s.analyticsStore.GetSiteStats(c.Request.Context(), id, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to get site stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (s *SSPService) handleGetPlacementStats(c *gin.Context) {
	id := c.Param("id")

	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()

	if startStr := c.Query("start"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = parsed
		}
	}

	if endStr := c.Query("end"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = parsed
		}
	}

	stats, err := s.analyticsStore.GetPlacementStats(c.Request.Context(), id, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to get placement stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
