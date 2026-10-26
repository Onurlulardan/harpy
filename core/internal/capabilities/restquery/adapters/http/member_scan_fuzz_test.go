package httpquery

import "testing"

func FuzzScanJSONMembersNeverPanics(f *testing.F) {
	f.Add([]byte(`{"a":1,"a":2}`))
	f.Add([]byte(`{"items":[{"x":1}]}`))
	f.Add([]byte(`null true`))
	f.Add([]byte{0xff})
	f.Fuzz(func(t *testing.T, document []byte) {
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("scanJSONMembers panicked: %v", recovered)
			}
		}()
		scan, err := scanJSONMembers(document)
		if err == nil {
			first := scan.duplicatePaths()
			if len(first) != 0 {
				first[0] = "/mutated"
				_ = scan.duplicatePaths()
			}
		}
	})
}
