package traefiklogger

import "strings"

func containsIgnoreCase(values []string, value string) bool {
	for _, str := range values {
		if strings.Contains(strings.ToLower(str), strings.ToLower(value)) {
			return true
		}
	}
	return false
}

func redact(text string) string {
	if len(text) == 0 {
		return ""
	}
	return "██"
}

func decodeEach(value []string, decoder func(string) (string, error)) ([]string, error) {
	decodedValues := make([]string, len(value))
	for i, v := range value {
		decoded, err := decoder(v)
		if err == nil {
			decodedValues[i] = decoded
		} else {
			return value, err
		}
	}
	return decodedValues, nil
}

func decodeHeaders(value []string, decoder func(string) string) []string {
	decodedValues := make([]string, len(value))
	for i, v := range value {
		decodedValues[i] = decoder(v)
	}
	return decodedValues
}
