package lifecycle

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"sort"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func (storage *store) readCurrentMenu(ctx context.Context) ([]menuRecord, error) {
	resultRows, err := storage.query(ctx, readCurrentMenuItemsSQL, menuDefinitionID)
	if err != nil {
		return nil, unavailable()
	}
	defer resultRows.Close()
	result := make([]menuRecord, 0)
	for resultRows.Next() {
		var itemID, packageID, kind, label string
		var parentID, targetID, externalURL *string
		var order int
		if err = resultRows.Scan(&itemID, &parentID, &packageID, &kind,
			&targetID, &externalURL, &label, &order); err != nil {
			return nil, unavailable()
		}
		id, parseErr := packagecontract.ParseMenuItemID(itemID)
		pin := catalogdomain.MenuPin{ItemID: id, Kind: packagecontract.MenuKind(kind),
			Label: label, Order: order, URL: stringValue(externalURL)}
		if parseErr != nil || !decodeMenuReferences(&pin, stringValue(parentID), stringValue(targetID)) {
			return nil, unavailable()
		}
		result = append(result, menuRecord{PackageID: packageID, Pin: pin})
	}
	if resultRows.Err() != nil {
		return nil, unavailable()
	}
	return result, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func decodeMenuReferences(pin *catalogdomain.MenuPin, parentRaw, targetRaw string) bool {
	if parentRaw != "" {
		parent, err := packagecontract.ParseMenuItemID(parentRaw)
		if err != nil {
			return false
		}
		pin.ParentID = &parent
	}
	if targetRaw != "" {
		target, err := packagecontract.ParseComponentID(targetRaw)
		if err != nil {
			return false
		}
		pin.TargetID = &target
	}
	return true
}

func menuDigest(items []menuRecord) string {
	copyItems := append([]menuRecord(nil), items...)
	sort.Slice(copyItems, func(i, j int) bool {
		return copyItems[i].Pin.ItemID.String() < copyItems[j].Pin.ItemID.String()
	})
	var document bytes.Buffer
	for _, item := range copyItems {
		writeMenuField(&document, item.PackageID)
		writeMenuField(&document, item.Pin.ItemID.String())
		if item.Pin.ParentID != nil {
			writeMenuField(&document, item.Pin.ParentID.String())
		} else {
			writeMenuField(&document, "")
		}
		writeMenuField(&document, string(item.Pin.Kind))
		if item.Pin.TargetID != nil {
			writeMenuField(&document, item.Pin.TargetID.String())
		} else {
			writeMenuField(&document, "")
		}
		writeMenuField(&document, item.Pin.Label)
		_ = binary.Write(&document, binary.BigEndian, int64(item.Pin.Order))
		writeMenuField(&document, item.Pin.URL)
	}
	digest := sha256.Sum256(document.Bytes())
	return hex.EncodeToString(digest[:])
}

func writeMenuField(buffer *bytes.Buffer, value string) {
	_ = binary.Write(buffer, binary.BigEndian, uint32(len(value)))
	buffer.WriteString(value)
}
