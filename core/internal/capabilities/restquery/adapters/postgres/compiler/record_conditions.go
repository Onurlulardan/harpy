package compiler

func recordConditions(input recordInput, binder *binder) ([]string, error) {
	conditions, err := rootConditions(input.query, input.visibility, binder)
	if err != nil {
		return nil, err
	}
	if len(input.last) != 0 {
		sorts, err := input.query.Sorts()
		if err != nil {
			return nil, ErrContract
		}
		compiled, err := compileKeyset(sorts, input.last, binder)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, parenthesize(compiled))
	}
	return conditions, nil
}
