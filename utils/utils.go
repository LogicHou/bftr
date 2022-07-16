package utils

import (
	"math"
	"strconv"
	"time"
)

func MsToTime(ms int64) string {
	tm := time.Unix(0, ms*int64(time.Millisecond))
	return tm.Format("2006-01-02 15:04:05.000")
}

func MsToDateTime(ms int64) string {
	tm := time.Unix(0, ms*int64(time.Millisecond))
	return tm.Format("2006-01-02 15:04:05")
}

func F64ToStr(f float64) string {
	s := strconv.FormatFloat(f, 'f', 3, 64)
	return s
}

func StrToF64(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func StrToI64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func SliceFind(slice []int64, val int64) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func FRound(x float64) float64 {
	return math.Round(x*1000) / 1000
}

func FRound2(x float64) float64 {
	return math.Round(x*100) / 100
}
