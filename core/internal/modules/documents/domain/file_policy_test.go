package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestFilePolicyRequiresExtensionMIMEAndDetectedKindAgreement(t *testing.T) {
	valid := FileDeclaration{Filename: "rapor.pdf", DeclaredMIME: "application/pdf",
		DetectedKind: FilePDF, Size: 42}
	if err := ValidateFile(valid); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []FileDeclaration{
		{Filename: "rapor.pdf", DeclaredMIME: "text/html", DetectedKind: FilePDF, Size: 42},
		{Filename: "rapor.svg", DeclaredMIME: "image/svg+xml", DetectedKind: FilePNG, Size: 42},
		{Filename: "../rapor.pdf", DeclaredMIME: "application/pdf", DetectedKind: FilePDF, Size: 42},
		{Filename: strings.Repeat("a", 256) + ".pdf", DeclaredMIME: "application/pdf", DetectedKind: FilePDF, Size: 42},
	} {
		if err := ValidateFile(invalid); !errors.Is(err, ErrFilePolicy) {
			t.Fatalf("invalid file accepted: %+v, %v", invalid, err)
		}
	}
}

func TestContentObjectKeyUsesOnlyStableIdentities(t *testing.T) {
	tenant := mustID(t, "11111111-1111-4111-8111-111111111111")
	document := mustID(t, "018f0000-0000-7000-8000-000000000002")
	version := mustID(t, "018f0000-0000-7000-8000-000000000003")
	key, err := ContentObjectKey(tenant, document, version)
	if err != nil {
		t.Fatal(err)
	}
	want := "tenants/ceirceircfardairceirceirce/documents/" + document.String() +
		"/versions/" + version.String() + "/content"
	if key != want || strings.Contains(key, "rapor") {
		t.Fatalf("object key = %q", key)
	}
}

func TestArchiveObjectKeyUsesOnlyStableIdentities(t *testing.T) {
	tenant := mustID(t, "11111111-1111-4111-8111-111111111111")
	archive := mustID(t, "018f0000-0000-7000-8000-000000000099")
	key, err := ArchiveObjectKey(tenant, archive)
	want := "tenants/ceirceircfardairceirceirce/document-archives/" +
		archive.String() + "/content"
	if err != nil || key != want {
		t.Fatalf("archive object key = %q, %v; want %q", key, err, want)
	}
}

func TestDeclaredFileKindRejectsExtensionMIMEDisagreementBeforeUpload(t *testing.T) {
	kind, err := DeclaredFileKind("report.xlsx",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", 512)
	if err != nil || kind != FileXLSX {
		t.Fatalf("kind=%s err=%v", kind, err)
	}
	if _, err = DeclaredFileKind("report.xlsx", "application/pdf", 512); err == nil {
		t.Fatal("mismatched declaration accepted")
	}
}
