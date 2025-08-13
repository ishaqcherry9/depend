package utils

import (
	"strconv"
)

const MaxStringID = "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"

func StrToInt(str string) int {
	v, _ := strconv.Atoi(str)
	return v
}

func StrToIntE(str string) (int, error) {
	return strconv.Atoi(str)
}

func StrToUint32(str string) uint32 {
	v, _ := strconv.ParseUint(str, 10, 64)
	return uint32(v)
}

func StrToUint32E(str string) (uint32, error) {
	v, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint32(v), nil
}

func StrToUint64(str string) uint64 {
	v, _ := strconv.ParseUint(str, 10, 64)
	return v
}

func StrToUint64E(str string) (uint64, error) {
	return strconv.ParseUint(str, 10, 64)
}

func StrToUint(str string) uint {
	return uint(StrToUint64(str))
}

func StrToUintE(str string) (uint, error) {
	v, err := StrToUint64E(str)
	return uint(v), err
}

func StrToFloat32(str string) float32 {
	v, _ := strconv.ParseFloat(str, 32)
	return float32(v)
}

func StrToFloat32E(str string) (float32, error) {
	v, err := strconv.ParseFloat(str, 32)
	if err != nil {
		return 0, err
	}
	return float32(v), nil
}

func StrToFloat64(str string) float64 {
	v, _ := strconv.ParseFloat(str, 64)
	return v
}

func StrToFloat64E(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

func IntToStr(v int) string {
	return strconv.Itoa(v)
}

func UintToStr(v uint) string {
	return Uint64ToStr(uint64(v))
}

func Uint64ToStr(v uint64) string {
	return strconv.FormatUint(v, 10)
}

func Int64ToStr(v int64) string {
	return strconv.FormatInt(v, 10)
}

func ProtoInt32ToInt(v int32) int {
	return int(v)
}

func IntToProtoInt32(v int) int32 {
	return int32(v)
}

func ProtoInt64ToUint64(v int64) uint64 {
	return uint64(v)
}
func Uint64ToProtoInt64(v uint64) int64 {
	return int64(v)
}
