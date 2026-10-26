package published

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestBindGETReportsUnknownMissingAndEscapedPaths(t *testing.T) {
	manifest := testBindingManifest(t)
	_, err := BindGET(manifest, map[string][]api.Literal{
		"x":     {mustTextLiteral(t, "value")},
		"a/b~c": {mustTextLiteral(t, "value")},
	}, nil)
	requirePublishedError(t, err, []expectedPublishedIssue{
		{"/param/a~1b~0c", restqueryv1.IssueParameterUnknown},
		{"/param/count", restqueryv1.IssueParameterRequired},
		{"/param/x", restqueryv1.IssueParameterUnknown},
	})
}

func TestBindReportsExactTypeAndArityPaths(t *testing.T) {
	manifest := testBindingManifest(t)
	tests := []struct {
		name     string
		post     bool
		provided map[string][]api.Literal
		path     string
		issue    restqueryv1.IssueCode
	}{
		{"range", false, bindingValues(t, "11"), "/param/count",
			restqueryv1.IssueValueOutOfRange},
		{"single arity", false, map[string][]api.Literal{
			"count": {mustTextLiteral(t, "1"), mustTextLiteral(t, "2")},
		}, "/param/count", restqueryv1.IssueParameterArityMismatch},
		{"list item", false, map[string][]api.Literal{
			"count": {mustTextLiteral(t, "1")},
			"flags": {mustTextLiteral(t, "true"), mustTextLiteral(t, "TRUE")},
		}, "/param/flags/1", restqueryv1.IssueValueNotCanonical},
		{"post path", true, map[string][]api.Literal{
			"count": {mustTextLiteral(t, "11")},
		}, "/params/count", restqueryv1.IssueValueOutOfRange},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var err error
			if test.post {
				_, err = BindPOST(manifest, test.provided, nil)
			} else {
				_, err = BindGET(manifest, test.provided, nil)
			}
			expected := []expectedPublishedIssue{{test.path, test.issue}}
			if _, found := test.provided["count"]; !found {
				expected = append(expected, expectedPublishedIssue{
					"/param/count", restqueryv1.IssueParameterRequired,
				})
			}
			requirePublishedError(t, err, expected)
		})
	}
}

func TestBindRejectsManifestPageMaximum(t *testing.T) {
	manifest := testBindingManifest(t)
	size := 21
	page, err := NewPageOverride(&size, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = BindGET(manifest, bindingValues(t, "1"), &page)
	requirePublishedError(t, err, []expectedPublishedIssue{
		{"/page/size", restqueryv1.IssuePageSizeExceeded},
	})
}

func TestBindReportsEveryInvalidListItem(t *testing.T) {
	_, err := BindGET(testBindingManifest(t), map[string][]api.Literal{
		"count": {mustTextLiteral(t, "1")},
		"flags": {mustTextLiteral(t, "TRUE"), mustTextLiteral(t, "yes")},
	}, nil)
	requirePublishedError(t, err, []expectedPublishedIssue{
		{"/param/flags/0", restqueryv1.IssueValueNotCanonical},
		{"/param/flags/1", restqueryv1.IssueValueNotCanonical},
	})
}

func bindingValues(t *testing.T, count string) map[string][]api.Literal {
	t.Helper()
	return map[string][]api.Literal{"count": {mustTextLiteral(t, count)}}
}
