package async

import "testing"

func TestPhaseThreeCatalogStartsEmptyButValid(t *testing.T) {
	catalogs, err := NewEventCatalogs(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalogs.Producers().Policies()) != 0 ||
		len(catalogs.Consumers().Policies()) != 0 {
		t.Fatal("platform composition invented a business event policy")
	}
}
