package config

import (
	"flag"
	"os"
)

const (
	RunADDRESS_KEY = "RUN_ADDRESS"
	DSN_KEY        = "DATABASE_URI"
)

type Logger struct {
	Level  string
	Format string
}

func (l *Logger) String() string {
	return "Level: " + l.Level + "\n" +
		"Format: " + l.Format + "\n"
}

type Config struct {
	Logger
	RunAddress string
	Dsn        string
	Secret     string
}

func (c *Config) String() string {
	return "Logger: " + c.Logger.String() + "\n" +
		"RunAddress: " + c.RunAddress + "\n" +
		"Dsn: " + c.Dsn + "\n"
}

func InitConfig() *Config {
	const (
		runAddressFlagUsage = `Base URL, e.g. "http://localhost:8080/"`
		DsnFlagUsage        = `Dsn address`
	)
	cfg := Config{}

	cfg.Logger.Level = "debug"
	if level, ok := os.LookupEnv("LOG_LEVEL"); ok {
		cfg.Logger.Level = level
	}
	cfg.Logger.Format = "text"
	if format, ok := os.LookupEnv("LOG_FORMAT"); ok {
		cfg.Logger.Format = format
	}

	cfg.RunAddress = "http://localhost:8080/"

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, runAddressFlagUsage)
	flag.StringVar(&cfg.Dsn, "d", cfg.Dsn, DsnFlagUsage)

	flag.Parse()

	secret, ok := os.LookupEnv("SECRET_KEY")
	if !ok {
		cfg.Secret = ""
	} else {
		cfg.Secret = secret
	}

	if runAddress, ok := os.LookupEnv(RunADDRESS_KEY); ok {
		cfg.RunAddress = runAddress
	}
	if dsn, ok := os.LookupEnv(DSN_KEY); ok {
		cfg.Dsn = dsn
	}

	return &cfg
}
