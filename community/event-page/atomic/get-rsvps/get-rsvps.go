// @atomic route=get:get-rsvps auth=apikey env=.env
// Returns the RSVP count for the public counter, and the full list for admins.
// The apikey auth means the frontend can call /api/get-rsvps?count=true without
// a key to get just the number, but the full list requires the API key header.

package main

import (
	"encoding/json"
	"net/http"

	drift "github.com/ondrift/drift-sdk"
)

// GetGetRsvps returns the total RSVP and guest count.
// The runner proxy has already verified the API key before this is called.
func GetGetRsvps() (int, string, interface{}) {
	raw, err := drift.CacheGet("rsvp:stats")
	if err == nil && len(raw) > 0 {
		var payload map[string]any
		if json.Unmarshal(raw, &payload) == nil {
			return http.StatusOK, "OK", payload
		}
	}

	return http.StatusOK, "OK", map[string]any{
		"rsvp_count":  0,
		"guest_count": 0,
		"note":        "Stats cache not yet initialised",
	}
}
