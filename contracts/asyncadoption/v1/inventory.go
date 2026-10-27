package asyncadoptionv1

import (
	"encoding/json"
	"fmt"
	"sort"
)

type Inventory struct {
	ContractVersion         string `json:"contract_version"`
	RunID                   string `json:"run_id"`
	ScopeKind               string `json:"scope_kind"`
	ScopeID                 string `json:"scope_id"`
	Domain                  string `json:"domain"`
	GroupID                 string `json:"group_id"`
	SourceGeneration        int64  `json:"source_generation"`
	TargetGeneration        int64  `json:"target_generation"`
	SourceGroupGeneration   int64  `json:"source_group_generation"`
	TargetGroupGeneration   int64  `json:"target_group_generation"`
	InventoryRevision       int    `json:"inventory_revision"`
	ComponentManifestSHA256 string `json:"component_manifest_sha256"`
	ExpectedEnqueueFences   int    `json:"expected_enqueue_fences"`
	ExpectedSchedulerFences int    `json:"expected_scheduler_fences"`
	ExpectedConsumerFences  int    `json:"expected_consumer_fences"`
	CapturedAt              string `json:"captured_at"`
	ItemCount               int    `json:"item_count"`
	InventorySHA256         string `json:"inventory_sha256"`
	Items                   []Item `json:"items"`
}

func NewInventory(value Inventory) (Inventory, error) {
	value.ContractVersion = InventoryContractVersion
	value.InventoryRevision = 1
	value.Items = append([]Item(nil), value.Items...)
	sort.Slice(value.Items, func(left, right int) bool { return itemLess(value.Items[left], value.Items[right]) })
	value.ItemCount = len(value.Items)
	lines := make([]string, len(value.Items))
	for index := range value.Items {
		lines[index] = value.Items[index].ItemSHA256
	}
	value.InventorySHA256 = digestLines(lines)
	return value, value.Validate()
}

func (value Inventory) Validate() error {
	if value.ContractVersion != InventoryContractVersion || !validUUID(value.RunID) ||
		(value.ScopeKind != "tenant" && value.ScopeKind != "global") || !validUUID(value.ScopeID) ||
		!domainName.MatchString(value.Domain) || !domainName.MatchString(value.GroupID) {
		return fmt.Errorf("async adoption inventory authority is invalid")
	}
	if value.SourceGeneration < 1 || value.TargetGeneration != value.SourceGeneration+1 ||
		value.SourceGroupGeneration < 1 ||
		value.TargetGroupGeneration != value.SourceGroupGeneration+1 || value.InventoryRevision != 1 {
		return fmt.Errorf("async adoption inventory generation is invalid")
	}
	if !lowerSHA256.MatchString(value.ComponentManifestSHA256) ||
		value.ExpectedEnqueueFences < 1 || value.ExpectedSchedulerFences < 1 ||
		value.ExpectedConsumerFences < 1 || value.ExpectedEnqueueFences > 64 ||
		value.ExpectedSchedulerFences > 64 || value.ExpectedConsumerFences > 64 {
		return fmt.Errorf("async adoption component manifest is invalid")
	}
	if !validTimestamp(value.CapturedAt) || value.ItemCount != len(value.Items) ||
		value.ItemCount > MaximumInventoryItems || !lowerSHA256.MatchString(value.InventorySHA256) {
		return fmt.Errorf("async adoption inventory capture is invalid")
	}
	lines := make([]string, len(value.Items))
	for index, item := range value.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("inventory item %d: %w", index, err)
		}
		if index > 0 && !itemLess(value.Items[index-1], item) {
			return fmt.Errorf("inventory item identity is duplicate or unsorted")
		}
		lines[index] = item.ItemSHA256
	}
	if digestLines(lines) != value.InventorySHA256 {
		return fmt.Errorf("async adoption inventory checksum differs")
	}
	return nil
}

func (value Inventory) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire Inventory
	return json.Marshal(wire(value))
}

func DecodeInventory(document []byte) (Inventory, error) {
	var value Inventory
	if err := decodeStrict(document, &value); err != nil {
		return Inventory{}, err
	}
	for index := range value.Items {
		value.Items[index].TargetEffectKey = cloneString(value.Items[index].TargetEffectKey)
	}
	return value, value.Validate()
}

func itemLess(left Item, right Item) bool {
	if left.InventoryClass != right.InventoryClass {
		return left.InventoryClass < right.InventoryClass
	}
	if left.SourceName != right.SourceName {
		return left.SourceName < right.SourceName
	}
	return left.SourceItemID < right.SourceItemID
}
