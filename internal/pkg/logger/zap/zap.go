package zap

import (
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger logger.Logger

func init() {
	Logger = NewLogger(Config{})
}

type zapLogger struct {
	logger *zap.SugaredLogger
}

const logDir = "./logs"

type Config struct {
	Dir     string   `yaml:"dir"`
	Targets []string `yaml:"log_targets"`
}

// NewLogger создаёт sugared-логгер zap с общей конфигурацией и пишет в источники из cfg.
func NewLogger(cfg Config) logger.Logger {
	dir := cfg.Dir
	if dir == "" {
		dir = logDir
	}

	if err := os.MkdirAll(dir, 0o750); err != nil {
		log.Fatalf("error creating log dir(%s): %s", dir, err)
	}

	logPath := filepath.Join(dir, "main.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600) //nolint:gosec
	if err != nil {
		log.Fatalf("error opening log file(%s): %s", logPath, err) //nolint:all
	}

	var zcfg zapcore.EncoderConfig
	if os.Getenv("APP_ENV") == "development" {
		zcfg = zap.NewDevelopmentEncoderConfig()
	} else {
		zcfg = zap.NewProductionEncoderConfig()
	}
	// формат времени 2006-01-02T15:04:05.000Z0700
	zcfg.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(zcfg)

	zcfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(zcfg)

	// LOG_LEVEL принимает debug, info, warn, error; по умолчанию info.
	levelEncoder := zapcore.InfoLevel
	if raw := os.Getenv("LOG_LEVEL"); raw != "" {
		if err := levelEncoder.Set(raw); err != nil {
			log.Printf("unknown LOG_LEVEL using info")
			levelEncoder = zapcore.InfoLevel
		}
	}

	cores := []zapcore.Core{
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), levelEncoder),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), levelEncoder),
	}

	for _, target := range cfg.Targets {
		conn, err := net.Dial("tcp", target)
		if err != nil {
			log.Fatalf("error connecting to target(%s): %s", target, err)
		}
		cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(conn), zapcore.DebugLevel))
		log.Printf("connected to logs target(%s)", target)
	}

	core := zapcore.NewTee(cores...)

	return &zapLogger{
		logger: zap.New(core).Sugar(),
	}
}

func (l *zapLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *zapLogger) Infof(template string, args ...interface{}) {
	l.logger.Infof(template, args...)
}

func (l *zapLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *zapLogger) Warnf(template string, args ...interface{}) {
	l.logger.Warnf(template, args...)
}

func (l *zapLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *zapLogger) Errorf(template string, args ...interface{}) {
	l.logger.Errorf(template, args...)
}

func (l *zapLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *zapLogger) Debugf(template string, args ...interface{}) {
	l.logger.Debugf(template, args...)
}

func (l *zapLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *zapLogger) Fatalf(template string, args ...interface{}) {
	l.logger.Fatalf(template, args...)
}

func (l *zapLogger) WithFields(args ...interface{}) logger.Logger {
	zl := zapLogger{
		logger: l.logger.With(args...),
	}
	return &zl
}
