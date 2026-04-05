// @atomic route=post:cancel-booking auth=none env=.env
// Cancels a booking by ID, releasing the slot so it can be rebooked.
// Body: {"booking_id": "...", "email": "..."}
// Email is required to prevent strangers from cancelling other people's slots.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"drift-sdk"
)

type RequestBody struct {
	BookingID string `json:"booking_id"`
	Email     string `json:"email"`
}

func PostCancelBooking(req RequestBody) (int, string, interface{}) {
	bookingID := strings.TrimSpace(req.BookingID)
	email := strings.TrimSpace(strings.ToLower(req.Email))

	if bookingID == "" || email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "booking_id and email are required",
		}
	}

	// ── Fetch the booking from cache ─────────────────────────────────────────
	cacheKey := fmt.Sprintf("booking:%s", bookingID)
	raw, err := drift.CacheGet(cacheKey)
	if err != nil || len(raw) == 0 {
		return http.StatusNotFound, "Not found", map[string]string{
			"error": "Booking not found. It may have expired or the ID is incorrect.",
		}
	}

	var booking map[string]any
	if err := json.Unmarshal(raw, &booking); err != nil {
		return http.StatusInternalServerError, "Parse error", map[string]string{
			"error": "Could not parse booking record.",
		}
	}

	storedEmail, _ := booking["email"].(string)
	if storedEmail != email {
		return http.StatusForbidden, "Forbidden", map[string]string{
			"error": "Email does not match booking.",
		}
	}

	date, _ := booking["date"].(string)
	timeSlot, _ := booking["time_slot"].(string)

	// ── Write cancelled record to NoSQL ──────────────────────────────────────
	booking["status"] = "cancelled"
	drift.BackboneWrite("bookings", booking)

	// ── Release the slot lock ────────────────────────────────────────────────
	if date != "" && timeSlot != "" {
		_ = drift.CacheDel(fmt.Sprintf("slot:%s:%s", date, timeSlot))
	}

	// ── Remove booking from cache ─────────────────────────────────────────────
	_ = drift.CacheDel(cacheKey)

	return http.StatusOK, "Cancelled", map[string]string{
		"message": "Your booking has been cancelled. The slot is now available for others.",
	}
}
