package config

type Configs struct {
	Debug    bool      `yaml:"debug" mapstructure:"debug" validate:"required"`
	App      *App      `yaml:"app" mapstructure:"app" validate:"required"`
	Rest     *Rest     `yaml:"rest" mapstructure:"rest" validate:"required"`
	Grpc     *Grpc     `yaml:"grpc" mapstructure:"grpc" validate:"required"`
	Cors     *Cors     `yaml:"cors" mapstructure:"cors" validate:"required"`
	Postgres *Postgres `yaml:"database" mapstructure:"database" validate:"required"`
	Redis    *Redis    `yaml:"redis" mapstructure:"redis" validate:"required"`
	Otp      *Otp      `yaml:"otp" mapstructure:"otp" validate:"required"`
	Jwt      *Jwt      `yaml:"jwt" mapstructure:"jwt" validate:"required"`
	AWS      *AWS      `yaml:"aws" mapstructure:"aws" validate:"required"`
	Temporal *Temporal `yaml:"temporal" mapstructure:"temporal" validate:"required"`
	Feed     *Feed     `yaml:"feed" mapstructure:"feed" validate:"required"`
	Neo4j    *Neo4j    `yaml:"neo4j" mapstructure:"neo4j" validate:"required"`
}

type Temporal struct {
	Host      string `yaml:"host"      mapstructure:"host"      validate:"required"`
	Namespace string `yaml:"namespace" mapstructure:"namespace" validate:"required"`
}

type AWS struct {
	S3 *AWSS3 `yaml:"s3" mapstructure:"s3" validate:"required"`
}

type AWSS3 struct {
	Region string `yaml:"region" mapstructure:"region" validate:"required"`
	Bucket string `yaml:"bucket" mapstructure:"bucket" validate:"required"`
}

type App struct {
	Name    string `yaml:"name" mapstructure:"name" validate:"required"`
	Version string `yaml:"version" mapstructure:"version" validate:"required"`
}

type Rest struct {
	Port         int `yaml:"port" mapstructure:"port" validate:"required"`
	ReadTimeout  int `yaml:"read_timeout" mapstructure:"read_timeout" validate:"required"`
	WriteTimeout int `yaml:"write_timeout" mapstructure:"write_timeout" validate:"required"`
	IdleTimeout  int `yaml:"idle_timeout" mapstructure:"idle_timeout" validate:"required"`
}

type Grpc struct {
	Port      int        `yaml:"port" mapstructure:"port" validate:"required"`
	Keepalive *Keepalive `yaml:"keepalive" mapstructure:"keepalive" validate:"required"`
}

type Keepalive struct {
	Time                int  `yaml:"time" mapstructure:"time" validate:"required"`
	Timeout             int  `yaml:"timeout" mapstructure:"timeout" validate:"required"`
	MinTime             int  `yaml:"min_time" mapstructure:"min_time" validate:"required"`
	PermitWithoutStream bool `yaml:"permit_without_stream" mapstructure:"permit_without_stream" validate:"required"`
}

type Cors struct {
	Hosts []string `yaml:"hosts" mapstructure:"hosts" validate:"required"`
}

type Postgres struct {
	DB       string `yaml:"db" mapstructure:"db" validate:"required"`
	Host     string `yaml:"host" mapstructure:"host" validate:"required"`
	Password string `yaml:"password" mapstructure:"password" validate:"required"`
	Port     string `yaml:"port" mapstructure:"port" validate:"required"`
	User     string `yaml:"user" mapstructure:"user" validate:"required"`
}

type Redis struct {
	Address  string `yaml:"address"  mapstructure:"address"  validate:"required"`
	Password string `yaml:"password" mapstructure:"password" validate:"required"`

	DB           int `yaml:"db"            mapstructure:"db"`
	PoolSize     int `yaml:"pool_size"     mapstructure:"pool_size"     validate:"required"`
	MinIdleConns int `yaml:"min_idle_conns" mapstructure:"min_idle_conns"`
	MaxIdleConns int `yaml:"max_idle_conns" mapstructure:"max_idle_conns" validate:"required"`
	MaxRetries   int `yaml:"max_retries"   mapstructure:"max_retries"   validate:"required"`

	DialTimeout     int `yaml:"dial_timeout"      mapstructure:"dial_timeout"      validate:"required"`
	ReadTimeout     int `yaml:"read_timeout"      mapstructure:"read_timeout"      validate:"required"`
	WriteTimeout    int `yaml:"write_timeout"     mapstructure:"write_timeout"     validate:"required"`
	PoolTimeout     int `yaml:"pool_timeout"      mapstructure:"pool_timeout"      validate:"required"`
	ConnMaxIdleTime int `yaml:"conn_max_idle_time" mapstructure:"conn_max_idle_time" validate:"required"`
	ConnMaxLifetime int `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime" validate:"required"`

	TLS bool `yaml:"tls" mapstructure:"tls"`
}

type Otp struct {
	TTL             int `yaml:"ttl" mapstructure:"ttl" validate:"required"`
	LockRequestTime int `yaml:"lockRequestTime" mapstructure:"lockRequestTime" validate:"required"`
	LockInvalidTime int `yaml:"lockInvalidTime" mapstructure:"lockInvalidTime" validate:"required"`
}

type Jwt struct {
	Algorithm    string    `yaml:"algorithm" mapstructure:"algorithm" validate:"required"`
	AccessToken  *JwtToken `yaml:"access_token" mapstructure:"access_token" validate:"required"`
	RefreshToken *JwtToken `yaml:"refresh_token" mapstructure:"refresh_token" validate:"required"`
}

type JwtToken struct {
	Expiration int    `yaml:"expiration" mapstructure:"expiration" validate:"required"`
	Secret     string `yaml:"secret" mapstructure:"secret" validate:"required"`
}

type Feed struct {
	HalfLifeHours    float64 `yaml:"half_life_hours"    mapstructure:"half_life_hours"`
	FeedSize         int     `yaml:"feed_size"          mapstructure:"feed_size"`
	ColdStartN       int     `yaml:"cold_start_n"       mapstructure:"cold_start_n"`
	ScoreChangeDelta float64 `yaml:"score_change_delta" mapstructure:"score_change_delta"`
	NewPostSeedScore float64 `yaml:"new_post_seed_score" mapstructure:"new_post_seed_score"`
	TaskQueue        string  `yaml:"task_queue"         mapstructure:"task_queue"`
	ScoreTTLHours    float64 `yaml:"score_ttl_hours"    mapstructure:"score_ttl_hours"`
	StaleLimit       int     `yaml:"stale_limit"        mapstructure:"stale_limit"`
	RefillThreshold  int     `yaml:"refill_threshold"   mapstructure:"refill_threshold"`
	MaxFollowerDepth int     `yaml:"max_follower_depth" mapstructure:"max_follower_depth"`
	MaxSimDepth      int     `yaml:"max_sim_depth"      mapstructure:"max_sim_depth"`
	FanOutLimit      int     `yaml:"fan_out_limit"      mapstructure:"fan_out_limit"`
	SeenCacheTTL     int64   `yaml:"seen_cache_ttl"     mapstructure:"seen_cache_ttl"`
}

type Neo4j struct {
	URI      string `yaml:"uri"      mapstructure:"uri"`
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
}
