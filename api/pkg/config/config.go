package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	OpenAI    OpenAI
	Database  Database
	WebServer WebServer
	Worker    Worker
	Auth      Auth
}

type OpenAI struct {
	APIKey string `envconfig:"OPENAI_KEY" description:"The api key for openAI" required:"true"`
	URL    string `envconfig:"OPENAI_URL"  description:"The URL for openAI" required:"true"`
}

// Database points at a Supabase Postgres instance (there is no local Postgres).
// Use the session pooler host (aws-0-<region>.pooler.supabase.com:5432, username
// postgres.<project-ref>) or the direct connection (db.<ref>.supabase.co:5432);
// both keep a 1:1 connection so GORM's prepared statements work. SSL is required.
type Database struct {
	Host            string        `envconfig:"POSTGRES_HOST" description:"Supabase host, e.g. aws-0-eu-west-1.pooler.supabase.com or db.<ref>.supabase.co." required:"true"`
	Port            int           `envconfig:"POSTGRES_PORT_OVERRIDE" default:"5432" description:"The Supabase port (5432 session pooler / direct)."`
	Database        string        `envconfig:"POSTGRES_DATABASE" default:"postgres" description:"The database to connect to (Supabase default: postgres)."`
	Username        string        `envconfig:"POSTGRES_USER" description:"Supabase user: postgres.<project-ref> for the pooler, or postgres for direct." required:"true"`
	Password        string        `envconfig:"POSTGRES_PASSWORD" description:"The Supabase database password." required:"true"`
	SSL             bool          `envconfig:"POSTGRES_SSL" default:"true" description:"Supabase requires SSL; keep this true."`
	Schema          string        `envconfig:"POSTGRES_SCHEMA"` // Defaults to public
	AutoMigrate     bool          `envconfig:"POSTGRES_AUTO_MIGRATE" default:"true" description:"Apply pending Go migrations on boot?"`
	MaxConns        int           `envconfig:"POSTGRES_MAX_CONNS" default:"20" description:"Keep modest - Supabase enforces connection ceilings per plan."`
	IdleConns       int           `envconfig:"POSTGRES_IDLE_CONNS" default:"5"`
	MaxConnLifetime time.Duration `envconfig:"POSTGRES_MAX_CONN_LIFETIME" default:"1h"`
	MaxConnIdleTime time.Duration `envconfig:"POSTGRES_MAX_CONN_IDLE_TIME" default:"1m"`
}

type WebServer struct {
	Enabled       bool   `envconfig:"SERVER_ENABLED" default:"true" description:"Run the HTTP API server in this process."`
	Host          string `envconfig:"SERVER_HOST" default:"0.0.0.0" description:"The host to bind the api server to."`
	Port          int    `envconfig:"SERVER_PORT" default:"80" description:"The port to bind the api server to."`
	URL           string `envconfig:"SERVER_URL" default:"http://localhost" description:"The base url for the api server."`
	APIPath       string `envconfig:"SERVER_API_PATH" default:"/api/v1" description:"The base path for the api server."`
	JWTSecret     string `envconfig:"SERVER_JWT_SECRET" description:"The secret for the jwt." required:"true"`
	FixedPassword string `envconfig:"SERVER_FIXED_PASSWORD" description:"The fixed password for the api." required:"true"`
}

type Worker struct {
	Enabled      bool          `envconfig:"WORKER_ENABLED" default:"true" description:"Run the job queue worker in this process."`
	Concurrency  int           `envconfig:"WORKER_CONCURRENCY" default:"10" description:"The number of parallel poll loops to run."`
	MaxAttempts  int           `envconfig:"WORKER_MAX_ATTEMPTS" default:"3" description:"The maximum number of attempts for a job."`
	PollInterval time.Duration `envconfig:"WORKER_POLL_INTERVAL" default:"1s" description:"How long to wait before polling the job queue again when it is empty."`
	APIURL       string        `envconfig:"WORKER_SERVER_URL" default:"http://api" description:"The url for workers to connect to the api."`
	Secret       string        `envconfig:"WORKER_SECRET" description:"The secret for the worker." required:"true"`
}

// Auth selects which authentication providers verify incoming bearer tokens.
// The HTTP API verifies (does not issue) Supabase JWTs via the project's JWKS
// endpoint, and optionally keeps a local fixed-password path for dev/CI. At
// least one provider must be enabled (validated in server.NewServer).
type Auth struct {
	SupabaseEnabled bool   `envconfig:"AUTH_SUPABASE_ENABLED" default:"true" description:"Verify Supabase-issued JWTs (Google etc.) via the project JWKS."`
	SupabaseURL     string `envconfig:"SUPABASE_URL" description:"Supabase project URL, e.g. https://<ref>.supabase.co (required when AUTH_SUPABASE_ENABLED)."`
	SupabaseAud     string `envconfig:"SUPABASE_JWT_AUD" default:"authenticated" description:"Expected audience claim on Supabase JWTs."`
	LocalEnabled    bool   `envconfig:"AUTH_LOCAL_ENABLED" default:"true" description:"Keep the local fixed-password login + HS256 token path (dev/CI)."`
}

func LoadConfig() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}
