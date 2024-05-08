// Package traefiklogger a Traefik HTTP logger plugin.
package traefiklogger

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
)

// TextualHTTPLogger a textual logger implementation.
type TextualHTTPLogger struct {
	logger *log.Logger
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
		thl.logger.Println("Failed to write:", err)
		return
	}
}

func formatHeaders(header http.Header) string {
	keys := make([]string, 0, len(header))
	for key := range header {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	headers := ""
	for _, key := range keys {
		headers += fmt.Sprintf("%s: %s\n", key, strings.Join(header[key], ","))
	}
	return headers
}
