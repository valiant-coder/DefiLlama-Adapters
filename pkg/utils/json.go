package utils

import "github.com/bytedance/sonic"



func UnmarshalFromString(str string, v interface{}) error {
	return sonic.UnmarshalString(str, v)
}

func MarshalToString(v interface{}) (string, error) {
	return sonic.MarshalString(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return sonic.Unmarshal(data, v)
}

func Marshal(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}

func MarshalE(v interface{}) []byte {
	data, _ := Marshal(v)
	return data
}


func MarshalToStringE(v interface{}) string {
	str, _ := MarshalToString(v)
	return str
}