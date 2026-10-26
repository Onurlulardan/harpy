package compatibility

import "github.com/google/uuid"

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}

func nullableDigest(value [32]byte) any {
	if value == [32]byte{} {
		return nil
	}
	return value[:]
}

func assignUUID(target *uuid.UUID, source *uuid.UUID) {
	if source != nil {
		*target = *source
	}
}
