package maintenance

import "testing"

func TestExecuteDirectoryChecksumRoutesClosedKind(t *testing.T) {
	executeFixture(t, KindDirectoryChecksum)
	if t.Failed() {
		t.FailNow()
	}
}
