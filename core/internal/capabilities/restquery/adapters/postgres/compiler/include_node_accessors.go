package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func (node IncludeNode) SourceCatalog() (api.CatalogReference, error) {
	if node.Validate() != nil {
		return api.CatalogReference{}, ErrContract
	}
	return node.sourceCatalog, nil
}

func (node IncludeNode) Catalog() (api.CatalogReference, error) {
	if node.Validate() != nil {
		return api.CatalogReference{}, ErrContract
	}
	return node.targetCatalog, nil
}

func (node IncludeNode) SourceOwner() (string, error) {
	catalog, err := node.SourceCatalog()
	if err != nil {
		return "", ErrContract
	}
	owner, err := catalog.Owner()
	if err != nil {
		return "", ErrContract
	}
	return owner, nil
}

func (node IncludeNode) Owner() (string, error) {
	catalog, err := node.Catalog()
	if err != nil {
		return "", ErrContract
	}
	owner, err := catalog.Owner()
	if err != nil {
		return "", ErrContract
	}
	return owner, nil
}

func (node IncludeNode) Path() (resolved.RequestPath, error) {
	if node.Validate() != nil {
		return resolved.RequestPath{}, ErrContract
	}
	return node.path, nil
}

func (node IncludeNode) RelationName() (string, error) {
	if node.Validate() != nil {
		return "", ErrContract
	}
	return node.relationName, nil
}

func (node IncludeNode) Cardinality() (api.RelationCardinality, error) {
	if node.Validate() != nil {
		return 0, ErrContract
	}
	return node.cardinality, nil
}

func (node IncludeNode) Limit() (int, error) {
	if node.Validate() != nil {
		return 0, ErrContract
	}
	return node.limit, nil
}

func (node IncludeNode) Depth() (int, error) {
	if node.Validate() != nil {
		return 0, ErrContract
	}
	return node.depth, nil
}

func (node IncludeNode) Ordinal() (int, error) {
	if node.Validate() != nil {
		return 0, ErrContract
	}
	return node.ordinal, nil
}

func (node IncludeNode) ParentOrdinal() (int, error) {
	if node.Validate() != nil {
		return 0, ErrContract
	}
	return node.parentOrdinal, nil
}

func (node IncludeNode) SiblingOrdinal() (int, error) {
	if node.Validate() != nil {
		return 0, ErrContract
	}
	return node.siblingOrdinal, nil
}
