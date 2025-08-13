package krand

import (
	"math/rand"
	"strconv"
	"time"
)

const (
	R_NUM   = 1
	R_UPPER = 2
	R_LOWER = 4
	R_All   = 7
)

var (
	refSlices = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	kinds     = [][]byte{refSlices[0:10], refSlices[10:36], refSlices[0:36], refSlices[36:62], refSlices[36:], refSlices[10:62], refSlices[0:62]}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func String(kind int, size ...int) string {
	return string(Bytes(kind, size...))
}

func Bytes(kind int, bytesLen ...int) []byte {
	if kind > 7 || kind < 1 {
		kind = R_All
	}

	length := 6
	if len(bytesLen) > 0 {
		length = bytesLen[0]
		if length < 1 {
			length = 6
		}
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = kinds[kind-1][rand.Intn(len(kinds[kind-1]))]
	}

	return result
}

func Int(rangeSize ...int) int {
	switch len(rangeSize) {
	case 0:
		return rand.Intn(101)
	case 1:
		return rand.Intn(rangeSize[0] + 1)
	default:
		if rangeSize[0] > rangeSize[1] {
			rangeSize[0], rangeSize[1] = rangeSize[1], rangeSize[0]
		}
		return rand.Intn(rangeSize[1]-rangeSize[0]+1) + rangeSize[0]
	}
}

func Float64(dpLength int, rangeSize ...int) float64 {
	dp := 0.0
	if dpLength > 0 {
		dpmax := 1
		for i := 0; i < dpLength; i++ {
			dpmax *= 10
		}
		dp = float64(rand.Intn(dpmax)) / float64(dpmax)
	}

	switch len(rangeSize) {
	case 0:
		return float64(rand.Intn(100)) + dp
	case 1:
		return float64(rand.Intn(rangeSize[0])) + dp
	default:
		if rangeSize[0] > rangeSize[1] {
			rangeSize[0], rangeSize[1] = rangeSize[1], rangeSize[0]
		}
		return float64(rand.Intn(rangeSize[1]-rangeSize[0])+rangeSize[0]) + dp
	}
}

func NewID() int64 {
	ns := time.Now().UnixMilli() * 1000000
	return ns + rand.Int63n(1000000)
}

func NewStringID() string {
	return strconv.FormatInt(NewID(), 16)
}

func NewSeriesID() string {
	var buf [27]byte
	t := time.Now()

	copy(buf[:], t.Format("20060102150405.000000"))

	random := rand.Intn(1000000)
	buf[21] = '0' + byte(random/100000%10)
	buf[22] = '0' + byte(random/10000%10)
	buf[23] = '0' + byte(random/1000%10)
	buf[24] = '0' + byte(random/100%10)
	buf[25] = '0' + byte(random/10%10)
	buf[26] = '0' + byte(random%10)

	return string(buf[:14]) + string(buf[15:])
}
