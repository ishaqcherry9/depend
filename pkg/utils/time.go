package utils

import "time"

const (
	DateTimeLayout = "2006-01-02 15:04:05"

	DateTimeLayoutWithMS = "2006-01-02 15:04:05.000"

	RFC3339 = "2006-01-02T15:04:05Z07:00"

	DateTimeLayoutWithMSAndTZ = "2006-01-02T15:04:05.000Z"

	TimeLayout = "15:04:05"

	DateLayout = "2006-01-02"
)

func FormatDateTimeLayout(t time.Time) string {
	return t.Format(DateTimeLayout)
}

func ParseDateTimeLayout(s string) (time.Time, error) {
	return time.Parse(DateTimeLayout, s)
}

func FormatDateTimeLayoutWithMS(t time.Time) string {
	return t.Format(DateTimeLayoutWithMS)
}

func ParseDateTimeLayoutWithMS(s string) (time.Time, error) {
	return time.Parse(DateTimeLayoutWithMS, s)
}

func FormatDateTimeRFC3339(t time.Time) string {
	return t.Format(RFC3339)
}

func ParseDateTimeRFC3339(s string) (time.Time, error) {
	return time.Parse(RFC3339, s)
}

func FormatDateTimeLayoutWithMSAndTZ(t time.Time) string {
	return t.Format(DateTimeLayoutWithMSAndTZ)
}

func ParseDateTimeLayoutWithMSAndTZ(s string) (time.Time, error) {
	return time.Parse(DateTimeLayoutWithMSAndTZ, s)
}
