package utils

import (
	"strconv"
	"testing"
	"time"
)

func TestNewTimeUtil(t *testing.T) {
	// Test case 1: Valid timezone
	tu, err := NewTimeUtil("Asia/Shanghai")
	if err != nil {
		t.Fatalf("NewTimeUtil failed: %v", err)
	}
	if tu.location.String() != "Asia/Shanghai" {
		t.Errorf("Timezone not initialized correctly, got %s", tu.location.String())
	}

	// Test case 2: Invalid timezone
	_, err = NewTimeUtil("Invalid/Timezone")
	if err == nil {
		t.Errorf("Expected error for invalid timezone")
	}
}

func TestTimeUtil_Now(t *testing.T) {
	tu, _ := NewTimeUtil("Asia/Shanghai")
	now := tu.Now()
	location, _ := time.LoadLocation("Asia/Shanghai")
	if now.Location().String() != location.String() {
		t.Errorf("Now() returned time with incorrect timezone, got %s, expected %s", now.Location().String(), location.String())
	}
}

func TestTimeUtil_FormatTime(t *testing.T) {
	tu, _ := NewTimeUtil("Asia/Shanghai")
	testTime := time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	formattedTime := tu.FormatTime(testTime, "2006-01-02 15:04:05")
	if formattedTime != "2024-01-02 23:04:05" { // +8 hours for Asia/Shanghai
		t.Errorf("FormatTime returned incorrect format, got %s", formattedTime)
	}
}

func TestTimeUtil_ParseTime(t *testing.T) {
	tu, _ := NewTimeUtil("Asia/Shanghai")
	timeStr := "2024-01-02 15:04:05"
	parsedTime, err := tu.ParseTime("2006-01-02 15:04:05", timeStr)
	if err != nil {
		t.Fatalf("ParseTime failed: %v", err)
	}
	expectedTime := time.Date(2024, 1, 2, 15, 4, 5, 0, tu.location)
	location, _ := time.LoadLocation("Asia/Shanghai")
	expectedTime = expectedTime.In(location)
	if !parsedTime.Equal(expectedTime) {
		t.Errorf("ParseTime returned incorrect time, got %v, expected %v", parsedTime, expectedTime)
	}
}

func TestTimeUtil_ToTimestamp(t *testing.T) {
	tu, _ := NewTimeUtil("Asia/Shanghai")
	testTime := time.Date(2024, 1, 2, 15, 4, 5, 123456789, time.UTC)

	// Test case 1: Seconds
	timestampS := tu.ToTimestamp(testTime, "s")
	expectedS := strconv.FormatInt(testTime.Unix(), 10)
	if timestampS != expectedS {
		t.Errorf("ToTimestamp (seconds) returned incorrect value, got %s, expected %s", timestampS, expectedS)
	}

	// Test case 2: Milliseconds
	timestampMS := tu.ToTimestamp(testTime, "ms")
	expectedMS := strconv.FormatInt(testTime.UnixMilli(), 10)
	if timestampMS != expectedMS {
		t.Errorf("ToTimestamp (milliseconds) returned incorrect value, got %s, expected %s", timestampMS, expectedMS)
	}

	// Test case 3: Microseconds
	timestampWS := tu.ToTimestamp(testTime, "ws")
	expectedWS := strconv.FormatInt(testTime.UnixMicro(), 10)
	if timestampWS != expectedWS {
		t.Errorf("ToTimestamp (microseconds) returned incorrect value, got %s, expected %s", timestampWS, expectedWS)
	}

	// Test case 4: Nanoseconds
	timestampNS := tu.ToTimestamp(testTime, "ns")
	expectedNS := strconv.FormatInt(testTime.UnixNano(), 10)
	if timestampNS != expectedNS {
		t.Errorf("ToTimestamp (nanoseconds) returned incorrect value, got %s, expected %s", timestampNS, expectedNS)
	}

	// Test case 5: Default (seconds)
	timestampDefault := tu.ToTimestamp(testTime, "")
	if timestampDefault != expectedS {
		t.Errorf("ToTimestamp (default) returned incorrect value, got %s, expected %s", timestampDefault, expectedS)
	}
}

func TestTimeUtil_TimeFromTimestamp(t *testing.T) {
	tu, _ := NewTimeUtil("Asia/Shanghai")

	// Define a test time
	testTime := time.Date(2024, 1, 2, 15, 4, 5, 123456789, time.UTC)
	location, _ := time.LoadLocation("Asia/Shanghai")
	testTime = testTime.In(location)

	// Test case 1: Seconds
	timestampS := strconv.FormatInt(testTime.Unix(), 10)
	timeS, err := tu.TimeFromTimestamp(timestampS, "s")
	if err != nil {
		t.Fatalf("TimeFromTimestamp (seconds) failed: %v", err)
	}
	expectedTimeS := time.Unix(testTime.Unix(), 0).In(location)
	if !timeS.Equal(expectedTimeS) {
		t.Errorf("TimeFromTimestamp (seconds) returned incorrect time, got %v, expected %v", timeS, expectedTimeS)
	}

	// Test case 2: Milliseconds
	timestampMS := strconv.FormatInt(testTime.UnixMilli(), 10)
	timeMS, err := tu.TimeFromTimestamp(timestampMS, "ms")
	if err != nil {
		t.Fatalf("TimeFromTimestamp (milliseconds) failed: %v", err)
	}
	expectedTimeMS := time.Unix(testTime.Unix(), int64((testTime.Nanosecond()/1000000)*1000000)).In(location)
	if !timeMS.Equal(expectedTimeMS) {
		t.Errorf("TimeFromTimestamp (milliseconds) returned incorrect time, got %v, expected %v", timeMS, expectedTimeMS)
	}

	// Test case 3: Microseconds
	timestampWS := strconv.FormatInt(testTime.UnixMicro(), 10)
	timeWS, err := tu.TimeFromTimestamp(timestampWS, "ws")
	if err != nil {
		t.Fatalf("TimeFromTimestamp (microseconds) failed: %v", err)
	}
	expectedTimeWS := time.Unix(testTime.Unix(), int64((testTime.Nanosecond()/1000)*1000)).In(location)

	if !timeWS.Equal(expectedTimeWS) {
		t.Errorf("TimeFromTimestamp (microseconds) returned incorrect time, got %v, expected %v", timeWS, expectedTimeWS)
	}

	// Test case 4: Nanoseconds
	timestampNS := strconv.FormatInt(testTime.UnixNano(), 10)
	timeNS, err := tu.TimeFromTimestamp(timestampNS, "ns")
	if err != nil {
		t.Fatalf("TimeFromTimestamp (nanoseconds) failed: %v", err)
	}
	expectedTimeNS := testTime
	if !timeNS.Equal(expectedTimeNS) {
		t.Errorf("TimeFromTimestamp (nanoseconds) returned incorrect time, got %v, expected %v", timeNS, expectedTimeNS)
	}

	// Test case 5: Default (seconds)
	timestampDefault := strconv.FormatInt(testTime.Unix(), 10)
	timeDefault, err := tu.TimeFromTimestamp(timestampDefault, "")
	if err != nil {
		t.Fatalf("TimeFromTimestamp (default) failed: %v", err)
	}
	expectedTimeDefault := time.Unix(testTime.Unix(), 0).In(location)

	if !timeDefault.Equal(expectedTimeDefault) {
		t.Errorf("TimeFromTimestamp (default) returned incorrect time, got %v, expected %v", timeDefault, expectedTimeDefault)
	}

	// Test case 6: Invalid timestamp string
	_, err = tu.TimeFromTimestamp("invalid", "ms")
	if err == nil {
		t.Errorf("Expected error for invalid timestamp string")
	}

	// Test case 7: Unsupported precision
	_, err = tu.TimeFromTimestamp(timestampS, "invalid")
	if err == nil {
		t.Errorf("Expected error for unsupported precision")
	}
}

func TestTimeUtil_BeginningOfDay(t *testing.T) {
	tu, _ := NewTimeUtil("Asia/Shanghai")
	testTime := time.Date(2024, 1, 2, 15, 4, 5, 123456789, time.UTC)
	beginningOfDay := tu.BeginningOfDay(testTime)
	location, _ := time.LoadLocation("Asia/Shanghai")
	expectedBeginningOfDay := time.Date(2024, 1, 2, 0, 0, 0, 0, location)
	if !beginningOfDay.Equal(expectedBeginningOfDay) {
		t.Errorf("BeginningOfDay returned incorrect time, got %v, expected %v", beginningOfDay, expectedBeginningOfDay)
	}
}

func TestTimeUtil_EndOfDay(t *testing.T) {
	tu, _ := NewTimeUtil("Asia/Shanghai")
	testTime := time.Date(2024, 1, 2, 15, 4, 5, 123456789, time.UTC)
	endOfDay := tu.EndOfDay(testTime)
	location, _ := time.LoadLocation("Asia/Shanghai")
	expectedEndOfDay := time.Date(2024, 1, 2, 23, 59, 59, int(time.Second-time.Nanosecond), location)
	if !endOfDay.Equal(expectedEndOfDay) {
		t.Errorf("EndOfDay returned incorrect time, got %v, expected %v", endOfDay, expectedEndOfDay)
	}
}

func TestTimeUtil_LastDayOfMonth(t *testing.T) {
	tu, _ := NewTimeUtil("Asia/Shanghai")
	// Test case 1: Regular month
	testTime := time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	lastDayOfMonth := tu.LastDayOfMonth(testTime)
	location, _ := time.LoadLocation("Asia/Shanghai")
	expectedLastDayOfMonth := time.Date(2024, 1, 31, 23, 59, 59, int(time.Second-time.Nanosecond), location)
	if !lastDayOfMonth.Equal(expectedLastDayOfMonth) {
		t.Errorf("LastDayOfMonth returned incorrect time, got %v, expected %v", lastDayOfMonth, expectedLastDayOfMonth)
	}

	// Test case 2: February in a leap year
	testTime = time.Date(2024, 2, 2, 15, 4, 5, 0, time.UTC)
	lastDayOfMonth = tu.LastDayOfMonth(testTime)
	expectedLastDayOfMonth = time.Date(2024, 2, 29, 23, 59, 59, int(time.Second-time.Nanosecond), location)
	if !lastDayOfMonth.Equal(expectedLastDayOfMonth) {
		t.Errorf("LastDayOfMonth returned incorrect time, got %v, expected %v", lastDayOfMonth, expectedLastDayOfMonth)
	}

	// Test case 3: February in a non-leap year
	testTime = time.Date(2023, 2, 2, 15, 4, 5, 0, time.UTC)
	lastDayOfMonth = tu.LastDayOfMonth(testTime)
	expectedLastDayOfMonth = time.Date(2023, 2, 28, 23, 59, 59, int(time.Second-time.Nanosecond), location)
	if !lastDayOfMonth.Equal(expectedLastDayOfMonth) {
		t.Errorf("LastDayOfMonth returned incorrect time, got %v, expected %v", lastDayOfMonth, expectedLastDayOfMonth)
	}
}

func TestTimeUtil_FirstDayOfNextMonth(t *testing.T) {
	tu, _ := NewTimeUtil("Asia/Shanghai")
	testTime := time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	firstDayOfNextMonth := tu.FirstDayOfNextMonth(testTime)
	location, _ := time.LoadLocation("Asia/Shanghai")
	expectedFirstDayOfNextMonth := time.Date(2024, 2, 1, 0, 0, 0, 0, location)
	if !firstDayOfNextMonth.Equal(expectedFirstDayOfNextMonth) {
		t.Errorf("FirstDayOfNextMonth returned incorrect time, got %v, expected %v", firstDayOfNextMonth, expectedFirstDayOfNextMonth)
	}
}

func TestTimeUtil_AddDays(t *testing.T) {
	tu, _ := NewTimeUtil("Asia/Shanghai")
	testTime := time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	addedTime := tu.AddDays(testTime, 5)
	location, _ := time.LoadLocation("Asia/Shanghai")
	expectedAddedTime := time.Date(2024, 1, 7, 15, 4, 5, 0, time.UTC)
	expectedAddedTime = expectedAddedTime.In(location)
	if !addedTime.Equal(expectedAddedTime) {
		t.Errorf("AddDays returned incorrect time, got %v, expected %v", addedTime, expectedAddedTime)
	}
}

func TestTimeUtil_AddMonths(t *testing.T) {
	tu, _ := NewTimeUtil("Asia/Shanghai")
	testTime := time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	addedTime := tu.AddMonths(testTime, 2)
	location, _ := time.LoadLocation("Asia/Shanghai")
	expectedAddedTime := time.Date(2024, 3, 2, 15, 4, 5, 0, time.UTC)
	expectedAddedTime = expectedAddedTime.In(location)
	if !addedTime.Equal(expectedAddedTime) {
		t.Errorf("AddMonths returned incorrect time, got %v, expected %v", addedTime, expectedAddedTime)
	}
}

func TestTimeUtil_AddYears(t *testing.T) {
	tu, _ := NewTimeUtil("Asia/Shanghai")
	testTime := time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)
	addedTime := tu.AddYears(testTime, 3)
	location, _ := time.LoadLocation("Asia/Shanghai")
	expectedAddedTime := time.Date(2027, 1, 2, 15, 4, 5, 0, time.UTC)
	expectedAddedTime = expectedAddedTime.In(location)
	if !addedTime.Equal(expectedAddedTime) {
		t.Errorf("AddYears returned incorrect time, got %v, expected %v", addedTime, expectedAddedTime)
	}
}

func TestTimeUtil_DiffDays(t *testing.T) {
	tu, _ := NewTimeUtil("Asia/Shanghai")

	// Test case 1: t1 > t2
	t1 := time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	diff := tu.DiffDays(t1, t2)
	if diff != 3 {
		t.Errorf("DiffDays returned incorrect value, got %d, expected 3", diff)
	}

	// Test case 2: t1 < t2
	t1 = time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	t2 = time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC)
	diff = tu.DiffDays(t1, t2)
	if diff != 3 {
		t.Errorf("DiffDays returned incorrect value, got %d, expected 3", diff)
	}

	// Test case 3: t1 == t2
	t1 = time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	t2 = time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	diff = tu.DiffDays(t1, t2)
	if diff != 0 {
		t.Errorf("DiffDays returned incorrect value, got %d, expected 0", diff)
	}

	// Test case 4: Different years
	t1 = time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)
	t2 = time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	diff = tu.DiffDays(t1, t2)
	if diff != 366 { // 2024 is a leap year
		t.Errorf("DiffDays returned incorrect value, got %d, expected 366", diff)
	}

	// Test case 5: Times within the same day
	t1 = time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	t2 = time.Date(2024, 1, 2, 15, 0, 0, 0, time.UTC)
	diff = tu.DiffDays(t1, t2)
	if diff != 0 {
		t.Errorf("DiffDays returned incorrect value, got %d, expected 0", diff)
	}

	// Test case 6: Large difference in days
	t1 = time.Date(2034, 1, 2, 12, 0, 0, 0, time.UTC)
	t2 = time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	diff = tu.DiffDays(t1, t2)
	expectedDiff := 3653
	if diff != expectedDiff {
		t.Errorf("DiffDays returned incorrect value, got %d, expected %d", diff, expectedDiff)
	}
}
