package util

import (
	"encoding/json"
	"errors"
	"strings"
)

//SplitComma - use custom split func instead of .Split to eliminate empty values (i.e Test,,,)
func SplitComma(c rune) bool {
	return c == ','
}

//DeleteEmptyAndTrim removes empty strings from a slice
func DeleteEmptyAndTrim(s []string) []string {
	var r []string
	for _, str := range s {
		str = strings.TrimSpace(str)
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

//CoerceType will accept a string, []byte, or json.Marshaler type and convert it to a []byte for use and consistency in the SDK
func CoerceType(param interface{}) ([]byte, error) {
	var data []byte
	var err error

	switch param.(type) {
	case string:
		input := param.(string)
		data = []byte(input)

	case []byte:
		data = param.([]byte)

	case json.Marshaler:
		marshaler := param.(json.Marshaler)
		data, err = marshaler.MarshalJSON()
		if err != nil {
			return nil, errors.New("marshaling input data to JSON failed")
		}

	default:
		return nil, errors.New("passed in data must be of type []byte, string or implement json.Marshaler")
	}
	return data, nil
}
