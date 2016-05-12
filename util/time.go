package util

import "time"

// 所有都是本地时间

const timeFormat = "2006-01-02 15:04:05"

func ConvertStrTime(strTime string) (time.Time, bool) {
	t, err := time.ParseInLocation(timeFormat, strTime, time.Local)
	return t, err == nil // if error, time will return "0001-01-01 00:00:00"
}

func FormatTime(t time.Time) string {
	return t.Format(timeFormat)
}

func NowTimeString() string {
	return time.Now().Format(timeFormat)
}

func NowTime() time.Time {
	return time.Now()
}

func IsToday(t time.Time) bool {
	nowTime := NowTime()
	return t.Year() == nowTime.Year() && t.Month() == nowTime.Month() && t.Day() == nowTime.Day()
}

func AddTime(dueTime time.Time, dur time.Duration) time.Time {
	nowTime := NowTime()
	if dueTime.Before(nowTime) {
		dueTime = nowTime
	}
	return dueTime.Add(dur)
}
