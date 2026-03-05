package constants

const (
	MAXOTP         int32 = 2
	MaxNotMatchOtp int32 = 3
)

const (
	DefaultPaginationLength = 15
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
