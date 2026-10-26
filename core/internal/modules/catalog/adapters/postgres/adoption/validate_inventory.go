package adoption

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"sort"
	"unicode/utf8"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/adoption"
)

//go:embed sql/tenant/inventory_rows.sql
var inventoryRowsSQL string

//go:embed sql/tenant/list_blockers.sql
var listBlockersSQL string

func (storage *store) ValidateInventory(ctx context.Context) (adoptionapp.Inventory, error) {
	if !storage.usable(ctx, false) {
		return adoptionapp.Inventory{}, unavailable()
	}
	resultRows, err := storage.query(ctx, inventoryRowsSQL, storage.tenantID)
	if err != nil {
		return adoptionapp.Inventory{}, unavailable()
	}
	defer resultRows.Close()
	source, target := make([]inventoryRow, 0), make([]inventoryRow, 0)
	for resultRows.Next() {
		var item inventoryRow
		if err = resultRows.Scan(&item.side, &item.kind, &item.identity, &item.document); err != nil ||
			!item.valid() {
			return adoptionapp.Inventory{}, unavailable()
		}
		if item.side == "source" {
			source = append(source, item)
		} else {
			target = append(target, item)
		}
	}
	if resultRows.Err() != nil {
		return adoptionapp.Inventory{}, unavailable()
	}
	blockers, err := storage.blockers(ctx)
	if err != nil {
		return adoptionapp.Inventory{}, err
	}
	return adoptionapp.Inventory{SourceCount: int64(len(source)), TargetCount: int64(len(target)),
		SourceRootSHA256: inventoryRoot(source), TargetRootSHA256: inventoryRoot(target),
		Blockers: blockers}, nil
}

type inventoryRow struct{ side, kind, identity, document string }

func (value inventoryRow) valid() bool {
	return (value.side == "source" || value.side == "target") && value.kind != "" &&
		value.identity != "" && len(value.document) >= 2 && len(value.document) <= 4<<20 &&
		utf8.ValidString(value.document)
}

func inventoryRoot(rows []inventoryRow) string {
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].kind+"\x00"+rows[i].identity < rows[j].kind+"\x00"+rows[j].identity
	})
	hash := sha256.New()
	for _, item := range rows {
		for _, field := range []string{item.kind, item.identity, item.document} {
			_ = binary.Write(hash, binary.BigEndian, uint32(len(field)))
			_, _ = hash.Write([]byte(field))
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func (storage *store) blockers(ctx context.Context) ([]string, error) {
	resultRows, err := storage.query(ctx, listBlockersSQL, storage.tenantID)
	if err != nil {
		return nil, unavailable()
	}
	defer resultRows.Close()
	result := make([]string, 0)
	for resultRows.Next() {
		var blocker string
		if err = resultRows.Scan(&blocker); err != nil || blocker == "" || len(blocker) > 256 {
			return nil, unavailable()
		}
		result = append(result, blocker)
	}
	if resultRows.Err() != nil {
		return nil, unavailable()
	}
	return result, nil
}
