package api

type CreateInput struct {
	ID         string
	Resource   ResourceKind
	Definition Definition
}

type UpdateInput struct {
	ID                  string
	ExpectedRowRevision int64
	Definition          Definition
}

type DeleteInput struct {
	ID                  string
	ExpectedRowRevision int64
}
