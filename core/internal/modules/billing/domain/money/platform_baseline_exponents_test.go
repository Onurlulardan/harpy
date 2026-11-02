package money

import (
	"os"
	"regexp"
	"strconv"
	"testing"
)

func TestPlatformBaselineCarriesCurrencyRegistryExponents(t *testing.T) {
	data, err := os.ReadFile(
		"../../../../platform/baseline/sql/20261028000000_platform_baseline.fresh.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	entries, ok := CurrencyEntries(RegistryVersion)
	if !ok || len(entries) == 0 {
		t.Fatal("currency registry has no entries")
	}
	pattern := regexp.MustCompile(
		`INSERT INTO platform\.billing_currency_exponents VALUES \(` +
			`'[^']+', ([0-9]+), '([A-Z]{3})', (?:NULL|'[0-9]{3}'), (NULL|[0-9]+), (true|false)\);`,
	)
	found := map[string]struct {
		exponent *int
		eligible bool
	}{}
	for _, match := range pattern.FindAllSubmatch(data, -1) {
		version, err := strconv.ParseInt(string(match[1]), 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		if version != RegistryVersion {
			t.Fatalf("baseline currency %s uses registry version %d", match[2], version)
		}
		code := string(match[2])
		row := struct {
			exponent *int
			eligible bool
		}{eligible: string(match[4]) == "true"}
		if string(match[3]) != "NULL" {
			exponent, err := strconv.Atoi(string(match[3]))
			if err != nil {
				t.Fatal(err)
			}
			row.exponent = &exponent
		}
		if _, exists := found[code]; exists {
			t.Fatalf("baseline repeats currency %s", code)
		}
		found[code] = row
	}
	if len(found) != len(entries) {
		t.Fatalf("baseline currencies = %d, registry = %d", len(found), len(entries))
	}
	for _, entry := range entries {
		row, exists := found[entry.Code]
		if !exists {
			t.Fatalf("baseline omits currency %s", entry.Code)
		}
		if entry.MinorUnitApplicable {
			if row.exponent == nil || *row.exponent != int(entry.Exponent) || !row.eligible {
				t.Fatalf("baseline exponent for %s = %v eligible %t", entry.Code, row.exponent, row.eligible)
			}
			continue
		}
		if row.exponent != nil || row.eligible {
			t.Fatalf("baseline keeps a minor unit for %s", entry.Code)
		}
	}
}
