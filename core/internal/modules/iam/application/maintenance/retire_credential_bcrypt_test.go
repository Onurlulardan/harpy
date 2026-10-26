package maintenance

import "testing"

func TestExecuteCredentialBcryptRetireRoutesClosedKind(t *testing.T) {
	executeFixture(t, KindCredentialBcryptRetire)
	if t.Failed() {
		t.FailNow()
	}
}
