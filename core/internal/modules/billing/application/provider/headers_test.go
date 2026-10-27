package providerapp

import "testing"

func TestHeadersAreCaseInsensitiveDefensiveAndRejectInjection(t *testing.T) {
	source := map[string][]string{"testpay-signature": {"v1,signed"}}
	headers, err := NewHeaders(source)
	if err != nil {
		t.Fatal(err)
	}
	source["testpay-signature"][0] = "changed"
	values := headers.Values("Testpay-Signature")
	if len(values) != 1 || values[0] != "v1,signed" {
		t.Fatalf("values=%v", values)
	}
	values[0] = "mutated"
	if headers.Values("TESTPAY-SIGNATURE")[0] != "v1,signed" {
		t.Fatal("returned header values alias internal storage")
	}
	for _, invalid := range []map[string][]string{
		{"Bad Header": {"value"}},
		{"Signature": {"value\r\nX-Forged: yes"}},
	} {
		if _, err := NewHeaders(invalid); err != ErrWebhookRequest {
			t.Fatalf("invalid header error=%v", err)
		}
	}
}
