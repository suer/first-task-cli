package firestore

import (
	"strconv"
	"time"
)

type Fields map[string]map[string]any

func String(s string) map[string]any { return map[string]any{"stringValue": s} }

func Int(n int) map[string]any { return map[string]any{"integerValue": strconv.Itoa(n)} }

func Null() map[string]any { return map[string]any{"nullValue": nil} }

func Time(t time.Time) map[string]any {
	return map[string]any{"timestampValue": t.UTC().Format(time.RFC3339Nano)}
}

func OptTime(t *time.Time) map[string]any {
	if t == nil {
		return Null()
	}
	return Time(*t)
}

func Strings(ss []string) map[string]any {
	values := make([]map[string]any, 0, len(ss))
	for _, s := range ss {
		values = append(values, String(s))
	}
	return map[string]any{"arrayValue": map[string]any{"values": values}}
}

func (f Fields) String(key string) string {
	s, _ := f[key]["stringValue"].(string)
	return s
}

func (f Fields) Int(key string) int {
	s, _ := f[key]["integerValue"].(string)
	n, _ := strconv.Atoi(s)
	return n
}

func (f Fields) Time(key string) *time.Time {
	s, ok := f[key]["timestampValue"].(string)
	if !ok {
		return nil
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return nil
	}
	return &t
}

func (f Fields) Strings(key string) []string {
	arr, _ := f[key]["arrayValue"].(map[string]any)
	values, _ := arr["values"].([]any)
	out := make([]string, 0, len(values))
	for _, v := range values {
		m, _ := v.(map[string]any)
		if s, ok := m["stringValue"].(string); ok {
			out = append(out, s)
		}
	}
	return out
}
