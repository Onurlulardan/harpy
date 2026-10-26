package domain

import (
	"time"
)

type Filter struct {
	OwnerModule, EntityType, EntityID, FormSlug string
	Action                                      Action
	ActorID, OrganizationID                     *ID
	CorrelationID                               *ID
	Start, End                                  *time.Time
	Search                                      string
	Page, Limit                                 int
}

type Page struct {
	Entries                        []Entry
	Page, Limit, Total, TotalPages int
}

type ActorStat struct {
	ActorID ID
	Count   int64
}

type EntityStat struct {
	EntityType, EntityID, FormSlug string
	Count                          int64
}

type Stats struct {
	Total, Last24Hours, LastHour int64
	ByAction, ByEntity           map[string]int64
	Actors                       []ActorStat
	Entities                     []EntityStat
}
