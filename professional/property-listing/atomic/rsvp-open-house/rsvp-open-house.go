// @atomic route=post:rsvp-open-house auth=none env=.env

package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	drift "github.com/ondrift/drift-sdk"
)

type RequestBody struct {
	ListingID string `json:"listing_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	PartySize int    `json:"party_size"`
}

func PostRsvpOpenHouse(req RequestBody) (int, string, interface{}) {
	if req.Name == "" || req.Email == "" || req.ListingID == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "name, email, and listing_id are required",
		}
	}

	if req.PartySize < 1 {
		req.PartySize = 1
	}
	if req.PartySize > 6 {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "party size must be between 1 and 6",
		}
	}

	// Generate RSVP ID from timestamp.
	nano := fmt.Sprintf("%d", time.Now().UnixNano())
	rsvpID := "RSVP-" + strings.ToUpper(nano[len(nano)-8:])

	doc := map[string]any{
		"id":         rsvpID,
		"listing_id": req.ListingID,
		"name":       req.Name,
		"email":      req.Email,
		"phone":      req.Phone,
		"party_size": req.PartySize,
		"rsvp_at":    time.Now().UTC().Format(time.RFC3339),
	}
	_, err := drift.BackboneWrite("rsvps", doc)
	if err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save RSVP",
		}
	}

	// Enqueue agent notification.
	_ = drift.QueuePush("agent-queue", map[string]any{
		"type":       "rsvp",
		"rsvp_id":    rsvpID,
		"listing_id": req.ListingID,
		"name":       req.Name,
		"email":      req.Email,
		"phone":      req.Phone,
		"party_size": req.PartySize,
	})

	return http.StatusOK, "RSVP confirmed", map[string]any{
		"rsvp_id": rsvpID,
		"message": fmt.Sprintf("Thanks %s! You're registered for the open house (party of %d). Reference: %s. We'll send a reminder before the event.", req.Name, req.PartySize, rsvpID),
	}
}
