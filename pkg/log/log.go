package log

import "time"

type Interface interface {
	// Context methods - returns a logger based off the root logger and decorates it with the given context and arguments.
	WithField(key string, value any) Interface
	WithFields(fields map[string]any) Interface
	WithError(err error) Interface

	// Standard log functions
	Print(args ...any)
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)
	Panic(args ...any)

	// Formatted log functions
	Printf(format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Panicf(format string, args ...any)
}

// Extended interface for smart logging features
type Smart interface {
	Interface // Embed the base interface

	// Enhanced logging methods
	Success(msg string)
	Failure(msg string)
	Progress(msg string, current, total int)
	Benchmark(name string, duration time.Duration)
	API(method, path string, statusCode int, duration time.Duration)

	// Context with map support
	WithContext(ctx map[string]any) Smart
}
