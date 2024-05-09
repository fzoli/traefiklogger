package traefiklogger

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// GenerateUUID4 generates RFC 4122 version 4 UUID.
func GenerateUUID4() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}

	bytes[6] = (bytes[6] & 0x0f) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3f) | 0x80 // Variant is 10

	uuid4 := hexEncodingStringConv(bytes[:])
	return uuid4
}

// Hex encoding of bytes and conversion into string type with proper formatting.
func hexEncodingStringConv(randomBits []byte) string {
	var builder strings.Builder
	builder.WriteString(hex.EncodeToString(randomBits[:4]))
	builder.WriteString("-")
	builder.WriteString(hex.EncodeToString(randomBits[4:6]))
	builder.WriteString("-")
	builder.WriteString(hex.EncodeToString(randomBits[6:8]))
	builder.WriteString("-")
	builder.WriteString(hex.EncodeToString(randomBits[8:10]))
	builder.WriteString("-")
	builder.WriteString(hex.EncodeToString(randomBits[10:]))
	return builder.String()
}
