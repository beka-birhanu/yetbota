package constants

import (
	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/identity/v1"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
)

const (
	MAXOTP         int32 = 3
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
	pb.AuthService_Login_FullMethodName:             {},
	pb.AuthService_Refresh_FullMethodName:           {},
	pb.AuthService_GenerateMobileOTP_FullMethodName: {},
	pb.AuthService_ValidateOTP_FullMethodName:       {},
	pb.AuthService_NewPassword_FullMethodName:       {},
	pb.UserService_Register_FullMethodName:          {},
	pb.UserService_CheckMobile_FullMethodName:       {},
	pb.UserService_ReadPublic_FullMethodName:        {},
}
