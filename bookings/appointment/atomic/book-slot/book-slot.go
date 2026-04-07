// @atomic route=post:book-slot auth=none env=.env
// Creates a booking for a named service at a given date + time slot.
// Uses a cache key as a lightweight lock to prevent double-booking.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	drift "github.com/ondrift/drift-sdk"
)

type RequestBody struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Date     string `json:"date"`      // YYYY-MM-DD
	TimeSlot string `json:"time_slot"` // e.g. "10:00"
	Service  string `json:"service"`
}

func PostBookSlot(req RequestBody) (int, string, interface{}) {
	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(strings.ToLower(req.Email))
	date := strings.TrimSpace(req.Date)
	timeSlot := strings.TrimSpace(req.TimeSlot)
	service := strings.TrimSpace(req.Service)

	if name == "" || email == "" || date == "" || timeSlot == "" || service == "" || !strings.Contains(email, "@") {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "name, email, date, time_slot, and service are all required",
		}
	}

	// ── Step 1: check whether the slot is already taken ───────────────────────
	lockKey := fmt.Sprintf("slot:%s:%s", date, timeSlot)
	raw, err := drift.CacheGet(lockKey)
	if err == nil && len(raw) > 0 {
		return http.StatusConflict, "Slot taken", map[string]string{
			"error": fmt.Sprintf("The %s slot on %s is no longer available.", timeSlot, date),
		}
	}

	// ── Step 2: claim the slot in cache (TTL 48 h = 172800 s) ────────────────
	if err := drift.CacheSet(lockKey, "1", 172800); err != nil {
		return http.StatusInternalServerError, "Cache error", map[string]string{
			"error": "Failed to reserve slot. Please try again.",
		}
	}

	// ── Step 3: persist booking to NoSQL ─────────────────────────────────────
	bookingID := fmt.Sprintf("booking-%d", time.Now().UnixNano())
	doc := map[string]any{
		"id":        bookingID,
		"name":      name,
		"email":     email,
		"date":      date,
		"time_slot": timeSlot,
		"service":   service,
		"booked_at": time.Now().UTC().Format(time.RFC3339),
		"status":    "confirmed",
	}
	if _, err := drift.BackboneWrite("bookings", doc); err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "Could not save booking. Please call us directly.",
		}
	}

	// ── Step 4: store booking in cache for cancel-booking lookups (TTL 90 days)
	bookingJSON, _ := json.Marshal(map[string]any{
		"id":        bookingID,
		"name":      name,
		"email":     email,
		"date":      date,
		"time_slot": timeSlot,
		"service":   service,
		"status":    "confirmed",
	})
	_ = drift.CacheSet(fmt.Sprintf("booking:%s", bookingID), string(bookingJSON), 7776000)

	// ── Step 5: enqueue for confirmation email ────────────────────────────────
	_ = drift.QueuePush("booking-queue", map[string]any{
		"booking_id": bookingID,
		"name":       name,
		"email":      email,
		"date":       date,
		"time_slot":  timeSlot,
		"service":    service,
	})

	return http.StatusOK, "Booking confirmed", map[string]string{
		"message":    fmt.Sprintf("Your %s is booked for %s at %s, %s. A confirmation email is on its way!", service, date, timeSlot, name),
		"booking_id": bookingID,
	}
}
