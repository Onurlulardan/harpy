package postgres

const V1RegistryRelation = "public.tenants"

var V1RegistryColumns = [...]string{
	"id",
	"name",
	"slug",
	"status",
	"settings",
	"owner_id",
	"created_at",
	"updated_at",
	"deleted_at",
}

var V1RegistryWritableColumns = [...]string{
	"name",
	"status",
	"settings",
	"owner_id",
	"updated_at",
	"deleted_at",
}
