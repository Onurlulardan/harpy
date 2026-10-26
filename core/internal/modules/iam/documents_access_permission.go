package iam

type documentsAccessTarget struct{ resource, action string }

var documentsAccessTargets = map[string]documentsAccessTarget{
	"documents.folder.create\x00documents.folder.create":                 {"documents.folder", "create"},
	"documents.folder.read\x00documents.folder.get":                      {"documents.folder", "read"},
	"documents.folder.read\x00documents.folder.list":                     {"documents.folder", "read"},
	"documents.folder.read\x00documents.folder.contents":                 {"documents.folder", "read"},
	"documents.folder.read\x00documents.compat.legacy_sync_archive":      {"documents.folder", "read"},
	"documents.folder.update\x00documents.folder.update":                 {"documents.folder", "update"},
	"documents.folder.move\x00documents.folder.move":                     {"documents.folder", "move"},
	"documents.folder.delete\x00documents.folder.delete":                 {"documents.folder", "delete"},
	"documents.folder.delete\x00documents.folder.restore":                {"documents.folder", "delete"},
	"documents.item.read\x00documents.item.list":                         {"documents.item", "read"},
	"documents.item.read\x00documents.item.batch":                        {"documents.item", "read"},
	"documents.item.read\x00documents.item.get":                          {"documents.item", "read"},
	"documents.item.read\x00documents.item.versions.list":                {"documents.item", "read"},
	"documents.item.read\x00documents.item.version.latest":               {"documents.item", "read"},
	"documents.item.read\x00documents.upload.get":                        {"documents.item", "read"},
	"documents.item.create\x00documents.upload.create":                   {"documents.item", "create"},
	"documents.item.create\x00documents.compat.legacy_multipart_upload":  {"documents.item", "create"},
	"documents.item.update\x00documents.item.update":                     {"documents.item", "update"},
	"documents.item.move\x00documents.item.move":                         {"documents.item", "move"},
	"documents.item.copy\x00documents.item.copy":                         {"documents.item", "copy"},
	"documents.item.delete\x00documents.item.delete":                     {"documents.item", "delete"},
	"documents.item.delete\x00documents.item.restore":                    {"documents.item", "delete"},
	"documents.item.download\x00documents.item.download":                 {"documents.item", "download"},
	"documents.item.download\x00documents.item.version.download":         {"documents.item", "download"},
	"documents.item.download\x00documents.archive.create":                {"documents.item", "download"},
	"documents.item.download\x00documents.archive.get":                   {"documents.item", "download"},
	"documents.item.download\x00documents.archive.download":              {"documents.item", "download"},
	"documents.item.version\x00documents.upload.put":                     {"documents.item", "version"},
	"documents.item.version\x00documents.upload.cancel":                  {"documents.item", "version"},
	"documents.item.version\x00documents.compat.legacy_multipart_upload": {"documents.item", "version"},
	"documents.item.purge\x00documents.item.purge":                       {"documents.item", "purge"},
	"documents.item.purge\x00documents.item.cancel_purge":                {"documents.item", "purge"},
}

func resolveDocumentsAccess(permission, operation string) (documentsAccessTarget, bool) {
	target, found := documentsAccessTargets[permission+"\x00"+operation]
	return target, found
}

func documentsOperationRequiresStepUp(operation string) bool {
	return operation == "documents.item.purge"
}
