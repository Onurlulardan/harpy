package lifecycle

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"sort"

	"github.com/google/uuid"
)

func sealDemoRemovalPlan(
	installationID uuid.UUID,
	components []DemoRemovalComponent,
) (DemoRemovalPlan, error) {
	if installationID.Version() != 7 || len(components) == 0 {
		return DemoRemovalPlan{}, ErrInvalid
	}
	sealed := make([]DemoRemovalComponent, len(components))
	total := 0
	for index, component := range components {
		if component.ComponentID.Version() != 7 || component.SourceRunID.Version() != 7 {
			return DemoRemovalPlan{}, ErrInvalid
		}
		ids, err := canonicalDemoRecordIDs(component.RecordIDs)
		if err != nil || len(ids) == 0 {
			return DemoRemovalPlan{}, ErrInvalid
		}
		sealed[index] = DemoRemovalComponent{
			ComponentID: component.ComponentID,
			SourceRunID: component.SourceRunID,
			RecordIDs:   ids,
		}
		total += len(ids)
	}
	if total > 10000 {
		return DemoRemovalPlan{}, ErrInvalid
	}
	sort.Slice(sealed, func(i, j int) bool {
		return sealed[i].ComponentID.String() < sealed[j].ComponentID.String()
	})
	for index := 1; index < len(sealed); index++ {
		if sealed[index-1].ComponentID == sealed[index].ComponentID {
			return DemoRemovalPlan{}, ErrInvalid
		}
	}
	var document bytes.Buffer
	writeDemoField(&document, "catalog-demo-removal-plan.v1")
	writeDemoField(&document, installationID.String())
	for _, component := range sealed {
		writeDemoField(&document, component.ComponentID.String())
		writeDemoField(&document, component.SourceRunID.String())
		for _, recordID := range component.RecordIDs {
			writeDemoField(&document, recordID.String())
		}
	}
	digest := sha256.Sum256(document.Bytes())
	return DemoRemovalPlan{InstallationID: installationID, Components: sealed,
		PlanSHA256: hex.EncodeToString(digest[:])}, nil
}

func canonicalDemoRecordIDs(values []uuid.UUID) ([]uuid.UUID, error) {
	result := append([]uuid.UUID(nil), values...)
	for _, value := range result {
		if value == uuid.Nil {
			return nil, ErrInvalid
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})
	for index := 1; index < len(result); index++ {
		if result[index-1] == result[index] {
			return nil, ErrInvalid
		}
	}
	return result, nil
}

func writeDemoField(buffer *bytes.Buffer, value string) {
	_ = binary.Write(buffer, binary.BigEndian, uint32(len(value)))
	buffer.WriteString(value)
}
