package ssp

import (
	"encoding/json"
	"testing"
)

func TestNewSupplyChainBuilder(t *testing.T) {
	asi := "ad.nexus"
	sid := "ssp-001"
	name := "AdNexus"
	domain := "ad.nexus"

	builder := NewSupplyChainBuilder(asi, sid, name, domain)

	if builder.ourASI != asi {
		t.Errorf("Expected ASI %s, got %s", asi, builder.ourASI)
	}
	if builder.ourSID != sid {
		t.Errorf("Expected SID %s, got %s", sid, builder.ourSID)
	}
	if builder.ourName != name {
		t.Errorf("Expected name %s, got %s", name, builder.ourName)
	}
	if builder.ourDomain != domain {
		t.Errorf("Expected domain %s, got %s", domain, builder.ourDomain)
	}
}

func TestBuildForPublisher(t *testing.T) {
	builder := NewSupplyChainBuilder("ad.nexus", "ssp-001", "AdNexus", "ad.nexus")

	// Test valid publisher
	schain, err := builder.BuildForPublisher("pub-001", "publisher.com")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if schain.Complete != 1 {
		t.Errorf("Expected complete=1, got %d", schain.Complete)
	}
	if schain.Ver != "1.0" {
		t.Errorf("Expected version 1.0, got %s", schain.Ver)
	}
	if len(schain.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(schain.Nodes))
	}

	node := schain.Nodes[0]
	if node.ASI != "ad.nexus" {
		t.Errorf("Expected ASI ad.nexus, got %s", node.ASI)
	}
	if node.SID != "pub-001" {
		t.Errorf("Expected SID pub-001, got %s", node.SID)
	}
	if node.HP != 0 {
		t.Errorf("Expected HP=0 (direct), got %d", node.HP)
	}
	if node.Domain != "publisher.com" {
		t.Errorf("Expected domain publisher.com, got %s", node.Domain)
	}

	// Test missing publisher ID
	_, err = builder.BuildForPublisher("", "publisher.com")
	if err == nil {
		t.Error("Expected error for missing publisher ID")
	}
}

func TestBuildForIntermediary(t *testing.T) {
	builder := NewSupplyChainBuilder("ad.nexus", "ssp-001", "AdNexus", "ad.nexus")

	intermediaries := []SupplyChainNode{
		{
			ASI:    "intermediary1.com",
			SID:    "int-001",
			HP:     1, // Reseller
			Name:   "Intermediary 1",
			Domain: "intermediary1.com",
		},
		{
			ASI:    "intermediary2.com",
			SID:    "int-002",
			HP:     1, // Reseller
			Name:   "Intermediary 2",
			Domain: "intermediary2.com",
		},
	}

	schain, err := builder.BuildForIntermediary("pub-001", "publisher.com", intermediaries)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if schain.Complete != 1 {
		t.Errorf("Expected complete=1, got %d", schain.Complete)
	}

	// Should have 3 nodes total (2 intermediaries + our node)
	if len(schain.Nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(schain.Nodes))
	}

	// Verify first intermediary
	if schain.Nodes[0].ASI != "intermediary1.com" {
		t.Errorf("Expected first node ASI intermediary1.com, got %s", schain.Nodes[0].ASI)
	}
	if schain.Nodes[0].HP != 1 {
		t.Errorf("Expected first node HP=1 (reseller), got %d", schain.Nodes[0].HP)
	}

	// Verify our node is last
	lastNode := schain.Nodes[len(schain.Nodes)-1]
	if lastNode.ASI != "ad.nexus" {
		t.Errorf("Expected last node ASI ad.nexus, got %s", lastNode.ASI)
	}
	if lastNode.SID != "pub-001" {
		t.Errorf("Expected last node SID pub-001, got %s", lastNode.SID)
	}
	if lastNode.HP != 0 {
		t.Errorf("Expected last node HP=0 (direct), got %d", lastNode.HP)
	}
}

func TestBuildIncomplete(t *testing.T) {
	builder := NewSupplyChainBuilder("ad.nexus", "ssp-001", "AdNexus", "ad.nexus")

	schain, err := builder.BuildIncomplete("pub-001")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if schain.Complete != 0 {
		t.Errorf("Expected complete=0, got %d", schain.Complete)
	}
	if len(schain.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(schain.Nodes))
	}

	// Test missing publisher ID
	_, err = builder.BuildIncomplete("")
	if err == nil {
		t.Error("Expected error for missing publisher ID")
	}
}

func TestValidateSupplyChain(t *testing.T) {
	tests := []struct {
		name        string
		schain      *SupplyChain
		expectError bool
	}{
		{
			name:        "Nil supply chain",
			schain:      nil,
			expectError: true,
		},
		{
			name: "Empty nodes",
			schain: &SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes:    []SupplyChainNode{},
			},
			expectError: true,
		},
		{
			name: "Missing version",
			schain: &SupplyChain{
				Complete: 1,
				Nodes: []SupplyChainNode{
					{ASI: "test.com", SID: "test-001", HP: 0},
				},
			},
			expectError: true,
		},
		{
			name: "Valid supply chain",
			schain: &SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes: []SupplyChainNode{
					{ASI: "test.com", SID: "test-001", HP: 0},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid HP value",
			schain: &SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes: []SupplyChainNode{
					{ASI: "test.com", SID: "test-001", HP: 2}, // Invalid
				},
			},
			expectError: true,
		},
		{
			name: "Missing ASI",
			schain: &SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes: []SupplyChainNode{
					{SID: "test-001", HP: 0}, // Missing ASI
				},
			},
			expectError: true,
		},
		{
			name: "Missing SID",
			schain: &SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes: []SupplyChainNode{
					{ASI: "test.com", HP: 0}, // Missing SID
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSupplyChain(tt.schain)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestSupplyChainToJSON(t *testing.T) {
	schain := &SupplyChain{
		Complete: 1,
		Ver:      "1.0",
		Nodes: []SupplyChainNode{
			{
				ASI:    "ad.nexus",
				SID:    "pub-001",
				HP:     0,
				Name:   "AdNexus",
				Domain: "ad.nexus",
			},
		},
	}

	jsonData, err := schain.ToJSON()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify JSON structure
	var decoded SupplyChain
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if decoded.Complete != 1 {
		t.Errorf("Expected complete=1, got %d", decoded.Complete)
	}
	if decoded.Ver != "1.0" {
		t.Errorf("Expected version 1.0, got %s", decoded.Ver)
	}
	if len(decoded.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(decoded.Nodes))
	}
}

func TestFromJSON(t *testing.T) {
	jsonData := []byte(`{
		"complete": 1,
		"ver": "1.0",
		"nodes": [
			{
				"asi": "ad.nexus",
				"sid": "pub-001",
				"hp": 0,
				"name": "AdNexus",
				"domain": "ad.nexus"
			}
		]
	}`)

	schain, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if schain.Complete != 1 {
		t.Errorf("Expected complete=1, got %d", schain.Complete)
	}
	if schain.Ver != "1.0" {
		t.Errorf("Expected version 1.0, got %s", schain.Ver)
	}
	if len(schain.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(schain.Nodes))
	}
	if schain.Nodes[0].ASI != "ad.nexus" {
		t.Errorf("Expected ASI ad.nexus, got %s", schain.Nodes[0].ASI)
	}
}

func TestAddToSource(t *testing.T) {
	schain := &SupplyChain{
		Complete: 1,
		Ver:      "1.0",
		Nodes: []SupplyChainNode{
			{ASI: "ad.nexus", SID: "pub-001", HP: 0},
		},
	}

	source := &Source{
		FD:  1,
		TID: "txn-001",
	}

	err := AddToSource(source, schain)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if source.Ext == nil {
		t.Fatal("Expected source.ext to be set")
	}

	// Verify schain is in ext
	extracted, err := ExtractFromSource(source)
	if err != nil {
		t.Fatalf("Failed to extract supply chain: %v", err)
	}

	if extracted.Complete != 1 {
		t.Errorf("Expected complete=1, got %d", extracted.Complete)
	}
	if len(extracted.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(extracted.Nodes))
	}
}

func TestExtractFromSource(t *testing.T) {
	// Create source with schain in ext
	extJSON := json.RawMessage(`{
		"schain": {
			"complete": 1,
			"ver": "1.0",
			"nodes": [
				{
					"asi": "ad.nexus",
					"sid": "pub-001",
					"hp": 0
				}
			]
		}
	}`)

	source := &Source{
		FD:  1,
		TID: "txn-001",
		Ext: extJSON,
	}

	schain, err := ExtractFromSource(source)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if schain.Complete != 1 {
		t.Errorf("Expected complete=1, got %d", schain.Complete)
	}
	if len(schain.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(schain.Nodes))
	}

	// Test source without ext
	source2 := &Source{FD: 1}
	_, err = ExtractFromSource(source2)
	if err == nil {
		t.Error("Expected error for source without ext")
	}

	// Test nil source
	_, err = ExtractFromSource(nil)
	if err == nil {
		t.Error("Expected error for nil source")
	}
}

func TestIsComplete(t *testing.T) {
	complete := &SupplyChain{Complete: 1}
	if !complete.IsComplete() {
		t.Error("Expected IsComplete() to return true")
	}

	incomplete := &SupplyChain{Complete: 0}
	if incomplete.IsComplete() {
		t.Error("Expected IsComplete() to return false")
	}
}

func TestNodeCount(t *testing.T) {
	schain := &SupplyChain{
		Nodes: []SupplyChainNode{
			{ASI: "node1.com", SID: "sid1"},
			{ASI: "node2.com", SID: "sid2"},
			{ASI: "node3.com", SID: "sid3"},
		},
	}

	if schain.NodeCount() != 3 {
		t.Errorf("Expected NodeCount() to return 3, got %d", schain.NodeCount())
	}

	empty := &SupplyChain{Nodes: []SupplyChainNode{}}
	if empty.NodeCount() != 0 {
		t.Errorf("Expected NodeCount() to return 0, got %d", empty.NodeCount())
	}
}

func TestGetFirstNode(t *testing.T) {
	schain := &SupplyChain{
		Nodes: []SupplyChainNode{
			{ASI: "first.com", SID: "sid1"},
			{ASI: "second.com", SID: "sid2"},
		},
	}

	first := schain.GetFirstNode()
	if first == nil {
		t.Fatal("Expected GetFirstNode() to return node")
	}
	if first.ASI != "first.com" {
		t.Errorf("Expected first node ASI first.com, got %s", first.ASI)
	}

	empty := &SupplyChain{Nodes: []SupplyChainNode{}}
	if empty.GetFirstNode() != nil {
		t.Error("Expected GetFirstNode() to return nil for empty chain")
	}
}

func TestGetLastNode(t *testing.T) {
	schain := &SupplyChain{
		Nodes: []SupplyChainNode{
			{ASI: "first.com", SID: "sid1"},
			{ASI: "last.com", SID: "sid2"},
		},
	}

	last := schain.GetLastNode()
	if last == nil {
		t.Fatal("Expected GetLastNode() to return node")
	}
	if last.ASI != "last.com" {
		t.Errorf("Expected last node ASI last.com, got %s", last.ASI)
	}

	empty := &SupplyChain{Nodes: []SupplyChainNode{}}
	if empty.GetLastNode() != nil {
		t.Error("Expected GetLastNode() to return nil for empty chain")
	}
}

func TestHasResellers(t *testing.T) {
	// Chain with reseller
	withReseller := &SupplyChain{
		Nodes: []SupplyChainNode{
			{ASI: "reseller.com", SID: "sid1", HP: 1}, // HP=1 means reseller
			{ASI: "direct.com", SID: "sid2", HP: 0},
		},
	}

	if !withReseller.HasResellers() {
		t.Error("Expected HasResellers() to return true")
	}

	// Chain without reseller
	withoutReseller := &SupplyChain{
		Nodes: []SupplyChainNode{
			{ASI: "direct1.com", SID: "sid1", HP: 0},
			{ASI: "direct2.com", SID: "sid2", HP: 0},
		},
	}

	if withoutReseller.HasResellers() {
		t.Error("Expected HasResellers() to return false")
	}
}
