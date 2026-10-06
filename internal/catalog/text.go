// Package catalog gives read-only access to the NISCAT data (data/data.db) and implements its business rules.
package catalog

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// NormalizeRef reduces a part reference to its search key: uppercase ASCII letters and digits only.
func NormalizeRef(ref string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(ref) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// NormalizeVIN removes whitespace and uppercases, exactly like the extraction did for vin.vin.
func NormalizeVIN(vin string) string {
	return strings.ToUpper(strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, vin))
}

// ItemKey returns a drawing item number without leading zeros, so "01" (parts list) matches "1" (hotspot).
func ItemKey(item string) string {
	item = strings.TrimSpace(item)
	if item == "" {
		return ""
	}
	if key := strings.TrimLeft(item, "0"); key != "" {
		return key
	}
	return "0"
}

// Period is an application period in YYYYMM; a zero bound is open.
type Period struct {
	From int
	To   int
}

// ParsePeriod parses a DATAPLIC value: "MMYY-MMYY", "-MMYY" or "MMYY-".
func ParsePeriod(s string) (Period, bool) {
	from, to, ok := strings.Cut(strings.TrimSpace(s), "-")
	if !ok || (from == "" && to == "") {
		return Period{}, false
	}
	var p Period
	if from != "" {
		if p.From, ok = parseMMYY(from); !ok {
			return Period{}, false
		}
	}
	if to != "" {
		if p.To, ok = parseMMYY(to); !ok {
			return Period{}, false
		}
	}
	return p, true
}

func parseMMYY(s string) (int, bool) {
	if len(s) != 4 {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	month, yy := n/100, n%100
	if month < 1 || month > 12 {
		return 0, false
	}
	return pivotYear(yy)*100 + month, true
}

// pivotYear maps a two-digit year to 1950-2049 (the catalogue covers 1986-2015).
func pivotYear(yy int) int {
	if yy >= 50 {
		return 1900 + yy
	}
	return 2000 + yy
}

// Contains reports whether yyyymm falls inside the period, bounds inclusive.
func (p Period) Contains(yyyymm int) bool {
	return (p.From == 0 || yyyymm >= p.From) && (p.To == 0 || yyyymm <= p.To)
}

// ParseYYYYMM parses a production date such as "198905".
func ParseYYYYMM(s string) (int, bool) {
	if len(s) != 6 {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil || n%100 < 1 || n%100 > 12 {
		return 0, false
	}
	return n, true
}

// yearMonth formats 198905 as "1989-05" for the API; nil for an unknown or open bound.
func yearMonth(v int) *string {
	if v == 0 {
		return nil
	}
	s := fmt.Sprintf("%04d-%02d", v/100, v%100)
	return &s
}

// parseMonthYear parses catalog and section dates "MM/YY" or "MM-YY".
func parseMonthYear(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if len(s) != 5 || (s[2] != '/' && s[2] != '-') {
		return 0, false
	}
	return parseMMYY(s[:2] + s[3:])
}

// yearMonthOf converts a "MM/YY" or "MM-YY" date to the API format; nil when absent or malformed.
func yearMonthOf(s string) *string {
	v, ok := parseMonthYear(s)
	if !ok {
		return nil
	}
	return yearMonth(v)
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
