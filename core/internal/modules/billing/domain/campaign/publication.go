package campaign

import (
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maximumNameBytes        = 128
	maximumBannerBytes      = 512
	maximumDescriptionBytes = 4096
	maximumDiscountVersion  = 256
)

type Publication struct {
	name               string
	description        string
	bannerText         string
	autoApply          bool
	priority           int32
	startsAt           time.Time
	endsAt             *time.Time
	discountVersionIDs []Identifier
}

func NewPublication(
	name string,
	description string,
	bannerText string,
	autoApply bool,
	priority int32,
	startsAt time.Time,
	endsAt *time.Time,
	discountVersionIDs []Identifier,
) (Publication, error) {
	start, err := canonicalInstant(startsAt)
	if err != nil || !validName(name) || !validDescription(description) ||
		!validBanner(bannerText) {
		return Publication{}, ErrPublication
	}
	end, err := canonicalEnd(endsAt, start)
	identifiers, valid := canonicalIdentifiers(discountVersionIDs)
	if err != nil || !valid {
		return Publication{}, ErrPublication
	}
	return Publication{name: name, description: description, bannerText: bannerText,
		autoApply: autoApply, priority: priority,
		startsAt: start, endsAt: end, discountVersionIDs: identifiers}, nil
}

func (publication Publication) Name() string        { return publication.name }
func (publication Publication) Description() string { return publication.description }
func (publication Publication) BannerText() string  { return publication.bannerText }
func (publication Publication) AutoApply() bool     { return publication.autoApply }
func (publication Publication) Priority() int32     { return publication.priority }
func (publication Publication) StartsAt() time.Time { return publication.startsAt }

func (publication Publication) EndsAt() *time.Time {
	if publication.endsAt == nil {
		return nil
	}
	value := *publication.endsAt
	return &value
}

func (publication Publication) DiscountVersionIDs() []Identifier {
	result := make([]Identifier, len(publication.discountVersionIDs))
	copy(result, publication.discountVersionIDs)
	return result
}

func (publication Publication) Valid() bool {
	start, err := canonicalInstant(publication.startsAt)
	if err != nil || !validName(publication.name) ||
		!validDescription(publication.description) || !validBanner(publication.bannerText) {
		return false
	}
	if _, err = canonicalEnd(publication.endsAt, start); err != nil {
		return false
	}
	_, valid := canonicalIdentifiers(publication.discountVersionIDs)
	return valid
}

func validDescription(value string) bool {
	return len(value) <= maximumDescriptionBytes && utf8.ValidString(value)
}

func validName(value string) bool {
	return value != "" && len(value) <= maximumNameBytes && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value
}

func validBanner(value string) bool {
	return len(value) <= maximumBannerBytes && utf8.ValidString(value)
}

func canonicalIdentifiers(values []Identifier) ([]Identifier, bool) {
	if len(values) > maximumDiscountVersion {
		return nil, false
	}
	result := make([]Identifier, len(values))
	copy(result, values)
	sort.Slice(result, func(left, right int) bool {
		return compareIdentifier(result[left], result[right]) < 0
	})
	for index, identifier := range result {
		if !validIdentifier(identifier) ||
			(index > 0 && compareIdentifier(result[index-1], identifier) == 0) {
			return nil, false
		}
	}
	return result, true
}
