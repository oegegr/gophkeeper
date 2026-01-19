package postgresql

const (
	// Table names
	TableUsers   = "users"
	TableSecrets = "secrets"

	// Common columns
	ColumnID        = "id"
	ColumnCreatedAt = "created_at"
	ColumnUpdatedAt = "updated_at"

	// Users table columns
	ColumnLogin        = "login"
	ColumnPasswordHash = "password_hash"

	// Secrets table columns
	ColumnUserID  = "user_id"
	ColumnType    = "type"
	ColumnData    = "data"
	ColumnMeta    = "meta"
	ColumnVersion = "version"
	ColumnDeleted = "deleted"
)
