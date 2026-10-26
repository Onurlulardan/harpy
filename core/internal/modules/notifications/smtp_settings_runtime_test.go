package notifications

import (
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func TestPlatformSMTPSettingsPortFailsClosedWithoutCurrentKEK(t *testing.T) {
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewPlatformSMTPSettingsPort(PlatformSMTPSettingsDependencies{
		Catalogs: catalogs, KEKKeys: map[int64][32]byte{1: {1}},
		CurrentKEKRevision: 2})
	if err == nil {
		t.Fatal("SMTP settings runtime accepted missing current KEK")
	}
}

func TestPlatformSMTPSettingsPortBuildsWithReleaseCatalogAndCurrentKEK(t *testing.T) {
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	key := [32]byte{1, 2, 3}
	port, err := NewPlatformSMTPSettingsPort(PlatformSMTPSettingsDependencies{
		Catalogs: catalogs, KEKKeys: map[int64][32]byte{2: key},
		CurrentKEKRevision: 2})
	if err != nil || port == nil {
		t.Fatalf("SMTP settings port = %v, %v", port, err)
	}
}
