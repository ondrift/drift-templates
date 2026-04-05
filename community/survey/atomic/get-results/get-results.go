// @atomic route=get:get-results auth=apikey env=.env
// Returns a lightweight summary of survey results.
// Protected by API key — only the survey owner should see results.
// The runner proxy has already verified the API key before this is called.

package main

import (
	"net/http"
	"strconv"

	"drift-sdk"
)

func GetGetResults() (int, string, interface{}) {
	countRaw, err := drift.CacheGet("survey:count")
	count := 0
	if err == nil && len(countRaw) > 0 {
		count, _ = strconv.Atoi(string(countRaw))
	}

	return http.StatusOK, "OK", map[string]any{
		"total_responses": count,
		"note":            "Detailed responses are stored in the 'responses' NoSQL collection. Use the Drift CLI to query them.",
	}
}
