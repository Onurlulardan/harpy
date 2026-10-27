package postgres

import (
	"context"
	"errors"
	"slices"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/jackc/pgx/v5"
)

type schemaExecutorIndexContract struct {
	owner                  string
	valid                  bool
	ready                  bool
	unique                 bool
	method                 string
	tableName              string
	fields                 []string
	operatorClasses        []string
	operatorClassSchemas   []string
	defaultOperatorClasses bool
	defaultCollations      bool
	defaultOrdering        bool
	noIncludedColumns      bool
	nullsDistinct          bool
	standalone             bool
	predicate              string
}

func schemaExecutorReadIndex(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	schemaName string,
	indexName string,
) (schemaExecutorIndexContract, bool, error) {
	var contract schemaExecutorIndexContract
	err := querier.QueryRow(
		ctx, schemaExecutorReadIndexSQL, schemaName, indexName,
	).Scan(
		&contract.owner, &contract.valid, &contract.ready, &contract.unique,
		&contract.method, &contract.tableName, &contract.fields,
		&contract.operatorClasses, &contract.operatorClassSchemas,
		&contract.defaultOperatorClasses, &contract.defaultCollations,
		&contract.defaultOrdering, &contract.noIncludedColumns,
		&contract.nullsDistinct, &contract.standalone, &contract.predicate,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return schemaExecutorIndexContract{}, false, nil
	}
	return contract, err == nil, err
}

func schemaExecutorIndexMatches(
	contract schemaExecutorIndexContract,
	mapping schemav1.IdentifierMapping,
	arguments schemav1.IndexArguments,
) bool {
	fields := make([]string, len(arguments.FieldIDs))
	for index, fieldID := range arguments.FieldIDs {
		for _, field := range mapping.Fields {
			if field.LogicalID == fieldID {
				fields[index] = field.PhysicalName
				break
			}
		}
	}
	predicateMatches := contract.predicate == ""
	if arguments.ActiveRecordsOnly {
		predicateMatches = contract.predicate == "(deleted_at IS NULL)" ||
			contract.predicate == "deleted_at IS NULL"
	}
	return contract.owner == "harpy_owner" &&
		contract.tableName == mapping.Form.PhysicalName &&
		contract.defaultOperatorClasses && contract.defaultCollations &&
		contract.defaultOrdering && contract.noIncludedColumns && contract.nullsDistinct &&
		contract.standalone &&
		contract.unique == arguments.Unique && contract.method == string(arguments.Method) &&
		slices.Equal(contract.fields, fields) && predicateMatches
}
