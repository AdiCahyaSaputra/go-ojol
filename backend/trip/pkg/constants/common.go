package constants

const (
	ENUM_RUN_PRODUCTION = "production"
	ENUM_RUN_TESTING    = "testing"

	ENUM_PAGINATION_PER_PAGE = 10
	ENUM_PAGINATION_PAGE     = 1

	ENUM_ROLE_ADMIN    = "admin"
	ENUM_ROLE_CUSTOMER = "customer"
	ENUM_ROLE_DRIVER   = "driver"

	ENUM_RESOURCE_TRIP     = "trip"
	ENUM_RESOURCE_DISPATCH = "dispatch"

	ENUM_ACTION_CREATE = "create"
	ENUM_ACTION_READ   = "read"
	ENUM_ACTION_UPDATE = "update"
	ENUM_ACTION_DELETE = "delete"

	DB             = "db"
	JWKSVerifier   = "JWKSVerifier"
	CasbinEnforcer = "CasbinEnforcer"
	Redis          = "redis"
)
