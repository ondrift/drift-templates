// @atomic route=get:check-position auth=none env=.env

package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	drift "github.com/ondrift/drift-sdk/go"
)

func GetCheckPosition() (int, string, interface{}) {
	email := strings.TrimSpace(strings.ToLower(drift.Env("DRIFT_QUERY_EMAIL")))
	if email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "email query parameter is required",
		}
	}

	raw, err := drift.Cache.Get("waitlist:" + email)
	if err != nil || len(raw) == 0 {
		return http.StatusNotFound, "Not found", map[string]string{
			"error": "this email is not on the waitlist",
		}
	}

	var entry map[string]any
	if err := json.Unmarshal(raw, &entry); err != nil {
		return http.StatusInternalServerError, "Error", map[string]string{
			"error": "failed to read waitlist entry",
		}
	}

	// Read total count from counter.
	total := 0
	counterRaw, _ := drift.Cache.Get("waitlist:counter")
	if len(counterRaw) > 0 {
		total, _ = strconv.Atoi(string(counterRaw))
	}

	return http.StatusOK, "OK", map[string]any{
		"position":      entry["position"],
		"total":         total,
		"referral_code": entry["referral_code"],
	}
}
