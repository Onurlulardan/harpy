package contract

import (
	"errors"
)

var errQueryDocumentInvalid = errors.New("REST Query document is invalid")

type QueryDocumentInput struct {
	Select    RawMember
	Where     RawMember
	Sort      RawMember
	Page      RawMember
	Search    RawMember
	Count     RawMember
	Include   RawMember
	Aggregate RawMember
}

type QueryDocument struct {
	selectMember    RawMember
	whereMember     RawMember
	sortMember      RawMember
	pageMember      RawMember
	searchMember    RawMember
	countMember     RawMember
	includeMember   RawMember
	aggregateMember RawMember
	valid           bool
}

func NewQueryDocument(input QueryDocumentInput) (QueryDocument, error) {
	if err := validateDocumentInput(input); err != nil {
		return QueryDocument{}, err
	}
	return QueryDocument{
		selectMember:    input.Select.clone(),
		whereMember:     input.Where.clone(),
		sortMember:      input.Sort.clone(),
		pageMember:      input.Page.clone(),
		searchMember:    input.Search.clone(),
		countMember:     input.Count.clone(),
		includeMember:   input.Include.clone(),
		aggregateMember: input.Aggregate.clone(),
		valid:           true,
	}, nil
}

func (document QueryDocument) Validate() error {
	if !document.valid {
		return errQueryDocumentInvalid
	}
	return validateDocumentInput(QueryDocumentInput{
		Select: document.selectMember, Aggregate: document.aggregateMember,
		Where: document.whereMember, Sort: document.sortMember,
		Page: document.pageMember, Search: document.searchMember,
		Count: document.countMember, Include: document.includeMember,
	})
}

func validateDocumentInput(input QueryDocumentInput) error {
	members := [...]RawMember{
		input.Select,
		input.Where,
		input.Sort,
		input.Page,
		input.Search,
		input.Count,
		input.Include,
		input.Aggregate,
	}
	for _, member := range members {
		if err := member.Validate(); err != nil {
			return errQueryDocumentInvalid
		}
	}
	return nil
}
