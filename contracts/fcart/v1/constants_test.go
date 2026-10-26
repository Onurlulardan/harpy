package fcartv1

import "testing"

func TestWireLimitsArePinned(t *testing.T) {
	if Version != "FCART1" || ChunkPlaintextBytes != 8_388_608 ||
		MaxCiphertextBytes != 2_147_483_648 || DEKBytes != 32 ||
		NonceBytes != 12 || TagBytes != 16 || HeaderBytes != 35 ||
		ChunkFramingBytes != 40 {
		t.Fatal("FCART1 wire constants changed")
	}
}
