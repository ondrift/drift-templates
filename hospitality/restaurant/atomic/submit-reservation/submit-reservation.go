// @atomic route=post:submit-reservation auth=none env=.env

package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	drift "github.com/ondrift/drift-sdk"
)

type RequestBody struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Date      string `json:"date"`       // "2026-08-14"
	Time      string `json:"time"`       // "19:30"
	PartySize int    `json:"party_size"` // 1–12
	Notes     string `json:"notes"`
}

func PostSubmitReservation(req RequestBody) (int, string, interface{}) {
	if req.Name == "" || req.Email == "" || req.Date == "" || req.Time == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "name, email, date, and time are required",
		}
	}
	if req.PartySize < 1 || req.PartySize > 12 {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "party size must be between 1 and 12",
		}
	}

	// Generate a short confirmation code (last 6 chars of unix nano timestamp).
	nano := fmt.Sprintf("%d", time.Now().UnixNano())
	confirmCode := strings.ToUpper(nano[len(nano)-6:])

	doc := map[string]any{
		"name":         req.Name,
		"email":        req.Email,
		"date":         req.Date,
		"time":         req.Time,
		"party_size":   req.PartySize,
		"notes":        req.Notes,
		"confirm_code": confirmCode,
		"status":       "pending",
		"created_at":   time.Now().UTC().Format(time.RFC3339),
	}
	_, err := drift.BackboneWrite("reservations", doc)
	if err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save reservation",
		}
	}

	// Enqueue confirmation email.
	_ = drift.QueuePush("reservation-queue", map[string]any{
		"name":         req.Name,
		"email":        req.Email,
		"date":         req.Date,
		"time":         req.Time,
		"party_size":   req.PartySize,
		"confirm_code": confirmCode,
	})

	return http.StatusOK, "Reservation received", map[string]any{
		"confirm_code": confirmCode,
		"message":      fmt.Sprintf("Thanks %s! Your reservation for %d on %s at %s is confirmed. Code: %s", req.Name, req.PartySize, req.Date, req.Time, confirmCode),
	}
}
