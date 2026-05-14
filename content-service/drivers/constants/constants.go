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
	MB                    = 1 << (10 * 2)
	MaxUploadSize         = 20 * MB
	URLExpiration         = 30
	MaxImageResolution    = 3840 // 4K cap for stored original
	WebImageResolution    = 1080
	MobileImageResolution = 750
)

const (
	MigrationFolder = "migrations"
)

const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

const (
	FeedUpdateWorkflowQueue    = "WF_FEED_UPDATE_QUEUE"
	NewPostWorkflowQueue       = "WF_NEW_POST_QUEUE"
	PostEmbeddingWorkflowQueue = "WF_POST_EMBEDDING_QUEUE"

	PostEmbeddingWorkflowName = "POST_EMBEDDING"
)

const (
	FeedFanOutBatchSize   = 500
	FanOutBatchTTLSeconds = 3600
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
