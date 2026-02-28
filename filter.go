package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Op int

const (
	OpEq Op = iota
	OpNeq
	OpGt
	OpGte
	OpLt
	OpLte
	OpContains
)

type Filter struct {
	Key   string
	Op    Op
	Value string
}

// parseFilters parses an input string into a list of filters.
// Supports operators: = != > >= < <= ~
// Filters can be separated by & or spaces.
// Returns the parsed filters and whether the input is valid.
func parseFilters(input string) ([]Filter, bool) {
	// Split on & first
	segments := strings.Split(input, "&")
	var filters []Filter
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		parsed, ok := parseSegment(seg)
		if !ok {
			return nil, false
		}
		filters = append(filters, parsed...)
	}
	return filters, true
}

// parseSegment parses one or more space-separated filters from a segment (no & present).
func parseSegment(s string) ([]Filter, bool) {
	var filters []Filter
	for len(s) > 0 {
		s = strings.TrimLeft(s, " ")
		if len(s) == 0 {
			break
		}
		// Find the operator by scanning for the first operator character
		opIdx, op, opLen := findOperator(s)
		if opIdx < 1 {
			return nil, false
		}
		key := s[:opIdx]
		rest := s[opIdx+opLen:]
		if len(rest) == 0 {
			return nil, false
		}
		var value string
		if rest[0] == '"' {
			end := strings.Index(rest[1:], "\"")
			if end == -1 {
				return nil, false
			}
			value = rest[1 : end+1]
			s = rest[end+2:]
		} else {
			end := strings.Index(rest, " ")
			if end == -1 {
				value = rest
				s = ""
			} else {
				value = rest[:end]
				s = rest[end:]
			}
		}
		filters = append(filters, Filter{Key: key, Op: op, Value: value})
	}
	return filters, true
}

// findOperator finds the first operator in the string.
// Returns the index where the operator starts, the Op type, and the operator length.
func findOperator(s string) (int, Op, int) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '!':
			if i+1 < len(s) && s[i+1] == '=' {
				return i, OpNeq, 2
			}
		case '>':
			if i+1 < len(s) && s[i+1] == '=' {
				return i, OpGte, 2
			}
			return i, OpGt, 1
		case '<':
			if i+1 < len(s) && s[i+1] == '=' {
				return i, OpLte, 2
			}
			return i, OpLt, 1
		case '~':
			return i, OpContains, 1
		case '=':
			return i, OpEq, 1
		}
	}
	return -1, OpEq, 0
}

// getNestedField traverses nested maps using dot-separated keys.
func getNestedField(obj map[string]any, key string) (any, bool) {
	parts := strings.Split(key, ".")
	var current any = obj
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = m[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// matchesFilter checks if a JSON log line matches all given filters.
func matchesFilter(line string, filters []Filter) bool {
	if len(filters) == 0 {
		return true
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		return false
	}
	for _, f := range filters {
		val, ok := getNestedField(obj, f.Key)
		if !ok {
			return false
		}
		strVal := fmt.Sprintf("%v", val)
		switch f.Op {
		case OpEq:
			if strVal != f.Value {
				return false
			}
		case OpNeq:
			if strVal == f.Value {
				return false
			}
		case OpContains:
			if !strings.Contains(strVal, f.Value) {
				return false
			}
		case OpGt, OpGte, OpLt, OpLte:
			a, errA := strconv.ParseFloat(strVal, 64)
			b, errB := strconv.ParseFloat(f.Value, 64)
			if errA != nil || errB != nil {
				return false
			}
			switch f.Op {
			case OpGt:
				if !(a > b) {
					return false
				}
			case OpGte:
				if !(a >= b) {
					return false
				}
			case OpLt:
				if !(a < b) {
					return false
				}
			case OpLte:
				if !(a <= b) {
					return false
				}
			}
		}
	}
	return true
}

// filtersEqual returns true if two filter slices are equivalent.
func filtersEqual(a, b []Filter) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
