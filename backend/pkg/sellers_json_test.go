package ssp

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewSellersJSONGenerator(t *testing.T) {
	email := "contact@ad.nexus"
	address := "123 Main St, City, State"

	gen := NewSellersJSONGenerator(email, address)

	if gen.ContactEmail != email {
		t.Errorf("Expected contact email %s, got %s", email, gen.ContactEmail)
	}
	if gen.ContactAddress != address {
		t.Errorf("Expected contact address %s, got %s", address, gen.ContactAddress)
	}
	if gen.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", gen.Version)
	}
}

func TestAddIdentifier(t *testing.T) {
	gen := NewSellersJSONGenerator("test@example.com", "")

	gen.AddIdentifier("TAG-ID", "12345")
	gen.AddIdentifier("IRS-TAX-ID", "67890")

	if len(gen.Identifiers) != 2 {
		t.Errorf("Expected 2 identifiers, got %d", len(gen.Identifiers))
	}

	if gen.Identifiers[0].Name != "TAG-ID" {
		t.Errorf("Expected first identifier name TAG-ID, got %s", gen.Identifiers[0].Name)
	}
	if gen.Identifiers[0].Value != "12345" {
		t.Errorf("Expected first identifier value 12345, got %s", gen.Identifiers[0].Value)
	}
}

func TestGenerateFromPublishers(t *testing.T) {
	gen := NewSellersJSONGenerator("contact@ad.nexus", "123 Main St")
	gen.AddIdentifier("TAG-ID", "ABC123")

	publishers := []Publisher{
		{
			ID:     "pub-001",
			Name:   "Publisher One",
			Domain: "publisher1.com",
			Email:  "pub1@example.com",
			Active: true,
		},
		{
			ID:     "pub-002",
			Name:   "Publisher Two",
			Domain: "publisher2.com",
			Email:  "pub2@example.com",
			Active: true,
		},
		{
			ID:     "pub-003",
			Name:   "Inactive Publisher",
			Domain: "inactive.com",
			Email:  "inactive@example.com",
			Active: false, // Inactive - should be skipped
		},
		{
			ID:     "pub-004",
			Name:   "Confidential Publisher",
			Domain: "", // No domain - should be confidential
			Email:  "confidential@example.com",
			Active: true,
		},
	}

	sellersJSON, err := gen.GenerateFromPublishers(publishers)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should include 3 active publishers (skip inactive)
	if len(sellersJSON.Sellers) != 3 {
		t.Errorf("Expected 3 sellers, got %d", len(sellersJSON.Sellers))
	}

	// Check contact info
	if sellersJSON.ContactEmail != "contact@ad.nexus" {
		t.Errorf("Expected contact email contact@ad.nexus, got %s", sellersJSON.ContactEmail)
	}

	// Check identifier
	if len(sellersJSON.Identifiers) != 1 {
		t.Errorf("Expected 1 identifier, got %d", len(sellersJSON.Identifiers))
	}

	// Verify first seller
	seller1 := sellersJSON.Sellers[0]
	if seller1.SellerID != "pub-001" {
		t.Errorf("Expected seller ID pub-001, got %s", seller1.SellerID)
	}
	if seller1.Name != "Publisher One" {
		t.Errorf("Expected seller name 'Publisher One', got %s", seller1.Name)
	}
	if seller1.SellerType != "PUBLISHER" {
		t.Errorf("Expected seller type PUBLISHER, got %s", seller1.SellerType)
	}
	if seller1.IsConfidential != 0 {
		t.Errorf("Expected is_confidential 0, got %d", seller1.IsConfidential)
	}

	// Verify confidential seller (no domain)
	seller4 := sellersJSON.Sellers[2]
	if seller4.IsConfidential != 1 {
		t.Errorf("Expected is_confidential 1 for seller without domain, got %d", seller4.IsConfidential)
	}
}

func TestToJSON(t *testing.T) {
	sellersJSON := &SellersJSON{
		ContactEmail: "contact@ad.nexus",
		Version:      "1.0",
		Identifiers: []Identifier{
			{Name: "TAG-ID", Value: "12345"},
		},
		Sellers: []Seller{
			{
				SellerID:   "pub-001",
				Name:       "Test Publisher",
				Domain:     "test.com",
				SellerType: "PUBLISHER",
			},
		},
	}

	jsonData, err := sellersJSON.ToJSON()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify JSON structure
	var decoded SellersJSON
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if decoded.ContactEmail != "contact@ad.nexus" {
		t.Errorf("Expected contact email contact@ad.nexus, got %s", decoded.ContactEmail)
	}
	if len(decoded.Sellers) != 1 {
		t.Errorf("Expected 1 seller, got %d", len(decoded.Sellers))
	}
	if decoded.Sellers[0].SellerID != "pub-001" {
		t.Errorf("Expected seller ID pub-001, got %s", decoded.Sellers[0].SellerID)
	}
}

func TestValidateSellersJSON(t *testing.T) {
	validJSON := []byte(`{
		"contact_email": "contact@ad.nexus",
		"version": "1.0",
		"sellers": [
			{
				"seller_id": "pub-001",
				"name": "Test Publisher",
				"domain": "test.com",
				"seller_type": "PUBLISHER"
			}
		]
	}`)

	err := ValidateSellersJSON(validJSON)
	if err != nil {
		t.Errorf("Expected valid JSON, got error: %v", err)
	}

	invalidJSON := []byte(`{invalid json}`)
	err = ValidateSellersJSON(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got none")
	}
}

func TestSellersJSONCache(t *testing.T) {
	cache := NewSellersJSONCache(1 * time.Second)

	// Test empty cache
	data, ok := cache.Get()
	if ok {
		t.Error("Expected cache miss for empty cache")
	}
	if data != nil {
		t.Error("Expected nil data for empty cache")
	}

	// Test set and get
	testData := []byte(`{"test": "data"}`)
	cache.Set(testData)

	data, ok = cache.Get()
	if !ok {
		t.Error("Expected cache hit after Set")
	}
	if string(data) != string(testData) {
		t.Errorf("Expected data %s, got %s", testData, data)
	}

	// Test cache expiration
	cache2 := NewSellersJSONCache(10 * time.Millisecond)
	cache2.Set(testData)

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	data, ok = cache2.Get()
	if ok {
		t.Error("Expected cache miss after expiration")
	}
}

func TestSellersJSONCacheIsExpired(t *testing.T) {
	cache := NewSellersJSONCache(50 * time.Millisecond)

	if !cache.IsExpired() {
		t.Error("Expected cache to be expired when empty")
	}

	cache.Set([]byte("test"))
	if cache.IsExpired() {
		t.Error("Expected cache to not be expired immediately after Set")
	}

	time.Sleep(60 * time.Millisecond)
	if !cache.IsExpired() {
		t.Error("Expected cache to be expired after TTL")
	}
}

func TestSellersJSONCacheClear(t *testing.T) {
	cache := NewSellersJSONCache(1 * time.Second)
	cache.Set([]byte("test data"))

	// Verify data is cached
	data, ok := cache.Get()
	if !ok || data == nil {
		t.Error("Expected cache to contain data before Clear")
	}

	// Clear cache
	cache.Clear()

	// Verify cache is empty
	data, ok = cache.Get()
	if ok || data != nil {
		t.Error("Expected cache to be empty after Clear")
	}
}

func TestSellerTypes(t *testing.T) {
	// Test all valid seller types
	validTypes := []string{"PUBLISHER", "INTERMEDIARY", "BOTH"}

	for _, sellerType := range validTypes {
		seller := Seller{
			SellerID:   "test",
			Name:       "Test",
			Domain:     "test.com",
			SellerType: sellerType,
		}

		if seller.SellerType != sellerType {
			t.Errorf("Expected seller type %s, got %s", sellerType, seller.SellerType)
		}
	}
}

func TestIsPassthrough(t *testing.T) {
	// Test intermediary with passthrough
	intermediary := Seller{
		SellerID:      "int-001",
		Name:          "Intermediary",
		Domain:        "intermediary.com",
		SellerType:    "INTERMEDIARY",
		IsPassthrough: 1,
	}

	if intermediary.IsPassthrough != 1 {
		t.Errorf("Expected is_passthrough 1, got %d", intermediary.IsPassthrough)
	}

	// Test publisher without passthrough
	publisher := Seller{
		SellerID:   "pub-001",
		Name:       "Publisher",
		Domain:     "publisher.com",
		SellerType: "PUBLISHER",
	}

	if publisher.IsPassthrough != 0 {
		t.Errorf("Expected is_passthrough 0, got %d", publisher.IsPassthrough)
	}
}
