package httpquery

import "testing"

func FuzzGETDecoderDoesNotPanic(f *testing.F) {
	for _, seed := range []string{
		"", "count=none", "filter[0][field]=x&filter[0][op]=eq&filter[0][value]=y",
		"search[term]=%FF", "select[01]=x", "a=b&a=c",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, rawQuery string) {
		_, _ = NewGETDecoder().Decode(rawQuery)
	})
}
