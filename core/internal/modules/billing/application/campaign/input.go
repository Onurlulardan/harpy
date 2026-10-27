package campaign

import (
	"time"

	campaigndomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/campaign"
	"github.com/google/uuid"
)

type PublicationInput struct {
	Name               string
	Description        string
	BannerText         string
	AutoApply          bool
	Available          bool
	Priority           int32
	StartsAt           time.Time
	EndsAt             *time.Time
	DiscountVersionIDs []string
}

func ParseKey(value string) (campaigndomain.Key, error) {
	key, err := campaigndomain.NewKey(value)
	if err != nil {
		return "", ErrRequest
	}
	return key, nil
}

func BuildPublication(input PublicationInput) (campaigndomain.Publication, error) {
	discounts, err := buildInputDiscountIDs(input.DiscountVersionIDs)
	if err != nil {
		return campaigndomain.Publication{}, err
	}
	publication, err := campaigndomain.NewPublication(
		input.Name, input.Description, input.BannerText, input.AutoApply,
		input.Priority, input.StartsAt, input.EndsAt, discounts,
	)
	if err != nil {
		return campaigndomain.Publication{}, ErrRequest
	}
	return publication, nil
}

func buildInputDiscountIDs(values []string) ([]campaigndomain.Identifier, error) {
	result := make([]campaigndomain.Identifier, len(values))
	for index, value := range values {
		identifier, err := uuid.Parse(value)
		if err != nil || identifier == uuid.Nil || identifier.Version() != uuid.Version(7) {
			return nil, ErrRequest
		}
		result[index] = campaigndomain.Identifier(identifier)
	}
	return result, nil
}
