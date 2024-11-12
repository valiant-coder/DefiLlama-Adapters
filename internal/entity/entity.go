package entity

import (
	"strconv"
	"time"
)

type Time time.Time

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var err error
	tt, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}
	*t = Time(time.Unix(tt, 0))
	return nil
}



type NumString string
// if empty str return "0"
func (n NumString) MarshalJSON() ([]byte, error) {
	if n == "" {
		return []byte(`"0"`), nil
	}
	return []byte(`"` + string(n) + `"`), nil
}

func (n *NumString) UnmarshalJSON(b []byte) error {
	*n = NumString(string(b))
	return nil
}


type RespNoContent struct{
	Ok bool `json:"ok"`
}
