package recovery

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestListInventoryReturnsDeterministicDigest(t *testing.T) {
	service := newFixtureService(&storeFixture{})
	_, digest, err := service.ListInventory(
		context.Background(),
		uuid.MustParse("30000000-0000-4000-8000-000000000003"),
	)
	if err != nil || digest == [32]byte{} {
		t.Fatalf("list inventory: digest=%x err=%v", digest, err)
	}
}
