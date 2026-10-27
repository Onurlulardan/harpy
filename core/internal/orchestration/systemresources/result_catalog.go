package systemresources

type CatalogResult struct {
	presentation Presentation
	validated    bool
}

func NewCatalogResult(presentation Presentation) (CatalogResult, error) {
	result := CatalogResult{presentation: presentation, validated: true}
	if result.Validate() != nil {
		return CatalogResult{}, ErrContract
	}
	return result, nil
}

func (result CatalogResult) Validate() error {
	if !result.validated || result.presentation.Validate() != nil {
		return ErrContract
	}
	return nil
}

func (result CatalogResult) Presentation() (Presentation, error) {
	if result.Validate() != nil {
		return Presentation{}, ErrContract
	}
	return result.presentation, nil
}
