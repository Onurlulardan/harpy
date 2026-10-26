package domain

import "testing"

func TestTemplateEscapesSystemVariablesAndRejectsMaliciousPlainText(t *testing.T) {
	templateID, _ := ParseID("11111111-1111-4111-8111-111111111111")
	versionID, _ := ParseID("22222222-2222-4222-8222-222222222222")
	value := PublishedTemplate{ID: templateID, VersionID: versionID,
		NotificationType: "security.login", Kind: TemplateSystemHTML, Locale: "tr",
		SubjectPattern: "Merhaba {{name}}", BodyPattern: "<p>{{name}}</p>",
		Variables: []string{"name"}, Revision: 1}
	rendered, err := value.Render(map[string]string{"name": "<script>alert(1)</script>"})
	if err != nil || rendered.Body != "<p>&lt;script&gt;alert(1)&lt;/script&gt;</p>" {
		t.Fatalf("Render() = %+v, %v", rendered, err)
	}
	value.Kind = TemplatePlainText
	if _, err = value.Render(map[string]string{"name": "<b>unsafe</b>"}); err == nil {
		t.Fatal("plain-text template accepted markup")
	}
}
