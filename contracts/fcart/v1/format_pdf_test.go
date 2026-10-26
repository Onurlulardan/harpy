package fcartv1

import "testing"

func TestPDFFormatUsesAdditiveWireCodeFour(t *testing.T) {
	code, err := FormatPDF.wireCode()
	if err != nil || code != 4 {
		t.Fatalf("PDF wire code = %d, %v", code, err)
	}
	decoded, err := formatFromWire(4)
	if err != nil || decoded != FormatPDF {
		t.Fatalf("PDF wire decode = %q, %v", decoded, err)
	}
}
