package money

import (
	"reflect"
	"sort"
	"testing"
)

func TestCurrencyRegistryV1PinnedProvenance(t *testing.T) {
	metadata, ok := CurrencyRegistryMetadataForVersion(1)
	if !ok {
		t.Fatal("registry v1 metadata is missing")
	}
	const sourceURL = "https://www.six-group.com/dam/download/financial-information/data-center/iso-currrency/lists/list-one.xml"
	if metadata.SourceURL != sourceURL || metadata.SourcePublished != "2026-01-01" ||
		metadata.SourceBytes != 47463 {
		t.Fatalf("source date/bytes = %s/%d", metadata.SourcePublished, metadata.SourceBytes)
	}
	const sourceSHA = "838dfb991648cf36df939edd5fe3811737962b75a32252847d239cedd1e291c9"
	const normalizedSHA = "894f71b85f203d76ef7dacb379caf4340638d7ce7c2a2ffafd0b5f65c1b6714f"
	if metadata.SourceSHA256 != sourceSHA || metadata.UniqueCodes != 178 ||
		metadata.NormalizedSHA256 != normalizedSHA || metadata.NotApplicableCodes != 13 {
		t.Fatalf("source provenance = %+v", metadata)
	}
	if _, ok := CurrencyRegistryMetadataForVersion(2); ok {
		t.Fatal("unknown registry version was accepted")
	}
}

func TestCurrencyRegistryV1ExponentDistribution(t *testing.T) {
	entries, ok := CurrencyEntries(1)
	if !ok || len(entries) != 178 {
		t.Fatalf("entries = %d, ok = %v", len(entries), ok)
	}
	distribution := map[uint8]int{}
	var notApplicable []string
	for index, entry := range entries {
		if index > 0 && entries[index-1].Code >= entry.Code {
			t.Fatalf("entries are not unique and sorted at %s", entry.Code)
		}
		if entry.MinorUnitApplicable {
			distribution[entry.Exponent]++
		} else {
			notApplicable = append(notApplicable, entry.Code)
		}
	}
	if want := map[uint8]int{0: 17, 2: 139, 3: 7, 4: 2}; !reflect.DeepEqual(distribution, want) {
		t.Fatalf("distribution = %v, want %v", distribution, want)
	}
	wantNA := []string{
		"XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XDR",
		"XPD", "XPT", "XSU", "XTS", "XUA", "XXX",
	}
	if !reflect.DeepEqual(notApplicable, wantNA) {
		t.Fatalf("N.A. codes = %v, want %v", notApplicable, wantNA)
	}
}

func TestCurrencyRegistryLookupIsExactAndImmutable(t *testing.T) {
	entry, ok := LookupCurrency(1, "USD")
	if !ok || entry.Exponent != 2 || !entry.MinorUnitApplicable {
		t.Fatalf("USD = %+v, ok = %v", entry, ok)
	}
	for _, code := range []string{"usd", " USD", "USD ", "ZZZ"} {
		if _, ok := LookupCurrency(1, code); ok {
			t.Fatalf("invalid code %q was accepted", code)
		}
	}
	entries, _ := CurrencyEntries(1)
	entries[0].Code = "ZZZ"
	again, _ := CurrencyEntries(1)
	if !sort.SliceIsSorted(again, func(i, j int) bool {
		return again[i].Code < again[j].Code
	}) || again[0].Code != "AED" {
		t.Fatal("caller mutated the registry snapshot")
	}
}
