package smart

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

type SmartLogger struct {
	*zerolog.Logger
}

// Color scheme for  output
var (
	// Level colors with gradients and styles
	traceColor = color.New(color.FgHiBlack, color.Bold)
	debugColor = color.New(color.FgHiBlue, color.Bold)
	infoColor  = color.New(color.FgHiGreen, color.Bold)
	warnColor  = color.New(color.FgHiYellow, color.Bold)
	errorColor = color.New(color.FgHiRed, color.Bold)
	fatalColor = color.New(color.FgRed, color.BgWhite, color.Bold)
	panicColor = color.New(color.FgWhite, color.BgRed, color.Bold, color.BlinkSlow)

	// Component colors
	timestampColor = color.New(color.FgHiCyan, color.Italic)
	callerColor    = color.New(color.FgHiMagenta)
	messageColor   = color.New(color.FgWhite)
	fieldKeyColor  = color.New(color.FgHiYellow)
	fieldValColor  = color.New(color.FgCyan)

	// Special symbols and decorations
	symbols = map[string]string{
		"trace": "◇",
		"debug": "◈",
		"info":  "◉",
		"warn":  "◎",
		"error": "✖",
		"fatal": "☠",
		"panic": "💥",
	}
)

func New(level, dateTimeLayout string, colored, jsonFormat bool) (*SmartLogger, error) {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	logMode, err := zerolog.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	zerolog.SetGlobalLevel(logMode)

	var logger zerolog.Logger

	if jsonFormat {
		//  JSON format with proper indentation
		logger = log.Output(zerolog.ConsoleWriter{
			Out:           os.Stdout,
			NoColor:       !colored,
			TimeFormat:    dateTimeLayout,
			PartsOrder:    []string{"time", "level", "caller", "message"},
			FieldsExclude: []string{"caller"},
		})
	} else {
		// Custom  console format
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			NoColor:    !colored,
			TimeFormat: dateTimeLayout,
			PartsOrder: []string{"time", "level", "caller", "message"},
		}

		if colored {
			output.FormatLevel = formatLevel
			output.FormatMessage = formatMessage
			output.FormatCaller = formatCaller
			output.FormatTimestamp = func(i any) string {
				return formatTimestamp(i, dateTimeLayout)
			}
			output.FormatFieldName = formatFieldName
			output.FormatFieldValue = formatFieldValue
		}

		logger = log.Output(output)
	}

	logger = logger.With().
		CallerWithSkipFrameCount(3).
		Logger()

	return &SmartLogger{&logger}, nil
}

func formatLevel(i any) string {
	levelStr, ok := i.(string)
	if !ok {
		return "❓ UNKNOWN"
	}

	symbol := symbols[levelStr]
	if symbol == "" {
		symbol = "◯"
	}

	switch levelStr {
	case zerolog.LevelTraceValue:
		return traceColor.Sprintf(" %s TRACE ", symbol)
	case zerolog.LevelDebugValue:
		return debugColor.Sprintf(" %s DEBUG ", symbol)
	case zerolog.LevelInfoValue:
		return infoColor.Sprintf(" %s INFO  ", symbol)
	case zerolog.LevelWarnValue:
		return warnColor.Sprintf(" %s WARN  ", symbol)
	case zerolog.LevelErrorValue:
		return errorColor.Sprintf(" %s ERROR ", symbol)
	case zerolog.LevelFatalValue:
		return fatalColor.Sprintf(" %s FATAL ", symbol)
	case zerolog.LevelPanicValue:
		return panicColor.Sprintf(" %s PANIC ", symbol)
	default:
		return color.HiWhiteString(" %s UNKNOWN ", symbol)
	}
}

func formatMessage(i any) string {
	const maxSize = 60

	msg, ok := i.(string)
	if !ok || len(msg) == 0 {
		return messageColor.Sprint("│ (empty message)")
	}

	// Add  prefix
	prefix := "│ "

	// Handle multiline messages ly
	if strings.Contains(msg, "\n") {
		lines := strings.Split(msg, "\n")
		var formattedLines []string
		for i, line := range lines {
			if i == 0 {
				formattedLines = append(formattedLines, messageColor.Sprintf("%s%s", prefix, line))
			} else {
				formattedLines = append(formattedLines, messageColor.Sprintf("│ %s", line))
			}
		}
		return strings.Join(formattedLines, "\n")
	}

	// Truncate message ifis greaten of max size
	if len(msg) > maxSize {
		msg = msg[:maxSize]
	}

	if len(msg) < maxSize {
		msg += strings.Repeat(" ", maxSize-len(msg))
	}

	return messageColor.Sprintf("%s%s", prefix, msg)
}

func formatCaller(i any) string {
	const maxFileSize = 22
	const maxLineSize = 4

	fname, ok := i.(string)
	if !ok || len(fname) == 0 {
		return ""
	}

	caller := filepath.Base(fname)
	callerSplit := strings.Split(caller, ":")
	if len(callerSplit) != 2 {
		return callerColor.Sprintf("┤ %s ├", caller)
	}

	fileBase := strings.TrimSuffix(callerSplit[0], ".go")
	line := callerSplit[1]

	if len(fileBase) > maxFileSize {
		fileBase = fileBase[:maxFileSize]
	} else {
		fileBase = fmt.Sprintf("%-*s", maxFileSize, fileBase)
	}

	if len(line) > maxLineSize {
		line = line[len(line)-maxLineSize:]
	} else {
		line = fmt.Sprintf("%0*s", maxLineSize, line)
	}

	return callerColor.Sprintf("┤ %s:%s ├", fileBase, line)
}

func formatTimestamp(i any, timeLayout string) string {
	strTime, ok := i.(string)
	if !ok {
		return timestampColor.Sprintf("[ %v ]", i)
	}

	ts, err := time.ParseInLocation(time.RFC3339, strTime, time.Local)
	if err != nil {
		return timestampColor.Sprintf("[ %s ]", strTime)
	}

	formattedTime := ts.In(time.Local).Format(timeLayout)
	return timestampColor.Sprintf("[ %s ]", formattedTime)
}

func formatFieldName(i any) string {
	name, ok := i.(string)
	if !ok {
		return fmt.Sprintf("%v", i)
	}
	return fieldKeyColor.Sprintf("%s", name)
}

func formatFieldValue(i any) string {
	switch v := i.(type) {
	case string:
		// Quote strings that contain spaces or special chars
		if strings.ContainsAny(v, " \t\n\r") {
			return "=" + fieldValColor.Sprintf("%s", v)
		}
		return fieldValColor.Sprint("=", v)
	case int, int8, int16, int32, int64:
		return fieldValColor.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fieldValColor.Sprintf("%d", v)
	case float32, float64:
		return fieldValColor.Sprintf("%.2f", v)
	case bool:
		if v {
			return color.HiGreenString("true")
		}
		return color.HiRedString("false")
	case nil:
		return color.HiBlackString("null")
	default:
		return fieldValColor.Sprintf("=%v", v)
	}
}

// Enhanced methods for  logging
func (l *SmartLogger) Success(msg string) {
	l.Info().Str("status", "✅ SUCCESS").Msg(msg)
}

func (l *SmartLogger) Failure(msg string) {
	l.Error().Str("status", "❌ FAILURE").Msg(msg)
}

func (l *SmartLogger) Progress(msg string, current, total int) {
	percentage := float64(current) / float64(total) * 100
	progressBar := createProgressBar(int(percentage))
	l.Info().
		Str("progress", progressBar).
		Float64("percent", percentage).
		Int("current", current).
		Int("total", total).
		Msg(msg)
}

func (l *SmartLogger) Benchmark(name string, duration time.Duration) {
	var emoji string
	switch {
	case duration < time.Millisecond:
		emoji = "⚡" // Very fast
	case duration < 10*time.Millisecond:
		emoji = "🚀" // Fast
	case duration < 100*time.Millisecond:
		emoji = "🏃" // Medium
	case duration < time.Second:
		emoji = "🚶" // Slow
	default:
		emoji = "🐌" // Very slow
	}

	l.Debug().
		Str("benchmark", emoji).
		Str("duration", duration.String()).
		Msgf("Benchmark: %s", name)
}

func (l *SmartLogger) API(method, path string, statusCode int, duration time.Duration) {
	var level zerolog.Level
	var statusEmoji string

	switch {
	case statusCode >= 200 && statusCode < 300:
		level = zerolog.InfoLevel
		statusEmoji = "✅"
	case statusCode >= 300 && statusCode < 400:
		level = zerolog.InfoLevel
		statusEmoji = "🔄"
	case statusCode >= 400 && statusCode < 500:
		level = zerolog.WarnLevel
		statusEmoji = "⚠️"
	case statusCode >= 500:
		level = zerolog.ErrorLevel
		statusEmoji = "❌"
	default:
		level = zerolog.InfoLevel
		statusEmoji = "❓"
	}

	l.WithLevel(level).
		Str("method", method).
		Str("path", path).
		Int("status_code", statusCode).
		Str("status", statusEmoji).
		Str("duration", duration.Round(time.Second).String()).
		Msg("API Request")
}

func createProgressBar(percentage int) string {
	const barLength = 20
	filled := percentage * barLength / 100

	bar := "["
	for i := 0; i < barLength; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += fmt.Sprintf("] %d%%", percentage)

	return bar
}

// Helper method to create a logger with context
func (l *SmartLogger) WithContext(ctx map[string]any) *SmartLogger {
	event := l.With()
	for k, v := range ctx {
		event = event.Interface(k, v)
	}
	logger := event.Logger()
	return &SmartLogger{&logger}
}
