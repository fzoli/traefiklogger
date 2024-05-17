package traefiklogger

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
)

// HTTPBodyDecoderFactory selects which decoder should run.
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

// GZipHTTPDecoder extracts the gzip content.
type GZipHTTPDecoder struct {
	logger *log.Logger
}

func (d *GZipHTTPDecoder) decode(content *bytes.Buffer) (string, error) {
	gzReader, err := gzip.NewReader(content)
	if err != nil {
		return "", err
	}
	defer func() {
		closeError := gzReader.Close()
		if closeError != nil {
			d.logger.Printf("Failed to close gzip reader: %s", closeError)
		}
	}()
	result, err := io.ReadAll(gzReader)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
