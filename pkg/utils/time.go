package utils

import "time"

const (
	TimeLayout = "2006-01-02T15:04:05"
)


func ParseTime(s string) (time.Time, error) {
	return time.Parse(TimeLayout, s)
}
