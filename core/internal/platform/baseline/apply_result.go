package baseline

func combineBaselineResults(steps ...Result) Result {
	combined := Result{AppliedVersions: []int64{}}
	for _, step := range steps {
		if step.Version > combined.Version {
			combined.Version = step.Version
		}
		if step.Applied {
			combined.Applied = true
			combined.AppliedVersions = append(combined.AppliedVersions, step.Version)
		}
	}
	return combined
}
