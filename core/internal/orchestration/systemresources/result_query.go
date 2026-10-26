package systemresources

import publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"

type QueryResult struct {
	query        publicresource.QueryResult
	presentation Presentation
	iamOnly      bool
	validated    bool
}

func NewQueryResult(
	query publicresource.QueryResult,
	presentation Presentation,
) (QueryResult, error) {
	result := QueryResult{query: query, presentation: presentation, validated: true}
	if result.Validate() != nil {
		return QueryResult{}, ErrContract
	}
	return result, nil
}

func newIAMOnlyQueryResult(query publicresource.QueryResult) (QueryResult, error) {
	result := QueryResult{query: query, iamOnly: true, validated: true}
	if result.Validate() != nil {
		return QueryResult{}, ErrContract
	}
	return result, nil
}

func (result QueryResult) Validate() error {
	if !result.validated || result.query.Validate() != nil {
		return ErrContract
	}
	if result.iamOnly {
		return nil
	}
	if result.presentation.Validate() != nil {
		return ErrContract
	}
	queryCatalog, err := result.query.Catalog()
	presentationCatalog, presentationErr := result.presentation.Catalog()
	if err != nil || presentationErr != nil {
		return ErrContract
	}
	equal, err := queryCatalog.ExactEqual(presentationCatalog)
	if err != nil || !equal {
		return ErrContract
	}
	return nil
}

func (result QueryResult) IAM() (publicresource.QueryResult, error) {
	if result.Validate() != nil {
		return publicresource.QueryResult{}, ErrContract
	}
	return result.query, nil
}

func (result QueryResult) Presentation() (Presentation, error) {
	if result.iamOnly {
		return Presentation{}, ErrUnavailable
	}
	if result.Validate() != nil {
		return Presentation{}, ErrContract
	}
	return result.presentation, nil
}
