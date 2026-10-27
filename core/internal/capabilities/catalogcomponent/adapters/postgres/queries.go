package catalogcomponentpostgres

import "fmt"

type Queries struct {
	Read     string
	Apply    string
	Blockers string
	Allows   string
}

func (queries Queries) validate() error {
	if queries.Read == "" || queries.Apply == "" ||
		queries.Blockers == "" || queries.Allows == "" {
		return fmt.Errorf("catalog component SQL documents are required")
	}
	return nil
}
