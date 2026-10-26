package api

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestResourceStorageAcceptsValidatedStaticAndDynamicTokens(t *testing.T) {
	minters := []struct {
		name  string
		value string
		mint  func(string) (sqlidentifier.Token, error)
	}{
		{"static", "forms_records", sqlidentifier.MintStatic},
		{"dynamic", "frm_orders", sqlidentifier.MintDynamic},
	}
	for _, test := range minters {
		t.Run(test.name, func(t *testing.T) {
			token, err := test.mint(test.value)
			if err != nil {
				t.Fatal(err)
			}
			storage, err := NewResourceStorage(token)
			if err != nil || storage.Validate() != nil {
				t.Fatalf("validated token rejected: %v", err)
			}
			actual, err := storage.Table()
			if err != nil || actual != token {
				t.Fatal("resource storage did not return the token copy")
			}
			actual = sqlidentifier.Token{}
			again, err := storage.Table()
			if err != nil || again != token {
				t.Fatal("resource storage token was externally mutated")
			}
		})
	}
}

func TestResourceStorageZeroAndTamperFailClosed(t *testing.T) {
	if _, err := NewResourceStorage(sqlidentifier.Token{}); err == nil {
		t.Fatal("zero table token accepted")
	}
	validToken, _ := sqlidentifier.MintStatic("forms_records")
	for _, storage := range []ResourceStorage{
		{},
		{table: validToken},
		{validated: true},
	} {
		if storage.Validate() == nil {
			t.Fatal("invalid resource storage validated")
		}
		if _, err := storage.Table(); err == nil {
			t.Fatal("invalid resource storage exposed a table token")
		}
	}
}
