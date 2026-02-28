package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

func transformLine(line string, cfg TransformConfig) string {
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		return line
	}

	// Extract header fields
	var tsStr, level, message string
	var usedKeys []string

	// Timestamp
	if cfg.Timestamp != "" {
		if val, ok := getNestedField(obj, cfg.Timestamp); ok {
			tsStr = formatTimestamp(val)
			usedKeys = append(usedKeys, cfg.Timestamp)
		}
	}

	// Level
	if cfg.Level != "" {
		if val, ok := getNestedField(obj, cfg.Level); ok {
			level = fmt.Sprintf("%v", val)
			usedKeys = append(usedKeys, cfg.Level)
		}
	}

	// Message
	if cfg.Message != "" {
		if val, ok := getNestedField(obj, cfg.Message); ok {
			message = fmt.Sprintf("%v", val)
			usedKeys = append(usedKeys, cfg.Message)
		}
	} else {
		// Try "message" then "msg"
		if val, ok := getNestedField(obj, "message"); ok {
			message = fmt.Sprintf("%v", val)
			usedKeys = append(usedKeys, "message")
		} else if val, ok := getNestedField(obj, "msg"); ok {
			message = fmt.Sprintf("%v", val)
			usedKeys = append(usedKeys, "msg")
		}
	}

	// Build header
	var header strings.Builder
	if tsStr != "" {
		header.WriteString("[")
		header.WriteString(tsStr)
		header.WriteString("] ")
	}
	if level != "" {
		header.WriteString(colorLevel(level))
		header.WriteString(": ")
	}
	header.WriteString(message)

	// Collect remaining fields
	var lines []string
	lines = append(lines, header.String())

	// Get sorted top-level keys, skip used ones
	keys := make([]string, 0, len(obj))
	for k := range obj {
		if !isUsedKey(k, usedKeys) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := obj[k]
		lines = append(lines, "    "+k+": "+formatValue(v))
	}

	return strings.Join(lines, "\r\n")
}

func isUsedKey(key string, used []string) bool {
	for _, u := range used {
		// Match top-level part of dot-path
		top := u
		if idx := strings.IndexByte(u, '.'); idx != -1 {
			top = u[:idx]
		}
		if key == top {
			return true
		}
	}
	return false
}

func formatTimestamp(val any) string {
	switch v := val.(type) {
	case string:
		// Try RFC3339Nano
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			return t.Local().Format("02.01.2006 15:04:05")
		}
		// Try RFC3339
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t.Local().Format("02.01.2006 15:04:05")
		}
		return v
	case float64:
		var t time.Time
		if v > 1e12 {
			// Unix milliseconds
			ms := int64(v)
			t = time.UnixMilli(ms)
		} else {
			// Unix seconds
			sec := int64(v)
			nsec := int64(math.Round((v - float64(sec)) * 1e9))
			t = time.Unix(sec, nsec)
		}
		return t.Local().Format("02.01.2006 15:04:05")
	default:
		return fmt.Sprintf("%v", val)
	}
}

func colorLevel(level string) string {
	upper := strings.ToUpper(level)
	switch upper {
	case "ERROR", "FATAL":
		return "\033[31m" + upper + "\033[0m" // red
	case "WARN", "WARNING":
		return "\033[33m" + upper + "\033[0m" // yellow
	case "INFO":
		return "\033[32m" + upper + "\033[0m" // green
	case "DEBUG", "TRACE":
		return "\033[90m" + upper + "\033[0m" // gray
	default:
		return upper
	}
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case map[string]any, []any:
		b, err := json.MarshalIndent(val, "    ", "  ")
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return strings.ReplaceAll(string(b), "\n", "\r\n")
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}
