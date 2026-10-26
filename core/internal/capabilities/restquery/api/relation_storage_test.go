package api

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/physicalid"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

func TestRelationStorageCanonicalRoundTrip(t *testing.T) {
	storage := mustRelationStorage(t, validRelationID)
	table, tableErr := storage.Table()
	fingerprint, fingerprintErr := storage.FingerprintSHA256()
	name, _ := physicalid.Name("rel_", uuid.MustParse(validRelationID))
	if tableErr != nil || fingerprintErr != nil || fingerprint == [32]byte{} ||
		!sqlidentifier.Matches(table, name) {
		t.Fatal("canonical relation storage did not round trip")
	}
}

func TestRelationStorageRejectsMismatchedMappingAndTransplant(t *testing.T) {
	otherID := "dddddddd-dddd-4ddd-8ddd-dddddddddddd"
	storage := mustRelationStorage(t, validRelationID)
	table, _ := storage.Table()
	if _, err := NewRelationStorage(otherID, table); err == nil {
		t.Fatal("relation table mapped to a different relation ID")
	}
	if _, err := NewRelationDefinition(
		otherID, "other", RelationCardinalityToOne, storage,
		mustRelationTarget(t), true, 1,
	); err == nil {
		t.Fatal("relation storage transplanted into another definition")
	}
	static, _ := sqlidentifier.MintStatic("forms_relations")
	if _, err := NewRelationStorage(validRelationID, static); err == nil {
		t.Fatal("non-canonical relation table accepted")
	}
}

func TestRelationStorageZeroAndTamperFailClosed(t *testing.T) {
	valid := mustRelationStorage(t, validRelationID)
	wrongTable := mustRelationStorage(
		t, "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
	)
	tamperedTable := valid
	tamperedTable.table = wrongTable.table
	tamperedID := valid
	tamperedID.relationID = uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd")
	tamperedSeal := valid
	tamperedSeal.stateSeal[0] ^= 1
	unvalidated := valid
	unvalidated.validated = false
	for _, storage := range []RelationStorage{
		{}, tamperedTable, tamperedID, tamperedSeal, unvalidated,
	} {
		if storage.Validate() == nil {
			t.Fatal("invalid relation storage validated")
		}
		if _, err := storage.Table(); err == nil {
			t.Fatal("invalid relation storage exposed its table")
		}
		if _, err := storage.FingerprintSHA256(); err == nil {
			t.Fatal("invalid relation storage exposed its fingerprint")
		}
	}
}

func FuzzRelationStorageBindsCanonicalTableToExactID(f *testing.F) {
	f.Add([]byte("0123456789abcdef"))
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) != 16 {
			return
		}
		raw = append([]byte{}, raw...)
		raw[8] = (raw[8] & 0x3f) | 0x80
		if allZero(raw) {
			raw[15] = 1
		}
		identifier, err := uuid.FromBytes(raw)
		if err != nil {
			return
		}
		name, err := physicalid.Name("rel_", identifier)
		if err != nil {
			return
		}
		table, err := sqlidentifier.MintDynamic(name)
		if err != nil {
			t.Fatal(err)
		}
		storage, err := NewRelationStorage(identifier.String(), table)
		if err != nil || storage.Validate() != nil {
			t.Fatal("canonical fuzz relation mapping rejected")
		}
		other := identifier
		other[15] ^= 1
		if other == uuid.Nil {
			other[15] = 2
		}
		if _, err := NewRelationStorage(other.String(), table); err == nil {
			t.Fatal("cross-relation fuzz storage transplant accepted")
		}
	})
}

func allZero(value []byte) bool {
	for _, current := range value {
		if current != 0 {
			return false
		}
	}
	return true
}
