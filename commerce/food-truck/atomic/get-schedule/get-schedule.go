// @atomic route=get:get-schedule auth=none env=.env
// Returns the food truck's weekly schedule.
// The schedule is stored in Backbone Cache under key "truck-schedule" as a JSON string.
// Seed it with:
//   drift backbone cache set truck-schedule '{"days":[...]}'
// Update it the same way — the new value takes effect immediately.

package main

import (
	"encoding/json"
	"net/http"

	drift "github.com/ondrift/drift-sdk"
)

func GetGetSchedule() (int, string, interface{}) {
	raw, err := drift.CacheGet("truck-schedule")
	if err != nil || len(raw) == 0 {
		return http.StatusOK, "OK", defaultSchedule()
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return http.StatusOK, "OK", defaultSchedule()
	}

	days, ok := doc["days"]
	if !ok {
		return http.StatusOK, "OK", defaultSchedule()
	}
	return http.StatusOK, "OK", map[string]any{"days": days}
}

func defaultSchedule() map[string]any {
	return map[string]any{
		"days": []map[string]any{
			{"day": "Monday", "location": "Central Park, Main Entrance", "hours": "11:30 – 14:00"},
			{"day": "Tuesday", "location": "Tech Campus, Building C", "hours": "11:00 – 14:00"},
			{"day": "Wednesday", "location": "Riverside Market", "hours": "10:00 – 15:00"},
			{"day": "Thursday", "location": "University Square", "hours": "11:30 – 14:30"},
			{"day": "Friday", "location": "Brewery District", "hours": "11:00 – 20:00"},
			{"day": "Saturday", "location": "Farmers Market", "hours": "09:00 – 14:00"},
			{"day": "Sunday", "location": "Closed", "hours": "—"},
		},
	}
}
