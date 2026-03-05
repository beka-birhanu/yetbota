package config

type Configs struct {
	Debug     bool        `yaml:"debug" mapstructure:"debug" validate:"required"`
	App       *App        `yaml:"app" mapstructure:"app" validate:"required"`
	Rest      *Rest       `yaml:"rest" mapstructure:"rest" validate:"required"`
	Grpc      *Grpc       `yaml:"grpc" mapstructure:"grpc" validate:"required"`
	Cors      *Cors       `yaml:"cors" mapstructure:"cors" validate:"required"`
	Embedding *Embedding  `yaml:"embedding" mapstructure:"embedding" validate:"required"`
	VectorDB  *VectorDB   `yaml:"vectordb" mapstructure:"vectordb" validate:"required"`
	LLM       *LLM        `yaml:"llm" mapstructure:"llm" validate:"required"`
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

type Embedding struct {
	Model     string `yaml:"model" mapstructure:"model" validate:"required"`
	Dimensions int   `yaml:"dimensions" mapstructure:"dimensions" validate:"required"`
	BaseURL   string `yaml:"base_url" mapstructure:"base_url" validate:"required"`
	APIKey    string `yaml:"api_key" mapstructure:"api_key" validate:"required"`
	TimeoutSeconds int `yaml:"timeout_seconds" mapstructure:"timeout_seconds" validate:"required"`
}

type VectorDB struct {
	URL        string `yaml:"url" mapstructure:"url" validate:"required"`
	APIKey     string `yaml:"api_key" mapstructure:"api_key" validate:"required"`
	Collection string `yaml:"collection" mapstructure:"collection" validate:"required"`
}

type LLM struct {
	Model           string  `yaml:"model" mapstructure:"model" validate:"required"`
	APIKey          string `yaml:"api_key" mapstructure:"api_key" validate:"required"`
	MaxTokens       int    `yaml:"max_tokens" mapstructure:"max_tokens" validate:"required"`
	Temperature     float64 `yaml:"temperature" mapstructure:"temperature" validate:"required"`
	TimeoutSeconds  int    `yaml:"timeout_seconds" mapstructure:"timeout_seconds" validate:"required"`
}