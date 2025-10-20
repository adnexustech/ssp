package ssp

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SupplyChain represents the OpenRTB SupplyChain object (ads.cert 1.0)
// https://github.com/InteractiveAdvertisingBureau/openrtb/blob/master/supplychainobject.md
type SupplyChain struct {
	Complete int               `json:"complete"` // 1=complete, 0=incomplete
	Nodes    []SupplyChainNode `json:"nodes"`
	Ver      string            `json:"ver"` // Version of supply chain spec (e.g., "1.0")
}

// SupplyChainNode represents a single node in the supply chain
type SupplyChainNode struct {
	ASI    string `json:"asi"`              // Advertising System Identifier (domain)
	SID    string `json:"sid"`              // Seller ID
	HP     int    `json:"hp,omitempty"`     // 1=reseller, 0=direct relationship
	RID    string `json:"rid,omitempty"`    // Request ID
	Name   string `json:"name,omitempty"`   // Business name
	Domain string `json:"domain,omitempty"` // Business domain
}

// SupplyChainBuilder builds supply chain objects for bid requests
type SupplyChainBuilder struct {
	ourASI    string // Our advertising system identifier (domain)
	ourSID    string // Our seller ID
	ourName   string // Our business name
	ourDomain string // Our business domain
}

// NewSupplyChainBuilder creates a new supply chain builder
func NewSupplyChainBuilder(asi, sid, name, domain string) *SupplyChainBuilder {
	return &SupplyChainBuilder{
		ourASI:    asi,
		ourSID:    sid,
		ourName:   name,
		ourDomain: domain,
	}
}

// BuildForPublisher builds a complete supply chain for a direct publisher
func (b *SupplyChainBuilder) BuildForPublisher(publisherID, publisherDomain string) (*SupplyChain, error) {
	if publisherID == "" {
		return nil, errors.New("publisher ID is required")
	}

	// For direct publishers, create a complete chain with one node
	schain := &SupplyChain{
		Complete: 1, // Complete chain
		Ver:      "1.0",
		Nodes: []SupplyChainNode{
			{
				ASI:    b.ourASI,    // Our domain
				SID:    publisherID, // Publisher's seller ID
				HP:     0,           // Direct relationship (not a reseller)
				Name:   b.ourName,
				Domain: publisherDomain,
			},
		},
	}

	return schain, nil
}

// BuildForIntermediary builds supply chain including intermediary nodes
func (b *SupplyChainBuilder) BuildForIntermediary(publisherID, publisherDomain string, intermediaries []SupplyChainNode) (*SupplyChain, error) {
	if publisherID == "" {
		return nil, errors.New("publisher ID is required")
	}

	// Start with intermediary nodes
	nodes := make([]SupplyChainNode, 0, len(intermediaries)+1)
	nodes = append(nodes, intermediaries...)

	// Add our node at the end (we're the SSP)
	nodes = append(nodes, SupplyChainNode{
		ASI:    b.ourASI,
		SID:    publisherID,
		HP:     0, // Direct relationship with publisher
		Name:   b.ourName,
		Domain: publisherDomain,
	})

	schain := &SupplyChain{
		Complete: 1, // Complete if all intermediaries are verified
		Ver:      "1.0",
		Nodes:    nodes,
	}

	return schain, nil
}

// BuildIncomplete builds an incomplete supply chain (when chain info is unknown)
func (b *SupplyChainBuilder) BuildIncomplete(publisherID string) (*SupplyChain, error) {
	if publisherID == "" {
		return nil, errors.New("publisher ID is required")
	}

	schain := &SupplyChain{
		Complete: 0, // Incomplete chain
		Ver:      "1.0",
		Nodes: []SupplyChainNode{
			{
				ASI: b.ourASI,
				SID: publisherID,
				HP:  0,
			},
		},
	}

	return schain, nil
}

// ValidateSupplyChain validates a supply chain object
func ValidateSupplyChain(schain *SupplyChain) error {
	if schain == nil {
		return errors.New("supply chain is nil")
	}

	if len(schain.Nodes) == 0 {
		return errors.New("supply chain must have at least one node")
	}

	if schain.Ver == "" {
		return errors.New("supply chain version is required")
	}

	// Validate each node
	for i, node := range schain.Nodes {
		if err := validateNode(&node); err != nil {
			return fmt.Errorf("node %d: %w", i, err)
		}
	}

	return nil
}

// validateNode validates a single supply chain node
func validateNode(node *SupplyChainNode) error {
	if node.ASI == "" {
		return errors.New("ASI (Advertising System Identifier) is required")
	}

	if node.SID == "" {
		return errors.New("SID (Seller ID) is required")
	}

	// HP must be 0 or 1
	if node.HP != 0 && node.HP != 1 {
		return fmt.Errorf("HP must be 0 or 1, got %d", node.HP)
	}

	return nil
}

// ToJSON converts supply chain to JSON
func (s *SupplyChain) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON parses supply chain from JSON
func FromJSON(data []byte) (*SupplyChain, error) {
	var schain SupplyChain
	if err := json.Unmarshal(data, &schain); err != nil {
		return nil, err
	}
	return &schain, nil
}

// AddToSource adds supply chain to OpenRTB Source object ext
func AddToSource(source *Source, schain *SupplyChain) error {
	if source == nil {
		return errors.New("source is nil")
	}

	if schain == nil {
		return errors.New("supply chain is nil")
	}

	// Validate supply chain before adding
	if err := ValidateSupplyChain(schain); err != nil {
		return fmt.Errorf("invalid supply chain: %w", err)
	}

	// Convert to JSON for Source.Ext
	schainJSON, err := schain.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert supply chain to JSON: %w", err)
	}

	// Create or update Source.Ext with schain
	extMap := make(map[string]interface{})
	if source.Ext != nil {
		// Parse existing ext
		if extStr, ok := source.Ext.(string); ok {
			json.Unmarshal([]byte(extStr), &extMap)
		} else if extBytes, ok := source.Ext.([]byte); ok {
			json.Unmarshal(extBytes, &extMap)
		}
	}

	// Add schain to ext
	var schainObj interface{}
	json.Unmarshal(schainJSON, &schainObj)
	extMap["schain"] = schainObj

	// Convert back to JSON
	extJSON, err := json.Marshal(extMap)
	if err != nil {
		return fmt.Errorf("failed to marshal ext: %w", err)
	}

	source.Ext = json.RawMessage(extJSON)
	return nil
}

// ExtractFromSource extracts supply chain from OpenRTB Source object ext
func ExtractFromSource(source *Source) (*SupplyChain, error) {
	if source == nil || source.Ext == nil {
		return nil, errors.New("no supply chain in source")
	}

	// Parse ext
	extMap := make(map[string]interface{})
	switch ext := source.Ext.(type) {
	case string:
		if err := json.Unmarshal([]byte(ext), &extMap); err != nil {
			return nil, err
		}
	case []byte:
		if err := json.Unmarshal(ext, &extMap); err != nil {
			return nil, err
		}
	case json.RawMessage:
		if err := json.Unmarshal(ext, &extMap); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported ext type")
	}

	// Extract schain
	schainObj, ok := extMap["schain"]
	if !ok {
		return nil, errors.New("no schain in source.ext")
	}

	// Convert to JSON and parse
	schainJSON, err := json.Marshal(schainObj)
	if err != nil {
		return nil, err
	}

	return FromJSON(schainJSON)
}

// IsComplete checks if supply chain is complete
func (s *SupplyChain) IsComplete() bool {
	return s.Complete == 1
}

// NodeCount returns the number of nodes in the supply chain
func (s *SupplyChain) NodeCount() int {
	return len(s.Nodes)
}

// GetFirstNode returns the first node in the supply chain (origin)
func (s *SupplyChain) GetFirstNode() *SupplyChainNode {
	if len(s.Nodes) == 0 {
		return nil
	}
	return &s.Nodes[0]
}

// GetLastNode returns the last node in the supply chain (final seller)
func (s *SupplyChain) GetLastNode() *SupplyChainNode {
	if len(s.Nodes) == 0 {
		return nil
	}
	return &s.Nodes[len(s.Nodes)-1]
}

// HasResellers checks if there are any resellers in the chain
func (s *SupplyChain) HasResellers() bool {
	for _, node := range s.Nodes {
		if node.HP == 1 {
			return true
		}
	}
	return false
}
