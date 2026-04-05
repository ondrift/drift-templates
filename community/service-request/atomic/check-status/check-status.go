// @atomic route=get:check-status auth=none env=.env

package main

import (
	"encoding/json"
	"net/http"

	"drift-sdk"
)

func GetCheckStatus() (int, string, interface{}) {
	ticket := drift.Env("DRIFT_QUERY_TICKET")
	if ticket == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "ticket query parameter is required",
		}
	}

	raw, err := drift.CacheGet("ticket:" + ticket)
	if err != nil || len(raw) == 0 {
		return http.StatusNotFound, "Not Found", map[string]string{
			"error": "Ticket not found",
		}
	}

	var status map[string]string
	if err := json.Unmarshal(raw, &status); err != nil {
		return http.StatusInternalServerError, "Parse error", map[string]string{
			"error": "failed to parse ticket data",
		}
	}

	return http.StatusOK, "OK", map[string]any{
		"ticket":       status["ticket"],
		"category":     status["category"],
		"status":       status["status"],
		"submitted_at": status["submitted_at"],
	}
}
