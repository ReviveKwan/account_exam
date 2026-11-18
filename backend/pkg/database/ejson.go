package database

import (
	"strconv"
	"time"
)

// NormalizeEJSON 将 Strict EJSON 转为普通 JSON 可解码的值
func NormalizeEJSON(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		// {"$numberInt":"1"} / {"$numberLong":"1"} / {"$numberDouble":"1.2"} / {"$numberDecimal":"1.2"}
		if s, ok := t["$numberInt"].(string); ok && len(t) == 1 {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				return n
			}
		}
		if s, ok := t["$numberLong"].(string); ok && len(t) == 1 {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				return n
			}
		}
		if s, ok := t["$numberDouble"].(string); ok && len(t) == 1 {
			if n, err := strconv.ParseFloat(s, 64); err == nil {
				return n
			}
		}
		if s, ok := t["$numberDecimal"].(string); ok && len(t) == 1 {
			if n, err := strconv.ParseFloat(s, 64); err == nil {
				return n
			}
		}
		// {"$oid":"..."}
		if s, ok := t["$oid"].(string); ok && len(t) == 1 {
			return s
		}
		// {"$date":{"$numberLong":"..."}} 或 {"$date":"..."}
		if d, ok := t["$date"]; ok && len(t) == 1 {
			switch dv := d.(type) {
			case string:
				return dv
			case map[string]interface{}:
				if s, ok := dv["$numberLong"].(string); ok {
					if ms, err := strconv.ParseInt(s, 10, 64); err == nil {
						return time.UnixMilli(ms).UTC().Format(time.RFC3339Nano)
					}
				}
			case float64:
				return time.UnixMilli(int64(dv)).UTC().Format(time.RFC3339Nano)
			}
		}
		out := make(map[string]interface{}, len(t))
		for k, vv := range t {
			out[k] = NormalizeEJSON(vv)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, vv := range t {
			out[i] = NormalizeEJSON(vv)
		}
		return out
	default:
		return v
	}
}
