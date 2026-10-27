package postgres

type fixtureStateError string

func (err fixtureStateError) Error() string    { return string(err) }
func (err fixtureStateError) SQLState() string { return string(err) }
