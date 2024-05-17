package traefiklogger

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
)

type HTTPBodyDecoderFactory struct {
	rawDecoder  *RawHTTPDecoder
	gzipDecoder *GZipHTTPDecoder
}

func (f *HTTPBodyDecoderFactory) create(encoding string) HTTPBodyDecoder {
	if encoding == "gzip" {
		return f.gzipDecoder
	}
	return f.rawDecoder
}

func createHTTPBodyDecoderFactory(logger *log.Logger) *HTTPBodyDecoderFactory {
	return &HTTPBodyDecoderFactory{
		rawDecoder:  &RawHTTPDecoder{},
		gzipDecoder: &GZipHTTPDecoder{logger: logger},
	}
}

// HTTPBodyDecoder a body decoder strategy.
type HTTPBodyDecoder interface {
	// decodes the content
	decode(content *bytes.Buffer) (string, error)
}

// RawHTTPDecoder just returns the content as-is.
type RawHTTPDecoder struct{}

func (d *RawHTTPDecoder) decode(content *bytes.Buffer) (string, error) {
	return content.String(), nil
}

type GZipHTTPDecoder struct {
	logger *log.Logger
}

func (d *GZipHTTPDecoder) decode(content *bytes.Buffer) (string, error) {
	gz, err := gzip.NewReader(content)
	if err != nil {
		return "", err
	}
	defer func() {
		err := gz.Close()
		if err != nil {
			d.logger.Printf("Failed to close gzip reader: %s", err)
		}
	}()
	result, err := io.ReadAll(gz)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
