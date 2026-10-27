package catalogpackagev1

import "testing"

func TestValidatePayloadTextAcceptsTurkishUTF8(t *testing.T) {
	for _, document := range [][]byte{
		[]byte(`{"title":"Eğitim Programı"}`),
		[]byte(`{"title":"Älg på Åland"}`),
	} {
		if err := ValidatePayloadText(document); err != nil {
			t.Fatalf("valid UTF-8 %q rejected: %v", document, err)
		}
	}
}

func TestValidatePayloadTextRejectsLegacyMojibakeMarkers(t *testing.T) {
	for _, document := range [][]byte{
		[]byte(`{"title":"EÄŸitim"}`),
		[]byte(`{"title":"GÃ¶rev"}`),
		[]byte(`{"title":"BaÅŸvuru"}`),
		[]byte("{\"title\":\"\xef\xbf\xbd\"}"),
		{0xff},
	} {
		if err := ValidatePayloadText(document); err == nil {
			t.Fatalf("legacy payload %q was accepted", document)
		}
	}
}
