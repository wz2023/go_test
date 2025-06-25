package util

import (
	"encoding/json"
	"runtime/debug"
	"strconv"
	"time"
)

var defaultTime = time.Date(2006, 1, 2, 15, 4, 5, 0, time.Local)

func ToInt(value interface{}) int {
	if value == nil {
		return 0
	}
	switch value.(type) { //多选语句switch
	case int32:
		return value.(int)
	case int64:
		return int(value.(int64))
	case int:
		return int(value.(int))
	case float64:
		return int(value.(float64))
	case []uint8:
		v1, err := strconv.ParseInt(string(value.([]uint8)), 10, 64)
		if err != nil {
			return 0
		}
		return int(v1)
	case string:
		if value.(string) == "" {
			return 0
		}
		ret, err := strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			debug.PrintStack()
		}
		return int(ret)
	}
	return 0
}

func ToInt32(value interface{}) int32 {
	if value == nil {
		return 0
	}
	switch value.(type) { //多选语句switch
	case int32:
		return value.(int32)
	case int64:
		return int32(value.(int64))
	case int:
		return int32(value.(int))
	case float64:
		return int32(value.(float64))
	case []uint8:
		v1, err := strconv.ParseInt(string(value.([]uint8)), 10, 64)
		if err != nil {
			return 0
		}
		return int32(v1)
	case string:
		if value.(string) == "" {
			return 0
		}
		ret, err := strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			debug.PrintStack()
		}
		return int32(ret)
	}
	return 0
}
func ToInt64(value interface{}) int64 {
	if value == nil {
		return 0
	}
	switch value.(type) { //多选语句switch
	case int32:
		return int64(value.(int32))
	case uint32:
		return int64(value.(uint32))
	case int64:
		return value.(int64)
	case int:
		return int64(value.(int))
	case float64:
		return int64(value.(float64))
	case []uint8:
		v1, err := strconv.ParseInt(string(value.([]uint8)), 10, 64)
		if err != nil {
			return 0
		}
		return int64(v1)
	case string:
		if value.(string) == "" {
			return 0
		}
		ret, err := strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			debug.PrintStack()
		}
		return ret
	}
	debug.PrintStack()
	return 0
}
func ToUint32(value interface{}) uint32 {
	return uint32(ToInt32(value))
}
func ToString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch value.(type) { //多选语句switch
	case int32:
		return strconv.Itoa(int(value.(int32)))
	case uint32:
		return strconv.Itoa(int(value.(uint32)))
	case int:
		return strconv.Itoa(int(value.(int)))
	case uint16:
		return strconv.Itoa(int(value.(uint16)))
	case float32:
		return strconv.FormatFloat(float64(value.(float32)), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(value.(float64), 'f', -1, 64)
	case []uint8:
		return string(value.([]uint8))
	case int64:
		return strconv.FormatInt(value.(int64), 10)
	case uint64:
		return strconv.FormatUint(value.(uint64), 10)
	case time.Time:
		return value.(time.Time).Format("2006-01-02 15:04:05")
	case string:
		return value.(string)
	default:
		debug.PrintStack()
		ret, err := json.Marshal(value)
		if err != nil {
		}
		return string(ret)
	}
	return ""
}
func ToFloat64(value interface{}) float64 {
	switch value.(type) { //多选语句switch
	case int32:
		return float64(value.(int32))
	case int:
		return float64(value.(int))
	case string:
		ret, err := strconv.ParseFloat(value.(string), 64)
		if err != nil {
			debug.PrintStack()
		}
		return ret
	}
	debug.PrintStack()
	return 0
}

func ToTime(value interface{}) time.Time {
	if value == nil {
		return defaultTime
	}
	switch value.(type) { //多选语句switch
	case time.Time:
		value := value.(time.Time)
		return value
	case string:
		str := value.(string)
		ret, err := time.ParseInLocation("2006-01-02 15:04:05", str, time.Local)
		if err != nil {
			return defaultTime
		}
		return ret
	}
	return defaultTime
}
func ToBytes(value interface{}) []byte {
	if value == nil {
		return []byte{}
	}
	switch value.(type) { //多选语句switch
	case []uint8:
		v := []byte(value.([]uint8))
		return v
	case uint32, int:
		v := make([]uint8, 4)
		_v := value.(uint32)
		v[0] = uint8(_v >> 24)
		v[1] = uint8(_v >> 16)
		v[2] = uint8(_v >> 8)
		v[3] = uint8(_v)
		return v
	}
	return []byte{}
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
