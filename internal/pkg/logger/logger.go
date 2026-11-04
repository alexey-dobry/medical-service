package logger

//go:generate mockgen -surce=loger.go -destination=mocks/logger_mock.go -package=mocks

type Logger interface {
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	WithFields(args ...interface{}) Logger
}
