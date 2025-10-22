package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TimeUtil 封装了时间相关的常用操作，并考虑了时区问题
type TimeUtil struct {
	location *time.Location // 时区
}

// NewTimeUtil 创建一个新的 TimeUtil 实例，需要传入时区名称
// 例如: NewTimeUtil("Asia/Shanghai")
func NewTimeUtil(timeZone string) (*TimeUtil, error) {
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, fmt.Errorf("加载时区失败: %w", err)
	}
	return &TimeUtil{location: loc}, nil
}

// Now 返回当前时间的 time.Time 对象，时区为 TimeUtil 实例指定的时区
func (tu *TimeUtil) Now() time.Time {
	return time.Now().In(tu.location)
}

// FormatTime 将 time.Time 对象格式化为字符串
// layout: 格式化字符串，例如 "2006-01-02 15:04:05"
func (tu *TimeUtil) FormatTime(t time.Time, layout string) string {
	return t.In(tu.location).Format(layout)
}

// ParseTime 将字符串解析为 time.Time 对象，时区为 TimeUtil 实例指定的时区
// layout: 格式化字符串，例如 "2006-01-02 15:04:05"
// timeStr: 时间字符串
func (tu *TimeUtil) ParseTime(layout string, timeStr string) (time.Time, error) {
	return time.ParseInLocation(layout, timeStr, tu.location)
}

// ToTimestamp 将 time.Time 对象转换为指定精度的 Unix 时间戳 (string类型)
// precision: 时间戳精度，支持 "s" (秒), "ms" (毫秒), "ws" (微秒), "ns" (纳秒)
func (tu *TimeUtil) ToTimestamp(t time.Time, precision string) string {
	precision = strings.ToLower(precision)
	switch precision {
	case "ms":
		return strconv.FormatInt(t.UnixMilli(), 10)
	case "ws":
		return strconv.FormatInt(t.UnixMicro(), 10)
	case "ns":
		return strconv.FormatInt(t.UnixNano(), 10)
	default:
		return strconv.FormatInt(t.Unix(), 10)
	}
}

// TimeFromTimestamp 将指定精度的 Unix 时间戳字符串转换为 time.Time 对象
// timestampStr: 时间戳字符串
// precision: 时间戳精度，支持 "s" (秒), "ms" (毫秒), "ws" (微秒), "ns" (纳秒)
func (tu *TimeUtil) TimeFromTimestamp(timestampStr string, precision string) (time.Time, error) {
	precision = strings.ToLower(precision)
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析时间戳字符串 %s 失败: %w", timestampStr, err)
	}

	var sec int64
	var nsec int64

	switch precision {
	case "s", "": // 默认秒
		sec = timestamp
	case "ms":
		sec = timestamp / 1000
		nsec = (timestamp % 1000) * int64(time.Millisecond)
	case "ws":
		sec = timestamp / 1000000
		nsec = (timestamp % 1000000) * int64(time.Microsecond)
	case "ns":
		nsec = timestamp
	default:
		return time.Time{}, fmt.Errorf("不支持的时间戳精度: %s, 使用默认精度 (秒)\n", precision)
	}
	return time.Unix(sec, nsec).In(tu.location), nil
}

// BeginningOfDay 返回指定时间所在日期的零点 (00:00:00)，时区为 TimeUtil 实例指定的时区
func (tu *TimeUtil) BeginningOfDay(t time.Time) time.Time {
	year, month, day := t.In(tu.location).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, tu.location)
}

// EndOfDay 返回指定时间所在日期的最后一刻 (23:59:59)，时区为 TimeUtil 实例指定的时区
func (tu *TimeUtil) EndOfDay(t time.Time) time.Time {
	year, month, day := t.In(tu.location).Date()
	return time.Date(year, month, day, 23, 59, 59, int(time.Second-time.Nanosecond), tu.location)
}

// LastDayOfMonth 返回指定时间所在月份的最后一天的最后一刻 (23:59:59)
func (tu *TimeUtil) LastDayOfMonth(t time.Time) time.Time {
	nextMonth := tu.AddMonths(t, 1)
	firstDayOfNextMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, tu.location)
	return tu.EndOfDay(firstDayOfNextMonth.AddDate(0, 0, -1))
}

// FirstDayOfNextMonth 返回下个月的第一天的零点 (00:00:00)
func (tu *TimeUtil) FirstDayOfNextMonth(t time.Time) time.Time {
	nextMonth := tu.AddMonths(t, 1)
	return time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, tu.location)
}

// AddDays 返回指定时间加上指定天数后的时间，时区为 TimeUtil 实例指定的时区
func (tu *TimeUtil) AddDays(t time.Time, days int) time.Time {
	return t.In(tu.location).AddDate(0, 0, days)
}

// AddMonths 返回指定时间加上指定月份后的时间，时区为 TimeUtil 实例指定的时区
func (tu *TimeUtil) AddMonths(t time.Time, months int) time.Time {
	return t.In(tu.location).AddDate(0, months, 0)
}

// AddYears 返回指定时间加上指定年份后的时间，时区为 TimeUtil 实例指定的时区
func (tu *TimeUtil) AddYears(t time.Time, years int) time.Time {
	return t.In(tu.location).AddDate(years, 0, 0)
}

// DiffDays 返回两个时间之间的天数差 (取绝对值)
func (tu *TimeUtil) DiffDays(t1 time.Time, t2 time.Time) int {
	t1 = tu.BeginningOfDay(t1)
	t2 = tu.BeginningOfDay(t2)
	diff := t1.Sub(t2).Hours() / 24
	if diff < 0 {
		return int(-diff)
	}
	return int(diff)
}
