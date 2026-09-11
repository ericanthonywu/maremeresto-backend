package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// OutletLocation is the wall-clock timezone every outlet trades in. Operating
// hours are stored as local "HH:MM" strings, so they must be compared against
// Jakarta time and not the server's clock.
var OutletLocation = mustLoadJakarta()

func mustLoadJakarta() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fixed offset fallback for images without tzdata.
		return time.FixedZone("WIB", 7*60*60)
	}
	return loc
}

type DayHours struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

// ParseOperatingHours reads the branch_settings.operating_hours JSON blob,
// which is shaped {"weekday": {"open","close"}, "weekend": {"open","close"}}
// and may also carry per-day overrides keyed by lowercase weekday name.
func ParseOperatingHours(raw map[string]any, day time.Weekday) (DayHours, bool) {
	if raw == nil {
		return DayHours{}, false
	}

	read := func(key string) (DayHours, bool) {
		entry, ok := raw[key].(map[string]any)
		if !ok {
			return DayHours{}, false
		}
		open, _ := entry["open"].(string)
		closed, _ := entry["close"].(string)
		if open == "" || closed == "" {
			return DayHours{}, false
		}
		return DayHours{Open: open, Close: closed}, true
	}

	// Most specific first: an explicit weekday name beats the weekday/weekend
	// buckets.
	if h, ok := read(strings.ToLower(day.String())); ok {
		return h, true
	}
	if day == time.Saturday || day == time.Sunday {
		if h, ok := read("weekend"); ok {
			return h, true
		}
	}
	if h, ok := read("weekday"); ok {
		return h, true
	}
	return DayHours{}, false
}

func parseClock(v string) (int, bool) {
	parts := strings.SplitN(strings.TrimSpace(v), ":", 2)
	if len(parts) != 2 {
		return 0, false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}

// IsWithinOperatingHours reports whether the outlet is trading at the given
// instant. An unparseable or absent schedule means "no restriction" so a bad
// settings row can never lock an outlet out of taking orders.
func IsWithinOperatingHours(raw map[string]any, at time.Time) bool {
	local := at.In(OutletLocation)

	hours, ok := ParseOperatingHours(raw, local.Weekday())
	if !ok {
		return true
	}

	openMin, ok1 := parseClock(hours.Open)
	closeMin, ok2 := parseClock(hours.Close)
	if !ok1 || !ok2 {
		return true
	}

	nowMin := local.Hour()*60 + local.Minute()

	if closeMin > openMin {
		return nowMin >= openMin && nowMin < closeMin
	}
	// Closing time past midnight, e.g. 17:00 -> 01:00.
	return nowMin >= openMin || nowMin < closeMin
}

// FormatOperatingHours renders today's schedule for display, e.g. "08.00–22.00".
func FormatOperatingHours(raw map[string]any, at time.Time) string {
	hours, ok := ParseOperatingHours(raw, at.In(OutletLocation).Weekday())
	if !ok {
		return ""
	}
	dot := func(v string) string { return strings.ReplaceAll(v, ":", ".") }
	return fmt.Sprintf("%s–%s", dot(hours.Open), dot(hours.Close))
}
