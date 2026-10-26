package authorization

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"

func requestMemberPath(
	path resolved.RequestPath,
	member string,
) (resolved.RequestPath, error) {
	value, err := path.Value()
	if err != nil || member == "" || member[0] != '/' {
		return resolved.RequestPath{}, errPolicyEvaluationContract
	}
	result, err := resolved.NewRequestPath(value + member)
	if err != nil {
		return resolved.RequestPath{}, errPolicyEvaluationContract
	}
	return result, nil
}
