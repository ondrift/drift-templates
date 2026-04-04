// @atomic route=post:cancel-booking auth=none env=.env
// Cancels a booking by ID, releasing the slot so it can be rebooked.
// Body: {"booking_id": "...", "email": "..."}
// Email is required to prevent strangers from cancelling other people's slots.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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

	backbone := os.Getenv("BACKBONE_URL")
	if backbone == "" {
		backbone = "http://backbone:8000"
	}

	// ── Fetch the booking from cache ─────────────────────────────────────────
	cacheKey := fmt.Sprintf("booking:%s", bookingID)
	getResp, err := http.Get(fmt.Sprintf("%s/cache/get?key=%s", backbone, cacheKey))
	if err != nil || getResp.StatusCode != http.StatusOK {
		if getResp != nil {
			getResp.Body.Close()
		}
		return http.StatusNotFound, "Not found", map[string]string{
			"error": "Booking not found. It may have expired or the ID is incorrect.",
		}
	}
	defer getResp.Body.Close()

	raw, err := io.ReadAll(getResp.Body)
	if err != nil {
		return http.StatusInternalServerError, "Read error", map[string]string{
			"error": "Could not read booking record.",
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
	booking["collection"] = "bookings"
	docBytes, _ := json.Marshal(booking)
	writeResp, err := http.Post(fmt.Sprintf("%s/write", backbone), "application/json", bytes.NewReader(docBytes))
	if err == nil && writeResp != nil {
		writeResp.Body.Close()
	}

	// ── Release the slot lock ────────────────────────────────────────────────
	if date != "" && timeSlot != "" {
		lockKey := fmt.Sprintf("slot:%s:%s", date, timeSlot)
		delPayload, _ := json.Marshal(map[string]string{"key": lockKey})
		delResp, _ := http.Post(fmt.Sprintf("%s/cache/del", backbone), "application/json", bytes.NewReader(delPayload))
		if delResp != nil {
			delResp.Body.Close()
		}
	}

	// ── Remove booking from cache ─────────────────────────────────────────────
	delPayload, _ := json.Marshal(map[string]string{"key": cacheKey})
	delResp, _ := http.Post(fmt.Sprintf("%s/cache/del", backbone), "application/json", bytes.NewReader(delPayload))
	if delResp != nil {
		delResp.Body.Close()
	}

	return http.StatusOK, "Cancelled", map[string]string{
		"message": "Your booking has been cancelled. The slot is now available for others.",
	}
}
