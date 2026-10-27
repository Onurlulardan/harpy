package deadletterdelivery

import "github.com/Onurlulardan/harpy/core/internal/platform/admission"

type Runner struct {
	connector Connector
	settings  Settings
	jitter    jitterSource
	admission *admission.Gate
	lifecycle runnerLifecycle
}
