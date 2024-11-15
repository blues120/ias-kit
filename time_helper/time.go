package time_helper

import "time"

type TIME_FORMAT string

const (
	TIME_FORMAT_DATETIME TIME_FORMAT = "2006-01-02 15:04:05"
	TIME_FORMAT_DATE     TIME_FORMAT = "2006-01-02"
	TIME_FORMAT_MONTH    TIME_FORMAT = "2006-01"
	TIME_FORMAT_YEAR     TIME_FORMAT = "2006"
)

// StringToTimestamp 时间字符串转时间戳(秒)
func StringToTimestamp(t string, f TIME_FORMAT) uint64 {
	format := string(TIME_FORMAT_DATETIME)
	if f != "" {
		format = string(f)
	}
	tvalue, err := time.ParseInLocation(format, t, time.Local)
	if err != nil {
		return 0
	}
	return uint64(tvalue.Unix())
}

// DateTimeToTimestamp 时间字符串转时间戳(秒)
func DateTimeToTimestamp(t string) uint64 {
	return StringToTimestamp(t, TIME_FORMAT_DATETIME)
}

// TimestampToString 时间戳(秒)转字符串
func TimestampToString(t uint64, f TIME_FORMAT) string {
	format := string(TIME_FORMAT_DATETIME)
	if f != "" {
		format = string(f)
	}
	return time.Unix(int64(t), 0).Format(format)
}

// TimestampToDateTime 时间戳(秒)转时间字符串
func TimestampToDateTime(t uint64) string {
	return TimestampToString(t, TIME_FORMAT_DATETIME)
}

// TimeToDateTime 时间转日期字符串
func TimeToDateTime(t time.Time) string {
	return t.Format(string(TIME_FORMAT_DATETIME))
}
