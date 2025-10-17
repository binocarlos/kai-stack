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
}

type OpenAI struct {
	APIKey string `envconfig:"OPENAI_KEY" description:"The api key for openAI" required:"true"`
	URL    string `envconfig:"OPENAI_URL"  description:"The URL for openAI" required:"true"`
}

type Database struct {
	Host            string        `envconfig:"POSTGRES_HOST" description:"The host to connect to the postgres server." required:"true"`
	Port            int           `envconfig:"POSTGRES_PORT_OVERRIDE" default:"5432" description:"The port to connect to the postgres server."`
	Database        string        `envconfig:"POSTGRES_DATABASE" default:"forum" description:"The database to connect to the postgres server."`
	Username        string        `envconfig:"POSTGRES_USER" description:"The username to connect to the postgres server." required:"true"`
	Password        string        `envconfig:"POSTGRES_PASSWORD" description:"The password to connect to the postgres server." required:"true"`
	SSL             bool          `envconfig:"POSTGRES_SSL" default:"false"`
	Schema          string        `envconfig:"POSTGRES_SCHEMA"` // Defaults to public
	AutoMigrate     bool          `envconfig:"POSTGRES_AUTO_MIGRATE" default:"true" description:"Should we automatically run the migrations?"`
	MaxConns        int           `envconfig:"POSTGRES_MAX_CONNS" default:"50"`
	IdleConns       int           `envconfig:"POSTGRES_IDLE_CONNS" default:"25"`
	MaxConnLifetime time.Duration `envconfig:"POSTGRES_MAX_CONN_LIFETIME" default:"1h"`
	MaxConnIdleTime time.Duration `envconfig:"POSTGRES_MAX_CONN_IDLE_TIME" default:"1m"`
}

type WebServer struct {
	Host          string `envconfig:"SERVER_HOST" default:"0.0.0.0" description:"The host to bind the api server to."`
	Port          int    `envconfig:"SERVER_PORT" default:"80" description:"The port to bind the api server to."`
	URL           string `envconfig:"SERVER_URL" default:"http://localhost" description:"The base url for the api server."`
	APIPath       string `envconfig:"SERVER_API_PATH" default:"/api/v1" description:"The base path for the api server."`
	JWTSecret     string `envconfig:"SERVER_JWT_SECRET" description:"The secret for the jwt." required:"true"`
	FixedPassword string `envconfig:"SERVER_FIXED_PASSWORD" description:"The fixed password for the api." required:"true"`
}

type Worker struct {
	Concurrency int    `envconfig:"WORKER_CONCURRENCY" default:"10" description:"The number parallel workers to run - this should be the number of cores on the machine."`
	MaxAttempts int    `envconfig:"WORKER_MAX_ATTEMPTS" default:"3" description:"The maximum number of attempts for a job."`
	APIURL      string `envconfig:"WORKER_SERVER_URL" default:"http://api" description:"The url for workers to connect to the api."`
	Secret      string `envconfig:"WORKER_SECRET" description:"The secret for the worker." required:"true"`
}

func LoadConfig() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}
