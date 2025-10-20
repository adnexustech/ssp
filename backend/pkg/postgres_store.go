package ssp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
)

// PostgresStore implements SSP storage using PostgreSQL
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new PostgreSQL store
func NewPostgresStore(connString string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PostgresStore{db: db}

	// Create tables if they don't exist
	if err := store.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return store, nil
}

// createTables creates the necessary database tables
func (ps *PostgresStore) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS publishers (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL,
		domain VARCHAR(255) NOT NULL,
		active BOOLEAN DEFAULT true,
		rev_share DECIMAL(3, 2) DEFAULT 0.70,
		payment_info TEXT,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS sites (
		id VARCHAR(255) PRIMARY KEY,
		publisher_id VARCHAR(255) NOT NULL REFERENCES publishers(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		domain VARCHAR(255) NOT NULL,
		page VARCHAR(500),
		cat JSONB,
		active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS placements (
		id VARCHAR(255) PRIMARY KEY,
		site_id VARCHAR(255) NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		ad_type VARCHAR(50) NOT NULL,
		width INT,
		height INT,
		min_bid_floor DECIMAL(10, 4) DEFAULT 0,
		active BOOLEAN DEFAULT true,
		formats JSONB,
		video JSONB,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_publishers_active ON publishers(active);
	CREATE INDEX IF NOT EXISTS idx_publishers_email ON publishers(email);
	CREATE INDEX IF NOT EXISTS idx_sites_publisher_id ON sites(publisher_id);
	CREATE INDEX IF NOT EXISTS idx_sites_active ON sites(active);
	CREATE INDEX IF NOT EXISTS idx_placements_site_id ON placements(site_id);
	CREATE INDEX IF NOT EXISTS idx_placements_active ON placements(active);
	`

	_, err := ps.db.Exec(schema)
	return err
}

// Publisher operations

// CreatePublisher creates a new publisher
func (ps *PostgresStore) CreatePublisher(ctx context.Context, pub *Publisher) error {
	query := `
		INSERT INTO publishers (id, name, email, domain, active, rev_share, payment_info, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := ps.db.ExecContext(ctx, query,
		pub.ID,
		pub.Name,
		pub.Email,
		pub.Domain,
		pub.Active,
		pub.RevShare,
		pub.PaymentInfo,
		pub.CreatedAt,
		pub.UpdatedAt,
	)

	return err
}

// GetPublisher retrieves a publisher by ID
func (ps *PostgresStore) GetPublisher(ctx context.Context, id string) (*Publisher, error) {
	query := `
		SELECT id, name, email, domain, active, rev_share, payment_info, created_at, updated_at
		FROM publishers
		WHERE id = $1
	`

	pub := &Publisher{}
	var paymentInfo sql.NullString

	err := ps.db.QueryRowContext(ctx, query, id).Scan(
		&pub.ID,
		&pub.Name,
		&pub.Email,
		&pub.Domain,
		&pub.Active,
		&pub.RevShare,
		&paymentInfo,
		&pub.CreatedAt,
		&pub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("publisher not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	if paymentInfo.Valid {
		pub.PaymentInfo = paymentInfo.String
	}

	return pub, nil
}

// ListPublishers lists publishers
func (ps *PostgresStore) ListPublishers(ctx context.Context, activeOnly bool) ([]*Publisher, error) {
	query := `
		SELECT id, name, email, domain, active, rev_share, payment_info, created_at, updated_at
		FROM publishers
	`

	if activeOnly {
		query += " WHERE active = true"
	}

	query += " ORDER BY created_at DESC"

	rows, err := ps.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	publishers := []*Publisher{}

	for rows.Next() {
		pub := &Publisher{}
		var paymentInfo sql.NullString

		err := rows.Scan(
			&pub.ID,
			&pub.Name,
			&pub.Email,
			&pub.Domain,
			&pub.Active,
			&pub.RevShare,
			&paymentInfo,
			&pub.CreatedAt,
			&pub.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if paymentInfo.Valid {
			pub.PaymentInfo = paymentInfo.String
		}

		publishers = append(publishers, pub)
	}

	return publishers, nil
}

// UpdatePublisher updates a publisher
func (ps *PostgresStore) UpdatePublisher(ctx context.Context, pub *Publisher) error {
	query := `
		UPDATE publishers
		SET name = $2, email = $3, domain = $4, active = $5, rev_share = $6, payment_info = $7, updated_at = $8
		WHERE id = $1
	`

	_, err := ps.db.ExecContext(ctx, query,
		pub.ID,
		pub.Name,
		pub.Email,
		pub.Domain,
		pub.Active,
		pub.RevShare,
		pub.PaymentInfo,
		pub.UpdatedAt,
	)

	return err
}

// DeletePublisher deletes a publisher
func (ps *PostgresStore) DeletePublisher(ctx context.Context, id string) error {
	query := "DELETE FROM publishers WHERE id = $1"
	_, err := ps.db.ExecContext(ctx, query, id)
	return err
}

// Site operations

// CreateSite creates a new site
func (ps *PostgresStore) CreateSite(ctx context.Context, site *Site) error {
	catJSON, err := json.Marshal(site.Cat)
	if err != nil {
		return fmt.Errorf("failed to marshal categories: %w", err)
	}

	query := `
		INSERT INTO sites (id, publisher_id, name, domain, page, cat, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = ps.db.ExecContext(ctx, query,
		site.ID,
		site.PublisherID,
		site.Name,
		site.Domain,
		site.Page,
		catJSON,
		site.Active,
		site.CreatedAt,
		site.UpdatedAt,
	)

	return err
}

// GetSite retrieves a site by ID
func (ps *PostgresStore) GetSite(ctx context.Context, id string) (*Site, error) {
	query := `
		SELECT id, publisher_id, name, domain, page, cat, active, created_at, updated_at
		FROM sites
		WHERE id = $1
	`

	site := &Site{}
	var page sql.NullString
	var catJSON []byte

	err := ps.db.QueryRowContext(ctx, query, id).Scan(
		&site.ID,
		&site.PublisherID,
		&site.Name,
		&site.Domain,
		&page,
		&catJSON,
		&site.Active,
		&site.CreatedAt,
		&site.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("site not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	if page.Valid {
		site.Page = page.String
	}

	if len(catJSON) > 0 {
		if err := json.Unmarshal(catJSON, &site.Cat); err != nil {
			return nil, fmt.Errorf("failed to unmarshal categories: %w", err)
		}
	}

	return site, nil
}

// ListSites lists sites for a publisher
func (ps *PostgresStore) ListSites(ctx context.Context, publisherID string, activeOnly bool) ([]*Site, error) {
	query := `
		SELECT id, publisher_id, name, domain, page, cat, active, created_at, updated_at
		FROM sites
		WHERE publisher_id = $1
	`

	if activeOnly {
		query += " AND active = true"
	}

	query += " ORDER BY created_at DESC"

	rows, err := ps.db.QueryContext(ctx, query, publisherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sites := []*Site{}

	for rows.Next() {
		site := &Site{}
		var page sql.NullString
		var catJSON []byte

		err := rows.Scan(
			&site.ID,
			&site.PublisherID,
			&site.Name,
			&site.Domain,
			&page,
			&catJSON,
			&site.Active,
			&site.CreatedAt,
			&site.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if page.Valid {
			site.Page = page.String
		}

		if len(catJSON) > 0 {
			if err := json.Unmarshal(catJSON, &site.Cat); err != nil {
				return nil, fmt.Errorf("failed to unmarshal categories: %w", err)
			}
		}

		sites = append(sites, site)
	}

	return sites, nil
}

// UpdateSite updates a site
func (ps *PostgresStore) UpdateSite(ctx context.Context, site *Site) error {
	catJSON, err := json.Marshal(site.Cat)
	if err != nil {
		return fmt.Errorf("failed to marshal categories: %w", err)
	}

	query := `
		UPDATE sites
		SET name = $2, domain = $3, page = $4, cat = $5, active = $6, updated_at = $7
		WHERE id = $1
	`

	_, err = ps.db.ExecContext(ctx, query,
		site.ID,
		site.Name,
		site.Domain,
		site.Page,
		catJSON,
		site.Active,
		site.UpdatedAt,
	)

	return err
}

// DeleteSite deletes a site
func (ps *PostgresStore) DeleteSite(ctx context.Context, id string) error {
	query := "DELETE FROM sites WHERE id = $1"
	_, err := ps.db.ExecContext(ctx, query, id)
	return err
}

// Placement operations

// CreatePlacement creates a new placement
func (ps *PostgresStore) CreatePlacement(ctx context.Context, placement *Placement) error {
	formatsJSON, err := json.Marshal(placement.Formats)
	if err != nil {
		return fmt.Errorf("failed to marshal formats: %w", err)
	}

	videoJSON, err := json.Marshal(placement.Video)
	if err != nil {
		return fmt.Errorf("failed to marshal video settings: %w", err)
	}

	query := `
		INSERT INTO placements (id, site_id, name, ad_type, width, height, min_bid_floor, active, formats, video, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = ps.db.ExecContext(ctx, query,
		placement.ID,
		placement.SiteID,
		placement.Name,
		placement.AdType,
		placement.Width,
		placement.Height,
		placement.MinBidFloor,
		placement.Active,
		formatsJSON,
		videoJSON,
		placement.CreatedAt,
		placement.UpdatedAt,
	)

	return err
}

// GetPlacement retrieves a placement by ID
func (ps *PostgresStore) GetPlacement(ctx context.Context, id string) (*Placement, error) {
	query := `
		SELECT id, site_id, name, ad_type, width, height, min_bid_floor, active, formats, video, created_at, updated_at
		FROM placements
		WHERE id = $1
	`

	placement := &Placement{}
	var width, height sql.NullInt32
	var formatsJSON, videoJSON []byte

	err := ps.db.QueryRowContext(ctx, query, id).Scan(
		&placement.ID,
		&placement.SiteID,
		&placement.Name,
		&placement.AdType,
		&width,
		&height,
		&placement.MinBidFloor,
		&placement.Active,
		&formatsJSON,
		&videoJSON,
		&placement.CreatedAt,
		&placement.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("placement not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	if width.Valid {
		placement.Width = int(width.Int32)
	}
	if height.Valid {
		placement.Height = int(height.Int32)
	}

	if len(formatsJSON) > 0 {
		if err := json.Unmarshal(formatsJSON, &placement.Formats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal formats: %w", err)
		}
	}

	if len(videoJSON) > 0 {
		if err := json.Unmarshal(videoJSON, &placement.Video); err != nil {
			return nil, fmt.Errorf("failed to unmarshal video settings: %w", err)
		}
	}

	return placement, nil
}

// ListPlacements lists placements for a site
func (ps *PostgresStore) ListPlacements(ctx context.Context, siteID string, activeOnly bool) ([]*Placement, error) {
	query := `
		SELECT id, site_id, name, ad_type, width, height, min_bid_floor, active, formats, video, created_at, updated_at
		FROM placements
		WHERE site_id = $1
	`

	if activeOnly {
		query += " AND active = true"
	}

	query += " ORDER BY created_at DESC"

	rows, err := ps.db.QueryContext(ctx, query, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	placements := []*Placement{}

	for rows.Next() {
		placement := &Placement{}
		var width, height sql.NullInt32
		var formatsJSON, videoJSON []byte

		err := rows.Scan(
			&placement.ID,
			&placement.SiteID,
			&placement.Name,
			&placement.AdType,
			&width,
			&height,
			&placement.MinBidFloor,
			&placement.Active,
			&formatsJSON,
			&videoJSON,
			&placement.CreatedAt,
			&placement.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if width.Valid {
			placement.Width = int(width.Int32)
		}
		if height.Valid {
			placement.Height = int(height.Int32)
		}

		if len(formatsJSON) > 0 {
			if err := json.Unmarshal(formatsJSON, &placement.Formats); err != nil {
				return nil, fmt.Errorf("failed to unmarshal formats: %w", err)
			}
		}

		if len(videoJSON) > 0 {
			if err := json.Unmarshal(videoJSON, &placement.Video); err != nil {
				return nil, fmt.Errorf("failed to unmarshal video settings: %w", err)
			}
		}

		placements = append(placements, placement)
	}

	return placements, nil
}

// UpdatePlacement updates a placement
func (ps *PostgresStore) UpdatePlacement(ctx context.Context, placement *Placement) error {
	formatsJSON, err := json.Marshal(placement.Formats)
	if err != nil {
		return fmt.Errorf("failed to marshal formats: %w", err)
	}

	videoJSON, err := json.Marshal(placement.Video)
	if err != nil {
		return fmt.Errorf("failed to marshal video settings: %w", err)
	}

	query := `
		UPDATE placements
		SET name = $2, ad_type = $3, width = $4, height = $5, min_bid_floor = $6, active = $7, formats = $8, video = $9, updated_at = $10
		WHERE id = $1
	`

	_, err = ps.db.ExecContext(ctx, query,
		placement.ID,
		placement.Name,
		placement.AdType,
		placement.Width,
		placement.Height,
		placement.MinBidFloor,
		placement.Active,
		formatsJSON,
		videoJSON,
		placement.UpdatedAt,
	)

	return err
}

// DeletePlacement deletes a placement
func (ps *PostgresStore) DeletePlacement(ctx context.Context, id string) error {
	query := "DELETE FROM placements WHERE id = $1"
	_, err := ps.db.ExecContext(ctx, query, id)
	return err
}

// Close closes the database connection
func (ps *PostgresStore) Close() error {
	return ps.db.Close()
}
