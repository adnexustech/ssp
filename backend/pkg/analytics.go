package ssp

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// AnalyticsStore handles SSP analytics storage in ClickHouse
type AnalyticsStore struct {
	conn clickhouse.Conn
}

// NewAnalyticsStore creates a new analytics store
func NewAnalyticsStore(addr string) (*AnalyticsStore, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: "default",
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	store := &AnalyticsStore{conn: conn}

	// Create tables
	if err := store.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return store, nil
}

// createTables creates analytics tables in ClickHouse
func (as *AnalyticsStore) createTables() error {
	ctx := context.Background()

	// SSP Ad Requests table
	adRequestsSchema := `
	CREATE TABLE IF NOT EXISTS ssp_ad_requests (
		request_id String,
		placement_id String,
		site_id String,
		publisher_id String,
		timestamp DateTime,
		url String,
		referer String,
		user_agent String,
		ip String,
		country String,
		device_type String,
		width UInt16,
		height UInt16,
		ad_type String,
		bid_floor Float64
	) ENGINE = MergeTree()
	ORDER BY (timestamp, publisher_id, site_id)
	PARTITION BY toYYYYMM(timestamp)
	TTL timestamp + INTERVAL 90 DAY;
	`

	if err := as.conn.Exec(ctx, adRequestsSchema); err != nil {
		return fmt.Errorf("failed to create ssp_ad_requests table: %w", err)
	}

	// SSP Bids table
	bidsSchema := `
	CREATE TABLE IF NOT EXISTS ssp_bids (
		bid_id String,
		request_id String,
		imp_id String,
		placement_id String,
		site_id String,
		publisher_id String,
		partner_id String,
		partner_name String,
		price Float64,
		currency String,
		adomain Array(String),
		timestamp DateTime,
		won UInt8,
		cleared_price Float64
	) ENGINE = MergeTree()
	ORDER BY (timestamp, publisher_id, partner_id)
	PARTITION BY toYYYYMM(timestamp)
	TTL timestamp + INTERVAL 90 DAY;
	`

	if err := as.conn.Exec(ctx, bidsSchema); err != nil {
		return fmt.Errorf("failed to create ssp_bids table: %w", err)
	}

	// SSP Impressions table
	impressionsSchema := `
	CREATE TABLE IF NOT EXISTS ssp_impressions (
		impression_id String,
		bid_id String,
		request_id String,
		placement_id String,
		site_id String,
		publisher_id String,
		partner_id String,
		price Float64,
		publisher_revenue Float64,
		timestamp DateTime,
		country String,
		device_type String
	) ENGINE = MergeTree()
	ORDER BY (timestamp, publisher_id, site_id)
	PARTITION BY toYYYYMM(timestamp)
	TTL timestamp + INTERVAL 90 DAY;
	`

	if err := as.conn.Exec(ctx, impressionsSchema); err != nil {
		return fmt.Errorf("failed to create ssp_impressions table: %w", err)
	}

	// SSP Clicks table
	clicksSchema := `
	CREATE TABLE IF NOT EXISTS ssp_clicks (
		click_id String,
		impression_id String,
		bid_id String,
		placement_id String,
		site_id String,
		publisher_id String,
		timestamp DateTime
	) ENGINE = MergeTree()
	ORDER BY (timestamp, publisher_id, site_id)
	PARTITION BY toYYYYMM(timestamp)
	TTL timestamp + INTERVAL 90 DAY;
	`

	if err := as.conn.Exec(ctx, clicksSchema); err != nil {
		return fmt.Errorf("failed to create ssp_clicks table: %w", err)
	}

	return nil
}

// AdRequestLog represents an ad request log entry
type AdRequestLog struct {
	RequestID   string
	PlacementID string
	SiteID      string
	PublisherID string
	Timestamp   time.Time
	URL         string
	Referer     string
	UserAgent   string
	IP          string
	Country     string
	DeviceType  string
	Width       int
	Height      int
	AdType      string
	BidFloor    float64
}

// LogAdRequest logs an ad request
func (as *AnalyticsStore) LogAdRequest(ctx context.Context, log *AdRequestLog) error {
	query := `
		INSERT INTO ssp_ad_requests (
			request_id, placement_id, site_id, publisher_id, timestamp,
			url, referer, user_agent, ip, country, device_type,
			width, height, ad_type, bid_floor
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return as.conn.Exec(ctx, query,
		log.RequestID,
		log.PlacementID,
		log.SiteID,
		log.PublisherID,
		log.Timestamp,
		log.URL,
		log.Referer,
		log.UserAgent,
		log.IP,
		log.Country,
		log.DeviceType,
		log.Width,
		log.Height,
		log.AdType,
		log.BidFloor,
	)
}

// BidLog represents a bid log entry
type BidLog struct {
	BidID        string
	RequestID    string
	ImpID        string
	PlacementID  string
	SiteID       string
	PublisherID  string
	PartnerID    string
	PartnerName  string
	Price        float64
	Currency     string
	ADomains     []string
	Timestamp    time.Time
	Won          bool
	ClearedPrice float64
}

// LogBid logs a bid
func (as *AnalyticsStore) LogBid(ctx context.Context, log *BidLog) error {
	won := uint8(0)
	if log.Won {
		won = 1
	}

	query := `
		INSERT INTO ssp_bids (
			bid_id, request_id, imp_id, placement_id, site_id, publisher_id,
			partner_id, partner_name, price, currency, adomain, timestamp,
			won, cleared_price
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return as.conn.Exec(ctx, query,
		log.BidID,
		log.RequestID,
		log.ImpID,
		log.PlacementID,
		log.SiteID,
		log.PublisherID,
		log.PartnerID,
		log.PartnerName,
		log.Price,
		log.Currency,
		log.ADomains,
		log.Timestamp,
		won,
		log.ClearedPrice,
	)
}

// ImpressionLog represents an impression log entry
type ImpressionLog struct {
	ImpressionID     string
	BidID            string
	RequestID        string
	PlacementID      string
	SiteID           string
	PublisherID      string
	PartnerID        string
	Price            float64
	PublisherRevenue float64
	Timestamp        time.Time
	Country          string
	DeviceType       string
}

// LogImpression logs an impression
func (as *AnalyticsStore) LogImpression(ctx context.Context, log *ImpressionLog) error {
	query := `
		INSERT INTO ssp_impressions (
			impression_id, bid_id, request_id, placement_id, site_id, publisher_id,
			partner_id, price, publisher_revenue, timestamp, country, device_type
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return as.conn.Exec(ctx, query,
		log.ImpressionID,
		log.BidID,
		log.RequestID,
		log.PlacementID,
		log.SiteID,
		log.PublisherID,
		log.PartnerID,
		log.Price,
		log.PublisherRevenue,
		log.Timestamp,
		log.Country,
		log.DeviceType,
	)
}

// ClickLog represents a click log entry
type ClickLog struct {
	ClickID      string
	ImpressionID string
	BidID        string
	PlacementID  string
	SiteID       string
	PublisherID  string
	Timestamp    time.Time
}

// LogClick logs a click
func (as *AnalyticsStore) LogClick(ctx context.Context, log *ClickLog) error {
	query := `
		INSERT INTO ssp_clicks (
			click_id, impression_id, bid_id, placement_id, site_id, publisher_id, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	return as.conn.Exec(ctx, query,
		log.ClickID,
		log.ImpressionID,
		log.BidID,
		log.PlacementID,
		log.SiteID,
		log.PublisherID,
		log.Timestamp,
	)
}

// GetPublisherStats retrieves publisher statistics
func (as *AnalyticsStore) GetPublisherStats(ctx context.Context, publisherID string, start, end time.Time) (*SupplyStats, error) {
	query := `
		SELECT
			publisher_id,
			'' as site_id,
			'' as placement_id,
			count(*) as requests,
			sum(CASE WHEN won = 1 THEN 1 ELSE 0 END) as impressions,
			sum(CASE WHEN won = 1 THEN cleared_price ELSE 0 END) as revenue,
			sum(CASE WHEN won = 1 THEN 1 ELSE 0 END) as fills,
			avg(CASE WHEN won = 1 THEN cleared_price ELSE 0 END) as avg_cpm,
			toDate(timestamp) as date
		FROM ssp_bids
		WHERE publisher_id = ?
			AND timestamp >= ?
			AND timestamp < ?
		GROUP BY publisher_id, date
	`

	rows, err := as.conn.Query(ctx, query, publisherID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := &SupplyStats{}
	for rows.Next() {
		var date string
		if err := rows.Scan(
			&stats.PublisherID,
			&stats.SiteID,
			&stats.PlacementID,
			&stats.Requests,
			&stats.Impressions,
			&stats.Revenue,
			&stats.Fills,
			&stats.AvgCPM,
			&date,
		); err != nil {
			return nil, err
		}
		stats.Date = date
	}

	return stats, nil
}

// GetSiteStats retrieves site statistics
func (as *AnalyticsStore) GetSiteStats(ctx context.Context, siteID string, start, end time.Time) ([]*SupplyStats, error) {
	query := `
		SELECT
			publisher_id,
			site_id,
			'' as placement_id,
			count(*) as requests,
			sum(CASE WHEN won = 1 THEN 1 ELSE 0 END) as impressions,
			sum(CASE WHEN won = 1 THEN cleared_price ELSE 0 END) as revenue,
			sum(CASE WHEN won = 1 THEN 1 ELSE 0 END) as fills,
			avg(CASE WHEN won = 1 THEN cleared_price ELSE 0 END) as avg_cpm,
			toDate(timestamp) as date
		FROM ssp_bids
		WHERE site_id = ?
			AND timestamp >= ?
			AND timestamp < ?
		GROUP BY publisher_id, site_id, date
		ORDER BY date DESC
	`

	rows, err := as.conn.Query(ctx, query, siteID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []*SupplyStats{}
	for rows.Next() {
		stat := &SupplyStats{}
		var date string
		if err := rows.Scan(
			&stat.PublisherID,
			&stat.SiteID,
			&stat.PlacementID,
			&stat.Requests,
			&stat.Impressions,
			&stat.Revenue,
			&stat.Fills,
			&stat.AvgCPM,
			&date,
		); err != nil {
			return nil, err
		}
		stat.Date = date
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetPlacementStats retrieves placement statistics
func (as *AnalyticsStore) GetPlacementStats(ctx context.Context, placementID string, start, end time.Time) ([]*SupplyStats, error) {
	query := `
		SELECT
			publisher_id,
			site_id,
			placement_id,
			count(*) as requests,
			sum(CASE WHEN won = 1 THEN 1 ELSE 0 END) as impressions,
			sum(CASE WHEN won = 1 THEN cleared_price ELSE 0 END) as revenue,
			sum(CASE WHEN won = 1 THEN 1 ELSE 0 END) as fills,
			avg(CASE WHEN won = 1 THEN cleared_price ELSE 0 END) as avg_cpm,
			toDate(timestamp) as date
		FROM ssp_bids
		WHERE placement_id = ?
			AND timestamp >= ?
			AND timestamp < ?
		GROUP BY publisher_id, site_id, placement_id, date
		ORDER BY date DESC
	`

	rows, err := as.conn.Query(ctx, query, placementID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []*SupplyStats{}
	for rows.Next() {
		stat := &SupplyStats{}
		var date string
		if err := rows.Scan(
			&stat.PublisherID,
			&stat.SiteID,
			&stat.PlacementID,
			&stat.Requests,
			&stat.Impressions,
			&stat.Revenue,
			&stat.Fills,
			&stat.AvgCPM,
			&date,
		); err != nil {
			return nil, err
		}
		stat.Date = date
		stats = append(stats, stat)
	}

	return stats, nil
}

// Close closes the ClickHouse connection
func (as *AnalyticsStore) Close() error {
	return as.conn.Close()
}
