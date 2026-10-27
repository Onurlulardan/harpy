package publicexecution

func NewResolver() (Roots, error) {
	query, dashboard, err := bindRoots()
	if err != nil {
		return nil, err
	}
	return &roots{query: query, dashboard: dashboard}, nil
}
