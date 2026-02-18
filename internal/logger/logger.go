package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/IvanOplesnin/gofermart.git/internal/config"
	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func SetupLogger(cfg *config.Logger) error {
	msg := "logger.setupLogger"
	level := cfg.Level
	format := cfg.Format

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("%s fail parse level string %s: %w", msg, level, err)
	}

	Log.SetLevel(logLevel)
	formatter, err := getFormatter(format)
	if err != nil {
		return fmt.Errorf("%s fail get formatter: %w", msg, err)
	}
	Log.SetFormatter(formatter)
	Log.SetOutput(os.Stdout)

	return nil
}

func getFormatter(format string) (logrus.Formatter, error) {
	form := strings.ToLower(string(format))

	switch form {
	case "text":
		return &logrus.TextFormatter{
			FullTimestamp: true,
		}, nil
	case "json":
		return &logrus.JSONFormatter{}, nil
	default:
		// Фолбэк по умолчанию, если формат неизвестен
		return nil, fmt.Errorf("unknown formatter: %s", format)
	}
}
