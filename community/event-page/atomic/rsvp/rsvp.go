// @atomic route=post:rsvp auth=none env=.env

package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"drift-sdk"
)

type RequestBody struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Guests int    `json:"guests"` // total attendees including the registrant
}

func PostRsvp(req RequestBody) (int, string, interface{}) {
	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(strings.ToLower(req.Email))

	if name == "" || email == "" || !strings.Contains(email, "@") {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "name and a valid email are required",
		}
	}
	if req.Guests < 1 {
		req.Guests = 1
	}
	if req.Guests > 10 {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "maximum 10 guests per RSVP",
		}
	}

	doc := map[string]any{
		"name":    name,
		"email":   email,
		"guests":  req.Guests,
		"rsvp_at": time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := drift.BackboneWrite("rsvps", doc); err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save RSVP",
		}
	}

	return http.StatusOK, "RSVP received", map[string]string{
		"message": fmt.Sprintf("See you there, %s! We've saved your spot%s.", name, guestSuffix(req.Guests)),
	}
}

func guestSuffix(guests int) string {
	if guests <= 1 {
		return ""
	}
	return fmt.Sprintf(" for %d", guests)
}
