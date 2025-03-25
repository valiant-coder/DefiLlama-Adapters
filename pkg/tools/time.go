package tools

import (
	"time"
)

const (
	Beijing int64 = 8
)

var (
	Location = time.FixedZone("CST", 8*3600)
)

func GetBeijingDayTime() int64 {
	
	return ToBeijingDayTime(time.Now().Unix())
}

func ToBeijingDayTime(value int64) int64 {
	
	return ToDayTimestamp(value, Beijing)
}

func ToDayTimestamp(value, timeZone int64) int64 {
	interval := GetTimeZoneInterval(timeZone)
	
	value = value + interval
	return value - (value % 86400) - interval
}

func GetTimeZoneInterval(timeZone int64) int64 {
	
	return timeZone * 3600
}

func GetNowHourTime() int64 {
	
	return GetHourTime(time.Now().Unix())
}

func GetHourTime(value int64) int64 {
	
	return value - (value % 3600)
}

func GetNowHalfHourTime() int64 {
	
	return GetHalfHourTime(time.Now().Unix())
}

func GetHalfHourTime(value int64) int64 {
	
	return value - (value % 1800)
}

func GetHfTime() time.Time {
	
	// 获取当前时间
	currentTime := time.Now()
	
	// 计算当前半小时的开始时间
	return currentTime.Truncate(30 * time.Minute).UTC()
}

func GetBeijingDayStart() time.Time {
	
	now := time.Now().In(Location)
	
	// 使用 time.Date 创建一个新的时间对象，将时、分、秒和纳秒设置为0
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	// 使用 time.Date 创建一个新的时间对象，将时、分、秒和纳秒设置为0
	// midnight = midnight.In(Location)
	return midnight
	
	// return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, midnight.Location())
}

func GetBeijingTime(timestamp int64) time.Time {
	
	return time.Unix(timestamp, 0).In(Location)
}

func FormatBeijingTime(timestamp int64) string {
	
	return GetBeijingTime(timestamp).Format("2006-01-02 15:04:05")
}
