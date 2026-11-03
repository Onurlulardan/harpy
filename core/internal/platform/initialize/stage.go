package initialize

const (
	StageInitialize  = "initialize"
	StageFoundation  = "foundation"
	StageGoose       = "goose"
	StageApplication = "application"
)

type StageStatus struct {
	Name string
	OK   bool
}

type Result struct {
	Stages []StageStatus
}

func (result *Result) succeed(name string) {
	result.Stages = append(result.Stages, StageStatus{Name: name, OK: true})
}

func (result *Result) fail(name string) {
	result.Stages = append(result.Stages, StageStatus{Name: name, OK: false})
}
