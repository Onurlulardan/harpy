package executor

func (runtime *graphRuntime) result() (Result, error) {
	if runtime == nil || runtime.page.Validate() != nil ||
		!uniqueMaterialRecords(runtime.root) {
		return Result{}, ErrContract
	}
	nodes := make([]materialNode, len(runtime.nodes))
	for index := range runtime.nodes {
		nodes[index] = runtime.nodes[index].material
	}
	page, err := materializeGraph(
		runtime.page, materialGraph{root: runtime.root, nodes: nodes},
	)
	if err != nil {
		return Result{}, err
	}
	return newResult(page, runtime.total, runtime.batch.hasExactCount)
}
