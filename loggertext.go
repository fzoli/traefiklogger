// Package traefiklogger a Traefik HTTP logger plugin.
package traefiklogger

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// TextualHTTPLogger a textual logger implementation.
type TextualHTTPLogger struct {
	logger *log.Logger
	writer LogWriter
}

func (thl *TextualHTTPLogger) print(system string, r *http.Request, mrw *multiResponseWriter, requestHeaders http.Header, requestBody *bytes.Buffer, responseHeaders http.Header) {
	logMessage := fmt.Sprintf("%s %s %s: %d %s %s\n",
		r.RemoteAddr, r.Method, r.URL.String(),
		mrw.status, http.StatusText(mrw.status), r.Proto,
	)

	if len(requestHeaders) > 0 {
		logMessage += "\nRequest Headers:\n" + formatHeaders(requestHeaders)
	}

	if requestBody.Len() > 0 {
		logMessage += "\nRequest Body:\n" + requestBody.String() + "\n"
	}

	if len(responseHeaders) > 0 {
		logMessage += "\nResponse Headers:\n" + formatHeaders(responseHeaders)
	}

	logMessage += fmt.Sprintf("\nResponse Content Length: %d\n", mrw.length)

	if mrw.body.Len() > 0 {
		logMessage += "\nResponse Body:\n" + mrw.body.String() + "\n"
	}

	err := thl.writer.Write(logMessage + "\n")
	if err != nil {
		thl.logger.Println("Failed to write:", err)
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
