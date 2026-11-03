package baseline

func combineBaselineResults(steps ...Result) Result {
	combined := Result{AppliedVersions: []int64{}}
	for _, step := range steps {
		if step.Version > combined.Version {
			combined.Version = step.Version
		}
		versions := step.AppliedVersions
		if len(versions) == 0 && step.Applied {
			versions = []int64{step.Version}
		}
		if len(versions) == 0 {
			continue
		}
		combined.Applied = true
		combined.AppliedVersions = append(combined.AppliedVersions, versions...)
	}
	return combined
}
