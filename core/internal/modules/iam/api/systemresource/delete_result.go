package systemresource

type DeleteResult struct {
	resource  Resource
	catalog   FieldCatalog
	receipt   MutationReceipt
	validated bool
}

func NewDeleteResult(
	resource Resource,
	catalog FieldCatalog,
	receipt MutationReceipt,
) (DeleteResult, error) {
	result := DeleteResult{
		resource: resource, catalog: catalog, receipt: receipt, validated: true,
	}
	if result.Validate() != nil {
		return DeleteResult{}, ErrContract
	}
	return result, nil
}

func (result DeleteResult) Validate() error {
	catalogResource, err := result.catalog.Resource()
	if !result.validated || !result.resource.Valid() || err != nil ||
		catalogResource != result.resource || result.receipt.Validate() != nil {
		return ErrContract
	}
	return nil
}

func (result DeleteResult) Catalog() (FieldCatalog, error) {
	if result.Validate() != nil {
		return FieldCatalog{}, ErrContract
	}
	return result.catalog, nil
}

func (result DeleteResult) Receipt() (MutationReceipt, error) {
	if result.Validate() != nil {
		return MutationReceipt{}, ErrContract
	}
	return result.receipt, nil
}
