package recoverykitv1

import "fmt"

func Parse(document []byte) (Kit, error) {
	var wire plaintextWire
	if err := parseCanonical(document, &wire); err != nil || !validatePlaintext(wire) {
		return Kit{}, fmt.Errorf("environment recovery kit is invalid")
	}
	return Kit{wire: wire, canonical: append([]byte(nil), document...)}, nil
}
