package catalogcomponentpostgres

import (
	"context"
	"fmt"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
)

func (storage *Store) Read(
	ctx context.Context, component contract.Component,
) (contract.Readiness, error) {
	if !storage.usable(ctx) || component.ID == "" || component.Owner != storage.owner ||
		component.Contract == "" || !digestPattern.MatchString(component.ContentSHA256) {
		return contract.Readiness{}, fmt.Errorf("catalog component read is invalid")
	}
	var id, componentContract, digest, state string
	if err := storage.queryRow(ctx, storage.queries.Read, component.ID, storage.generation).
		Scan(&id, &componentContract, &digest, &state); err != nil {
		return contract.Readiness{}, fmt.Errorf("catalog component is unavailable")
	}
	value := contract.Readiness{Component: contract.Component{
		ID: id, Owner: storage.owner, Generation: storage.generation,
		Contract: componentContract, ContentSHA256: digest,
	}, WriterV2: true, State: state}
	if value.ID != component.ID || value.Contract != component.Contract ||
		value.ContentSHA256 != component.ContentSHA256 || state == "PURGED" {
		return contract.Readiness{}, fmt.Errorf("catalog component differs")
	}
	return value, nil
}
