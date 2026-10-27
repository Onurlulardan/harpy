package compiler

const explainJSONPrefix = "EXPLAIN (FORMAT JSON) "

func (statement RecordStatement) ExplainSQL() (string, error) {
	if statement.Validate() != nil {
		return "", ErrContract
	}
	return explainJSONPrefix + statement.sql, nil
}

func (statement ExactCountStatement) ExplainSQL() (string, error) {
	if statement.Validate() != nil {
		return "", ErrContract
	}
	return explainJSONPrefix + statement.sql, nil
}
