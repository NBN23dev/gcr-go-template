package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
)

// A Level is the importance or severity of a log event.
// The higher the level, the more important or severe the event.
type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Payload
type Payload = map[string]string

var logLevel = LevelInfo
var logClient *logging.Client
var logger *logging.Logger

// Init
func Init(name string, level Level) error {
	ctx := context.Background()

	projectId, err := metadata.ProjectID()

	if err != nil {
		return nil
	}

	client, err := logging.NewClient(ctx, fmt.Sprintf("projects/'%s'", projectId))

	if err != nil {
		return err
	}

	logLevel = level
	logClient = client

	logger = logClient.Logger(name)

	return nil
}

func levelToSeverity(level Level) logging.Severity {
	switch level {
	case LevelDebug:
		return logging.Debug
	case LevelInfo:
		return logging.Info
	case LevelWarn:
		return logging.Warning
	case LevelError:
		return logging.Error
	default:
		return logging.Default
	}
}

func levelToSLogLevel(level Level) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

func log(level Level, message string, payload Payload) {
	if level >= logLevel {
		// Stackdriver
		if logger != nil {
			defer logger.Flush()

			logger.Log(logging.Entry{
				Payload:  payload,
				Severity: levelToSeverity(level),
			})

			return
		}

		// Console
		attrs := []slog.Attr{}

		for key, value := range payload {
			attr := slog.Attr{
				Key:   key,
				Value: slog.StringValue(value),
			}

			attrs = append(attrs, attr)
		}

		ctx := context.Background()

		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		logger.LogAttrs(ctx, levelToSLogLevel(level), message, attrs...)
	}
}

// Debug
func Debug(message string, payload Payload) {
	log(LevelDebug, message, payload)
}

// Info
func Info(message string, payload Payload) {
	log(LevelInfo, message, payload)
}

// Warn
func Warn(message string, payload Payload) {
	log(LevelWarn, message, payload)
}

// Error
func Error(message string, payload Payload) {
	log(LevelError, message, payload)
}

func Fatal(err error) {
	if err != nil {
		log(LevelError, err.Error(), nil)
	} else {
		log(LevelError, "unknown fatal error...", nil)
	}

	os.Exit(1)
}

// Close
func Close() {
	if logClient != nil {
		logClient.Close()
	}
}
