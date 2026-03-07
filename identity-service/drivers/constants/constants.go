package constants

import "github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"

const (
	MAXOTP         int32 = 2
	MaxNotMatchOtp int32 = 3
)

const (
	DefaultPaginationLength = 15
	DefaultPhoneRegion      = "ETH"
)

const (
	MB                 = 1 << (10 * 2)
	MaxUploadSize      = 10 * MB
	URLExpiration      = 30
	MaxImageResolution = 480
	MaxImageSize       = MaxImageResolution * MaxImageResolution * 3
)

const (
	MigrationFolder = "migrations"
)

var SkipAuth = map[string]struct{}{}

var AllowedAccessMap = map[string]struct{}{
	dbmodels.RolesADMIN: {},
	dbmodels.RolesUSER:  {},
}

var AllowedAdminAccessMap = map[string]struct{}{
	dbmodels.RolesADMIN: {},
}

var SkipAuthGrpc = map[string]struct{}{
	"/identity.v1.AuthService/Login":             {},
	"/identity.v1.AuthService/GenerateMobileOTP": {},
	"/identity.v1.AuthService/ValidateOTP":       {},
	"/identity.v1.AuthService/NewPassword":       {},
	"/identity.v1.UserService/Register":          {},
	"/identity.v1.UserService/CheckMobile":       {},
	"/identity.v1.UserService/ReadPublic":        {},
}
