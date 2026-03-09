package contracts

// Logger is the contract for logging backends in addon modules
// (e.g. ss-keel-logger).
type Logger interface {
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Debug(format string, args ...interface{})
}
