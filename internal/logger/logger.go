package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"

	"github.com/forest-shadow/hm-cached-products/internal/config"
)

type Config struct {
	App    config.AppConfig
	Logger config.LoggerConfig
}

type Logger struct {
	l *zap.SugaredLogger
}

func New(c Config) (*Logger, error) {
	loggerConfig := c.Logger.WithDefaults(c.App.Env)
	if err := loggerConfig.Validate(); err != nil {
		return nil, fmt.Errorf("validate logger config: %w", err)
	}

	level, err := zapcore.ParseLevel(string(loggerConfig.Level))
	if err != nil {
		return nil, fmt.Errorf("parse logger level %q: %w", loggerConfig.Level, err)
	}

	encoder, out, err := buildOutput(loggerConfig.Format)
	if err != nil {
		return nil, err
	}

	l := zap.New(
		zapcore.NewCore(
			encoder,
			zapcore.AddSync(out),
			zap.NewAtomicLevelAt(level),
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	).With(
		zap.String("app_name", c.App.Name),
	)

	return &Logger{l: l.Sugar()}, nil
}

func (l *Logger) Debug(msg string, args ...any) {
	l.l.Debugw(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.l.Infow(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.l.Warnw(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.l.Errorw(msg, args...)
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{l: l.l.With(args...)}
}

func (l *Logger) Sync() error {
	return l.l.Sync()
}

func buildOutput(format config.LogFormat) (zapcore.Encoder, zapcore.WriteSyncer, error) {
	out := zapcore.Lock(os.Stderr)

	switch format {
	case config.LogFormatJSON:
		return zapcore.NewJSONEncoder(jsonEncoderConfig()), out, nil
	case config.LogFormatConsole:
		return zapcore.NewConsoleEncoder(consoleEncoderConfig(term.IsTerminal(int(os.Stderr.Fd())))), out, nil
	default:
		return nil, nil, fmt.Errorf("unsupported logger format %q", format)
	}
}

func jsonEncoderConfig() zapcore.EncoderConfig {
	cfg := baseEncoderConfig()
	cfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder

	return cfg
}

func consoleEncoderConfig(useColor bool) zapcore.EncoderConfig {
	cfg := baseEncoderConfig()
	cfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(applyColor(gruvboxGray, t.Format(time.DateTime), useColor))
	}
	cfg.EncodeCaller = func(c zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(applyColor(gruvboxBlue, c.TrimmedPath(), useColor))
	}
	cfg.EncodeLevel = func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(applyColor(levelColor(level), level.CapitalString(), useColor))
	}

	return cfg
}

func baseEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		CallerKey:      "caller",
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}
