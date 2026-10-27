package postgres

import (
	"strings"
	"testing"
)

func TestInstallationLayoutRepairPublishesImmutableCuratedSuccessors(t *testing.T) {
	repair := string(readLayoutMigration(
		t,
		"20260730120104_repair_installation_layout_publications_v1",
	))
	for _, marker := range []string{
		"system_resource_layout_installation_catalog_valid_v2(",
		"system_resource_layout_installation_baseline_evidence_valid_v2(",
		"system_resource_layout_repair_installation_defaults_v2(",
		"JOIN system_resource_layout_installation_baselines baseline",
		"WHERE publication.generation=1",
		"INSERT INTO system_resource_layout_revisions",
		"SET state='published'",
		"generation=current_record.generation+1",
		"catalog_json=current_record.catalog_json",
		"catalog_sha256=current_record.catalog_sha256",
		"compiler_version='system-resource-presentation-compiler.v2'",
		"UPDATE system_resource_layout_publications AS publication",
		"SET revision_id=v_revision_id",
		"AND publication.revision_id=current_record.revision_id",
		"SELECT system_resource_layout_repair_installation_defaults_v2(",
	} {
		if !strings.Contains(repair, marker) {
			t.Fatalf("installation layout repair lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"UPDATE system_resource_layout_revisions\n        SET definition_json",
		"system_resource_layout_successor_catalog",
		"c1000000-0000-4000-8000-00000000000c",
		"jsonb_set(",
		"AND revision_id=current_record.revision_id",
	} {
		if strings.Contains(repair, forbidden) {
			t.Fatalf("installation layout repair contains unsafe %q", forbidden)
		}
	}
}

func TestInstallationLayoutValidatorV2IsOrderIndependentAndExact(t *testing.T) {
	validator := string(readLayoutMigration(
		t,
		"20260730120104_repair_installation_layout_publications_v1",
	))
	for _, marker := range []string{
		"LEFT JOIN expected ON expected.id = field.value->>'id'",
		"v_static = v_static_count AND v_custom = 1",
		"pg_catalog.count(DISTINCT id)",
		"id LIKE v_prefix||'%'",
		"7[0-9a-f]{3}-[89ab]",
		"value->>'type_id' IS DISTINCT FROM type_id",
		"value->>'mode' IS DISTINCT FROM CASE",
	} {
		if !strings.Contains(validator, marker) {
			t.Fatalf("installation layout validator v2 lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"WITH ORDINALITY", "item.ordinality", "v_static_count + 1 THEN 'text'",
	} {
		if strings.Contains(validator, forbidden) {
			t.Fatalf("installation layout validator v2 is order-coupled by %q",
				forbidden)
		}
	}
}

func TestInstallationLayoutRepairValidatesExactPredecessorEvidence(t *testing.T) {
	validator := string(readLayoutMigration(
		t,
		"20260730120103_create_layout_installation_evidence_validator_v1",
	))
	for _, marker := range []string{
		"system_resource_layout_installation_catalog_valid_v1(",
		"system_resource_layout_installation_baseline_evidence_valid_v1(",
		"'installation-baseline-evidence.v1'",
		"publication.generation = 1",
		"publication.catalog_sha256 IS DISTINCT FROM pg_catalog.sha256",
		"publication.source_definition_sha256 IS DISTINCT FROM",
		"baseline.baseline_sha256 IS DISTINCT FROM pg_catalog.sha256",
	} {
		if !strings.Contains(validator, marker) {
			t.Fatalf("installation layout validator lacks %q", marker)
		}
	}
}

func TestCuratedDefinitionCarriesHumanLabelsAndRelationFields(t *testing.T) {
	definition := string(readLayoutMigration(
		t,
		"20260730120101_create_layout_curated_definition_v1",
	))
	for _, marker := range []string{
		"'a1000000-0000-4000-8000-000000000006','Roles','Roller'",
		"'a1000000-0000-4000-8000-000000000007','Organizations','Organizasyonlar'",
		"'b1000000-0000-4000-8000-000000000004','Organization','Organizasyon'",
		"'c1000000-0000-4000-8000-000000000005','Parent organization','Üst organizasyon'",
		"'c1000000-0000-4000-8000-00000000000c','Owners','Sahipler'",
	} {
		if !strings.Contains(definition, marker) {
			t.Fatalf("curated definition lacks %q", marker)
		}
	}
}

func TestCuratedParitySuccessorIsExactAndPreservesCustomLayouts(t *testing.T) {
	definition := string(readLayoutMigration(
		t, "20260802232100_create_layout_curated_definition_v2",
	))
	for _, marker := range []string{
		"system_resource_layout_curated_definition_v2(",
		"WHEN 'role:record_create'",
		"'b1000000-0000-4000-8000-000000000009'",
		"WHEN 'organization:record_create'",
		"WHEN 'organization:record_update'",
		"'c1000000-0000-4000-8000-00000000000c'",
	} {
		if !strings.Contains(definition, marker) {
			t.Fatalf("curated parity definition lacks %q", marker)
		}
	}
	userCreate := definition[strings.Index(definition,
		"WHEN 'user:record_create'"):strings.Index(definition,
		"WHEN 'user:record_update'")]
	if strings.Contains(userCreate, "a1000000-0000-4000-8000-000000000003") {
		t.Fatal("curated user create still exposes non-mutation portal identifier")
	}

	evidence := string(readLayoutMigration(
		t, "20260802232101_create_layout_curated_successor_evidence_v1",
	))
	for _, marker := range []string{
		"system_resource_layout_curated_successor_catalog_v1(",
		"system_resource_layout_curated_predecessor_valid_v1(",
		"v_publication.generation<>2",
		"v_revision.definition_json IS DISTINCT FROM v_definition",
		"v_publication.presentation_json IS NOT DISTINCT FROM v_presentation",
		"'mode','editable','type_id','uuid_list'",
	} {
		if !strings.Contains(evidence, marker) {
			t.Fatalf("curated parity evidence lacks %q", marker)
		}
	}

	repair := string(readLayoutMigration(
		t, "20260802232102_repair_layout_curated_parity_v2",
	))
	for _, marker := range []string{
		"system_resource_layout_repair_curated_defaults_v3(",
		"WHERE publication.generation=2",
		"IF NOT system_resource_layout_curated_predecessor_valid_v1(",
		"INSERT INTO system_resource_layout_revisions",
		"generation=v_current.generation+1",
		"catalog_json=v_catalog,catalog_sha256=v_catalog_sha",
		"AND publication.revision_id=v_current.revision_id",
		"'curated-system-resource-parity-successor.v2'",
	} {
		if !strings.Contains(repair, marker) {
			t.Fatalf("curated parity repair lacks %q", marker)
		}
	}
}
