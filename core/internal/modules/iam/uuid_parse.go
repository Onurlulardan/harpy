package iam

import "github.com/google/uuid"

func ParseUUID(value string) (UUID, error) {
	return uuid.Parse(value)
}
