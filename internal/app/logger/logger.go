package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(name, version, env, level string) (*zap.Logger, error) {
	var config zap.Config
	switch env {
	case "local":
		fallthrough
	case "docker":
		config = zap.NewDevelopmentConfig()
	default:
		config = zap.NewProductionConfig()
	}

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	err := config.Level.UnmarshalText([]byte(level))
	if err != nil || len(level) == 0 {
		config.Level.SetLevel(zap.DebugLevel)
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	logger = logger.With(
		zap.String("name", name),
		zap.String("ver", version),
		zap.String("env", env),
	)

	return logger, nil
}
