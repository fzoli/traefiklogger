package traefiklogger

import (
	"crypto/rand"
	"encoding/hex"
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

	uuid4 := hexExcodingStringConv(bytes)
	return uuid4
}

// Hex encoding of bytes and conversion into string type with proper formatting.
func hexExcodingStringConv(randomBits []byte) string {
	var uuid4 string
	part1 := hex.EncodeToString(randomBits[:4])
	uuid4 = part1
	uuid4 += "-"
	part2 := hex.EncodeToString(randomBits[4:6])
	uuid4 += part2
	uuid4 += "-"
	part3 := hex.EncodeToString(randomBits[6:8])
	uuid4 += part3
	uuid4 += "-"
	part4 := hex.EncodeToString(randomBits[8:10])
	uuid4 += part4
	uuid4 += "-"
	part5 := hex.EncodeToString(randomBits[10:])
	uuid4 += part5
	return uuid4
}
