package installationdelivery

import (
	"testing"

	"github.com/google/uuid"
)

func TestInitialAttemptIDIsStableAndPurposeBound(t *testing.T) {
	operation := uuid.MustParse("019bfcc1-8362-7a01-8757-6c7046864c89")
	credential := uuid.MustParse("019bfcc1-8362-7a02-8757-6c7046864c89")
	first, err := InitialAttemptID(operation, credential)
	second, secondErr := InitialAttemptID(operation, credential)
	if err != nil || secondErr != nil || first != second || first.Version() != uuid.Version(5) {
		t.Fatalf("attempt ids %s/%s err=%v/%v", first, second, err, secondErr)
	}
}
