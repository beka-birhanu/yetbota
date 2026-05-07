package constants

const (
	MAXOTP         int32 = 2
	MaxNotMatchOtp int32 = 3
)

const (
	DefaultPaginationLength = 15
	MaxPaginationLength     = 20
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

const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

const (
	FeedUpdateWorkflowQueue = "WF_FEED_UPDATE"
)

var SkipAuth = map[string]struct{}{}

var SkipAuthGrpc = map[string]struct{}{
	"/content.v1.PostService/List": {},
	"/content.v1.PostService/Read": {},

	"/content.comment.v1.CommentService/List": {},
	"/content.comment.v1.CommentService/Read": {},
}

var AllowedAccessMap = map[string]struct{}{
	RoleAdmin: {},
	RoleUser:  {},
}

var AllowedAdminAccessMap = map[string]struct{}{
	RoleAdmin: {},
}

var AllowedCSAAccessMap = map[string]struct{}{
	RoleAdmin: {},
}
