package config

import (
	"flag"
	"os"
)

const (
	RunADDRESSKEY         = "RUN_ADDRESS"
	DSNKEY                = "DATABASE_URI"
	ACCRUALSYSTEMADDRESS = "ACCRUAL_SYSTEM_ADDRESS"
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
	RunAddress            string
	Dsn                   string
	Secret                string
	AccrualServiceAddress string
}

func (c *Config) String() string {
	return "Logger: " + c.Logger.String() + "\n" +
		"RunAddress: " + c.RunAddress + "\n"
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
	cfg.AccrualServiceAddress = "http://localhost:8081/"

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, runAddressFlagUsage)
	flag.StringVar(&cfg.Dsn, "d", cfg.Dsn, DsnFlagUsage)
	flag.StringVar(&cfg.AccrualServiceAddress, "r", cfg.AccrualServiceAddress, "Accrual service address")

	flag.Parse()

	secret, ok := os.LookupEnv("SECRET_KEY")
	if !ok {
		cfg.Secret = ""
	} else {
		cfg.Secret = secret
	}

	if runAddress, ok := os.LookupEnv(RunADDRESSKEY); ok {
		cfg.RunAddress = runAddress
	}
	if dsn, ok := os.LookupEnv(DSNKEY); ok {
		cfg.Dsn = dsn
	}
	if accrualServiceAddress, ok := os.LookupEnv(ACCRUALSYSTEMADDRESS); ok {
		cfg.AccrualServiceAddress = accrualServiceAddress
	}

	return &cfg
}
