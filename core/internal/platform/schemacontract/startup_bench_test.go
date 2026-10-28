package schemacontract

import "testing"

func BenchmarkLoadBoundStartup(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		if _, err := LoadBoundStartup(); err != nil {
			b.Fatal(err)
		}
	}
}
