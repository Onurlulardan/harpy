package money

import "sort"

type CurrencyEntry struct {
	Code                string
	Exponent            uint8
	MinorUnitApplicable bool
}

type CurrencyRegistryMetadata struct {
	Version            int64
	SourceURL          string
	SourcePublished    string
	SourceBytes        int64
	SourceSHA256       string
	NormalizedSHA256   string
	UniqueCodes        int
	NotApplicableCodes int
}

func LookupCurrency(version int64, code string) (CurrencyEntry, bool) {
	if version != RegistryVersion {
		return CurrencyEntry{}, false
	}
	index := sort.Search(len(currencyEntries), func(index int) bool {
		return currencyEntries[index].Code >= code
	})
	if index == len(currencyEntries) || currencyEntries[index].Code != code {
		return CurrencyEntry{}, false
	}
	return currencyEntries[index], true
}

func CurrencyEntries(version int64) ([]CurrencyEntry, bool) {
	if version != RegistryVersion {
		return nil, false
	}
	result := make([]CurrencyEntry, len(currencyEntries))
	copy(result, currencyEntries[:])
	return result, true
}

func CurrencyRegistryMetadataForVersion(
	version int64,
) (CurrencyRegistryMetadata, bool) {
	if version != RegistryVersion {
		return CurrencyRegistryMetadata{}, false
	}
	return CurrencyRegistryMetadata{
		Version:            RegistryVersion,
		SourceURL:          currencySourceURL,
		SourcePublished:    currencySourcePublished,
		SourceBytes:        currencySourceBytes,
		SourceSHA256:       currencySourceSHA256,
		NormalizedSHA256:   currencyNormalizedInputSHA256,
		UniqueCodes:        currencyUniqueCodeCount,
		NotApplicableCodes: currencyNotApplicableCodeCount,
	}, true
}
