package utils

import (
	"encoding/json"

)

func UnmarshalFromString(str string, v interface{}) error {
	return json.Unmarshal([]byte(str), v)
}

func MarshalToString(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func MarshalE(v interface{}) []byte {
	data, _ := Marshal(v)
	return data
}

func MarshalToStringE(v interface{}) string {
	str, _ := MarshalToString(v)
	return str
}
