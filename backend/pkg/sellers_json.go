package ssp

import (
	"encoding/json"
	"time"
)

// SellersJSON represents the IAB sellers.json specification
// https://iabtechlab.com/sellers-json/
type SellersJSON struct {
	ContactEmail   string       `json:"contact_email"`
	ContactAddress string       `json:"contact_address,omitempty"`
	Version        string       `json:"version"`
	Identifiers    []Identifier `json:"identifiers,omitempty"`
	Sellers        []Seller     `json:"sellers"`
}

// Seller represents a single seller entry in sellers.json
type Seller struct {
	SellerID       string `json:"seller_id"`
	Name           string `json:"name,omitempty"`
	Domain         string `json:"domain,omitempty"`
	SellerType     string `json:"seller_type"`               // "PUBLISHER", "INTERMEDIARY", "BOTH"
	IsConfidential int    `json:"is_confidential,omitempty"` // 0=public, 1=confidential
	IsPassthrough  int    `json:"is_passthrough,omitempty"`  // 0=no, 1=yes (for intermediaries)
}

// Identifier represents a seller identifier
type Identifier struct {
	Name  string `json:"name"` // e.g., "TAG-ID", "IRS-TAX-ID"
	Value string `json:"value"`
}

// SellersJSONGenerator generates sellers.json from publisher data
type SellersJSONGenerator struct {
	ContactEmail   string
	ContactAddress string
	Version        string
	Identifiers    []Identifier
}

// NewSellersJSONGenerator creates a new sellers.json generator
func NewSellersJSONGenerator(contactEmail, contactAddress string) *SellersJSONGenerator {
	return &SellersJSONGenerator{
		ContactEmail:   contactEmail,
		ContactAddress: contactAddress,
		Version:        "1.0",
		Identifiers:    []Identifier{},
	}
}

// AddIdentifier adds a seller identifier (e.g., TAG-ID)
func (g *SellersJSONGenerator) AddIdentifier(name, value string) {
	g.Identifiers = append(g.Identifiers, Identifier{
		Name:  name,
		Value: value,
	})
}

// GenerateFromPublishers generates sellers.json from a list of publishers
func (g *SellersJSONGenerator) GenerateFromPublishers(publishers []Publisher) (*SellersJSON, error) {
	sellers := make([]Seller, 0, len(publishers))

	for _, pub := range publishers {
		if !pub.Active {
			continue // Skip inactive publishers
		}

		seller := Seller{
			SellerID:       pub.ID,
			Name:           pub.Name,
			Domain:         pub.Domain,
			SellerType:     "PUBLISHER",
			IsConfidential: 0, // Default to public
		}

		// If no domain provided, mark as confidential
		if pub.Domain == "" {
			seller.IsConfidential = 1
		}

		sellers = append(sellers, seller)
	}

	sellersJSON := &SellersJSON{
		ContactEmail:   g.ContactEmail,
		ContactAddress: g.ContactAddress,
		Version:        g.Version,
		Identifiers:    g.Identifiers,
		Sellers:        sellers,
	}

	return sellersJSON, nil
}

// ToJSON converts sellers.json to JSON string
func (s *SellersJSON) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// ValidateSellersJSON validates sellers.json structure
func ValidateSellersJSON(data []byte) error {
	var sellersJSON SellersJSON
	return json.Unmarshal(data, &sellersJSON)
}

// SellersJSONCache provides caching for sellers.json
type SellersJSONCache struct {
	data      []byte
	updatedAt time.Time
	ttl       time.Duration
}

// NewSellersJSONCache creates a new sellers.json cache with TTL
func NewSellersJSONCache(ttl time.Duration) *SellersJSONCache {
	return &SellersJSONCache{
		ttl: ttl,
	}
}

// Set updates the cached sellers.json data
func (c *SellersJSONCache) Set(data []byte) {
	c.data = data
	c.updatedAt = time.Now()
}

// Get returns cached sellers.json if not expired
func (c *SellersJSONCache) Get() ([]byte, bool) {
	if c.data == nil {
		return nil, false
	}

	if time.Since(c.updatedAt) > c.ttl {
		return nil, false // Expired
	}

	return c.data, true
}

// IsExpired checks if cache is expired
func (c *SellersJSONCache) IsExpired() bool {
	if c.data == nil {
		return true
	}
	return time.Since(c.updatedAt) > c.ttl
}

// Clear clears the cache
func (c *SellersJSONCache) Clear() {
	c.data = nil
	c.updatedAt = time.Time{}
}
