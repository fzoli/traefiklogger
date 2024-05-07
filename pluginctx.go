// Package traefiklogger a Traefik HTTP logger plugin.
package traefiklogger

import (
	"context"
	"log"
	"os"
	"time"
)

type clockContextKey string

// ClockContextKey can be used to fake time.
const ClockContextKey clockContextKey = "clock"

// LoggerClock is the source of current time.
type LoggerClock interface {
	Now() time.Time
}

// SystemLoggerClock uses OS system time.
type SystemLoggerClock struct{}

// Now returns current OS system time.
func (*SystemLoggerClock) Now() time.Time {
	return time.Now()
}

type logWriterContextKey string

// LogWriterContextKey can be used to spy log writes.
const LogWriterContextKey logWriterContextKey = "log-writer"

// LogWriter is a write strategy.
type LogWriter interface {
	Write(log string) error
}

// FileLogWriter writes logs to a File (like stdout).
type FileLogWriter struct {
	file *os.File
}

func (w *FileLogWriter) Write(log string) error {
	_, err := w.file.WriteString(log)
	return err
}

// LoggerLogWriter writes logs to a Logger.
type LoggerLogWriter struct {
	logger *log.Logger
}

func (w *LoggerLogWriter) Write(log string) error {
	w.logger.Print(log)
	return nil
}

func createTextualHTTPLogger(ctx context.Context, logger *log.Logger) *TextualHTTPLogger {
	externalLogWriter, hasExternalLogWriter := ctx.Value(LogWriterContextKey).(LogWriter)
	if hasExternalLogWriter {
		return &TextualHTTPLogger{logger: logger, writer: externalLogWriter}
	}
	return &TextualHTTPLogger{logger: logger, writer: &LoggerLogWriter{logger: logger}}
}

func createJSONHTTPLogger(ctx context.Context, logger *log.Logger) *JSONHTTPLogger {
	var clock LoggerClock
	externalClock, hasExternalClock := ctx.Value(ClockContextKey).(LoggerClock)
	if hasExternalClock {
		clock = externalClock
	} else {
		clock = &SystemLoggerClock{}
	}
	externalLogWriter, hasExternalLogWriter := ctx.Value(LogWriterContextKey).(LogWriter)
	if hasExternalLogWriter {
		return &JSONHTTPLogger{clock: clock, logger: logger, writer: externalLogWriter}
	}
	return &JSONHTTPLogger{clock: clock, logger: logger, writer: &FileLogWriter{file: os.Stdout}}
}
