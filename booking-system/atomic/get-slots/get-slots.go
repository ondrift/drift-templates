// @atomic route=get:get-slots auth=none env=.env
// Returns available time slots for a given date and optional service.
// Query params: date=YYYY-MM-DD  (required)
//               service=<name>   (optional filter, not yet used server-side)
//
// All configured slots are checked against the cache lock keys written by
// book-slot. Slots without a lock key are returned as available.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	// The runner injects the raw *http.Request; we reach for env vars instead
	// since the function signature is fixed. The date must be passed as a
	// query param; the runner forwards query strings to the function URL.
	// We read it from the DRIFT_QUERY_DATE env that the runner sets, or fall
	// back to returning all slots if it isn't available.
	date := os.Getenv("DRIFT_QUERY_DATE") // set by runner from ?date= param
	if date == "" {
		// Without a date we can't check availability; return full list
		return http.StatusOK, "OK", map[string]any{
			"date":  "unknown",
			"slots": allSlots,
			"note":  "Pass ?date=YYYY-MM-DD to get real availability",
		}
	}

	backbone := os.Getenv("BACKBONE_URL")
	if backbone == "" {
		backbone = "http://backbone:8000"
	}

	available := []string{}
	for _, slot := range allSlots {
		lockKey := fmt.Sprintf("slot:%s:%s", date, slot)
		resp, err := http.Get(fmt.Sprintf("%s/cache/get?key=%s", backbone, lockKey))
		if err != nil || resp.StatusCode != http.StatusOK {
			available = append(available, slot)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	return http.StatusOK, "OK", map[string]any{
		"date":      date,
		"slots":     available,
		"total":     len(allSlots),
		"available": len(available),
	}
}

// keep compiler happy when get-slots is built as part of the template
var _ = json.Marshal
