// Package traefiklogger a Traefik HTTP logger plugin.
package traefiklogger

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

// TextualHTTPLogger a textual logger implementation.
type TextualHTTPLogger struct {
	logger *zap.Logger
	writer LogWriter
}

func (thl *TextualHTTPLogger) print(record *LogRecord) {
	logMessage := fmt.Sprintf("%s %s %s: %d %s %s\n",
		record.RemoteAddr, record.Method, record.URL,
		record.StatusCode, http.StatusText(record.StatusCode), record.Proto,
	)

	if len(record.RequestHeaders) > 0 {
		logMessage += "\nRequest Headers:\n" + formatHeaders(record.RequestHeaders)
	}

	if record.RequestBody.Len() > 0 {
		logMessage += "\nRequest Body:\n" + record.RequestBody.String() + "\n"
	}

	if len(record.ResponseHeaders) > 0 {
		logMessage += "\nResponse Headers:\n" + formatHeaders(record.ResponseHeaders)
	}

	logMessage += fmt.Sprintf("\nResponse Content Length: %d\n", record.ResponseContentLength)

	if record.ResponseBody.Len() > 0 {
		logMessage += "\nResponse Body:\n" + record.ResponseBody.String() + "\n"
	}

	err := thl.writer.Write(logMessage + "\n")
	if err != nil {
		thl.logger.Error("Failed to write", zap.Error(err))
		return
	}
}

func formatHeaders(header http.Header) string {
	headers := ""
	for key, values := range header {
		headers += fmt.Sprintf("%s: %s\n", key, strings.Join(values, ","))
	}
	return headers
}
