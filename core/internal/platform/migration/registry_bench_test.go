package migration

import "testing"

func BenchmarkLoadReleaseRegistry(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		registry, err := LoadReleaseRegistry()
		if err != nil {
			b.Fatal(err)
		}
		if len(registry.Digests()) == 0 {
			b.Fatal("release registry is empty")
		}
	}
}
