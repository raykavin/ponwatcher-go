package smart

import (
	"fmt"
	"pon_watcher/pkg/log"
	"time"
)

type SmartZerologContext struct {
	*SmartLogger
}

// Ensure it implements both interfaces
var (
	_ log.Interface = (*SmartZerologContext)(nil)
	_ log.Smart     = (*SmartZerologContext)(nil)
)

// ========== Base Interface Methods ==========

// Print implements Logging.
func (s *SmartZerologContext) Print(args ...any) {
	s.Logger.Print(args...)
}

// Debug implements Logging.
func (s *SmartZerologContext) Debug(args ...any) {
	s.Logger.Debug().Msg(fmt.Sprint(args...))
}

// Info implements Logging.
func (s *SmartZerologContext) Info(args ...any) {
	s.Logger.Info().Msg(fmt.Sprint(args...))
}

// Warn implements Logging.
func (s *SmartZerologContext) Warn(args ...any) {
	s.Logger.Warn().Msg(fmt.Sprint(args...))
}

// Error implements Logging.
func (s *SmartZerologContext) Error(args ...any) {
	s.Logger.Error().Msg(fmt.Sprint(args...))
}

// Fatal implements Logging.
func (s *SmartZerologContext) Fatal(args ...any) {
	s.Logger.Fatal().Msg(fmt.Sprint(args...))
}

// Panic implements Logging.
func (s *SmartZerologContext) Panic(args ...any) {
	s.Logger.Panic().Msg(fmt.Sprint(args...))
}

// Printf implements Logging.
func (s *SmartZerologContext) Printf(format string, args ...any) {
	s.Logger.Printf(format, args...)
}

// Debugf implements Logging.
func (s *SmartZerologContext) Debugf(format string, args ...any) {
	s.Logger.Debug().Msgf(format, args...)
}

// Infof implements Logging.
func (s *SmartZerologContext) Infof(format string, args ...any) {
	s.Logger.Info().Msgf(format, args...)
}

// Warnf implements Logging.
func (s *SmartZerologContext) Warnf(format string, args ...any) {
	s.Logger.Warn().Msgf(format, args...)
}

// Errorf implements Logging.
func (s *SmartZerologContext) Errorf(format string, args ...any) {
	s.Logger.Error().Msgf(format, args...)
}

// Fatalf implements Logging.
func (s *SmartZerologContext) Fatalf(format string, args ...any) {
	s.Logger.Fatal().Msgf(format, args...)
}

// Panicf implements Logging.
func (s *SmartZerologContext) Panicf(format string, args ...any) {
	s.Logger.Panic().Msgf(format, args...)
}

// ========== Context Methods ==========

// WithError implements Logging.
func (s *SmartZerologContext) WithError(err error) log.Interface {
	newLogger := s.With().Err(err).Logger()
	return &SmartZerologContext{&SmartLogger{&newLogger}}
}

// WithField implements Logging.
func (s *SmartZerologContext) WithField(key string, value any) log.Interface {
	newLogger := s.With().Interface(key, value).Logger()
	return &SmartZerologContext{&SmartLogger{&newLogger}}
}

// WithFields implements Logging.
func (s *SmartZerologContext) WithFields(fields map[string]any) log.Interface {
	newLogger := s.With().Fields(fields).Logger()
	return &SmartZerologContext{&SmartLogger{&newLogger}}
}

// ========== Enhanced Smart Methods ==========

// Success implements SmartLogging.
func (s *SmartZerologContext) Success(msg string) {
	s.SmartLogger.Success(msg)
}

// Failure implements SmartLogging.
func (s *SmartZerologContext) Failure(msg string) {
	s.SmartLogger.Failure(msg)
}

// Progress implements SmartLogging.
func (s *SmartZerologContext) Progress(msg string, current, total int) {
	s.SmartLogger.Progress(msg, current, total)
}

// Benchmark implements SmartLogging.
func (s *SmartZerologContext) Benchmark(name string, duration time.Duration) {
	s.SmartLogger.Benchmark(name, duration)
}

// API implements SmartLogging.
func (s *SmartZerologContext) API(method, path string, statusCode int, duration time.Duration) {
	s.SmartLogger.API(method, path, statusCode, duration)
}

// WithContext implements SmartLogging.
func (s *SmartZerologContext) WithContext(ctx map[string]interface{}) log.Smart {
	smartLogger := s.SmartLogger.WithContext(ctx)
	return &SmartZerologContext{smartLogger}
}

// ========== Factory Functions ==========

// NewSmartZerologContext creates a new smart zerolog context adapter
func NewSmartZerologContext(logger *SmartLogger) *SmartZerologContext {
	return &SmartZerologContext{SmartLogger: logger}
}

// NewSmartZerologContextFromConfig creates a new smart zerolog context with configuration
func NewSmartZerologContextFromConfig(level, dateTimeLayout string, colored, jsonFormat bool) (*SmartZerologContext, error) {
	logger, err := New(level, dateTimeLayout, colored, jsonFormat)
	if err != nil {
		return nil, err
	}
	return NewSmartZerologContext(logger), nil
}

// ========== Migration Helper ==========

// LegacyZerologContext provides backward compatibility
type LegacyZerologContext = SmartZerologContext

// NewZerologContext creates a legacy context (for backward compatibility)
// Deprecated: Use NewSmartZerologContext instead
func NewZerologContext(logger *SmartLogger) *LegacyZerologContext {
	return NewSmartZerologContext(logger)
}
