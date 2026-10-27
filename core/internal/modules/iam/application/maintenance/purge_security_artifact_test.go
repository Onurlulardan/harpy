package maintenance

import "testing"

func TestExecuteSecurityArtifactPurgeRoutesClosedKind(t *testing.T) {
	executeFixture(t, KindSecurityArtifactPurge)
	if t.Failed() {
		t.FailNow()
	}
}
