package iam

type formsRecordAccessTarget struct {
	action string
}

var formsRecordAccessTargets = map[string]formsRecordAccessTarget{
	"create\x00forms.record_kernel.create":        {action: "create"},
	"read\x00forms.record_kernel.get":             {action: "read"},
	"read\x00restquery.execute":                   {action: "read"},
	"update\x00forms.record_kernel.update":        {action: "update"},
	"delete\x00forms.record_kernel.delete":        {action: "delete"},
	"export\x00forms.record_export.submit":        {action: "export"},
	"export\x00forms.record_export.get":           {action: "export"},
	"export\x00forms.record_export.delete":        {action: "export"},
	"export\x00forms.record_export.download":      {action: "export"},
	"import\x00forms.record_import.create_source": {action: "import"},
	"import\x00forms.record_import.load_source":   {action: "import"},
	"import\x00forms.record_import.submit":        {action: "import"},
	"import\x00forms.record_import.get":           {action: "import"},
	"import\x00forms.record_import.list":          {action: "import"},
	"import\x00forms.record_import.cancel":        {action: "import"},
	"import\x00forms.record_import.resume":        {action: "import"},
	"import\x00forms.record_import.list_errors":   {action: "import"},
}

func resolveFormsRecordAccess(
	action, operation string,
) (formsRecordAccessTarget, bool) {
	target, found := formsRecordAccessTargets[action+"\x00"+operation]
	return target, found
}
