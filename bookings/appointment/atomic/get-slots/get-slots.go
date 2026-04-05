// @atomic route=get:get-slots auth=none env=.env
// Returns available time slots for a given date and optional service.
// Query params: date=YYYY-MM-DD  (required)
//               service=<name>   (optional filter, not yet used server-side)
//
// All configured slots are checked against the cache lock keys written by
// book-slot. Slots without a lock key are returned as available.

package main

import (
	"fmt"
	"net/http"
	"os"

	"drift-sdk"
)

// slot definitions — owner can update these to match their actual schedule
var allSlots = []string{
	"09:00", "09:30",
	"10:00", "10:30",
	"11:00", "11:30",
	"12:00", "12:30",
	"14:00", "14:30",
	"15:00", "15:30",
	"16:00", "16:30",
}

func GetGetSlots() (int, string, interface{}) {
	date := os.Getenv("DRIFT_QUERY_DATE") // set by runner from ?date= param
	if date == "" {
		// Without a date we can't check availability; return full list
		return http.StatusOK, "OK", map[string]any{
			"date":  "unknown",
			"slots": allSlots,
			"note":  "Pass ?date=YYYY-MM-DD to get real availability",
		}
	}

	available := []string{}
	for _, slot := range allSlots {
		lockKey := fmt.Sprintf("slot:%s:%s", date, slot)
		raw, err := drift.CacheGet(lockKey)
		if err != nil || len(raw) == 0 {
			available = append(available, slot)
		}
	}

	return http.StatusOK, "OK", map[string]any{
		"date":      date,
		"slots":     available,
		"total":     len(allSlots),
		"available": len(available),
	}
}
