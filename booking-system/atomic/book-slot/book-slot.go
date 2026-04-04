// @atomic route=post:book-slot auth=none env=.env
// Creates a booking for a named service at a given date + time slot.
// Uses a cache key as a lightweight lock to prevent double-booking.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
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

	backbone := os.Getenv("BACKBONE_URL")
	if backbone == "" {
		backbone = "http://backbone:8000"
	}

	// ── Step 1: check whether the slot is already taken ───────────────────────
	lockKey := fmt.Sprintf("slot:%s:%s", date, timeSlot)
	checkResp, err := http.Get(fmt.Sprintf("%s/cache/get?key=%s", backbone, lockKey))
	if err == nil && checkResp.StatusCode == http.StatusOK {
		checkResp.Body.Close()
		return http.StatusConflict, "Slot taken", map[string]string{
			"error": fmt.Sprintf("The %s slot on %s is no longer available.", timeSlot, date),
		}
	}
	if checkResp != nil {
		checkResp.Body.Close()
	}

	// ── Step 2: claim the slot in cache (TTL 48 h = 172800 s) ────────────────
	lockPayload, _ := json.Marshal(map[string]any{
		"key":   lockKey,
		"value": "1",
		"ttl":   172800, // 48 hours in seconds
	})
	setResp, err := http.Post(fmt.Sprintf("%s/cache/set", backbone), "application/json", bytes.NewReader(lockPayload))
	if err != nil || setResp.StatusCode != http.StatusOK {
		if setResp != nil {
			setResp.Body.Close()
		}
		return http.StatusInternalServerError, "Cache error", map[string]string{
			"error": "Failed to reserve slot. Please try again.",
		}
	}
	setResp.Body.Close()

	// ── Step 3: persist booking to NoSQL ─────────────────────────────────────
	bookingID := fmt.Sprintf("booking-%d", time.Now().UnixNano())
	bookingData := map[string]any{
		"collection": "bookings",
		"id":         bookingID,
		"name":       name,
		"email":      email,
		"date":       date,
		"time_slot":  timeSlot,
		"service":    service,
		"booked_at":  time.Now().UTC().Format(time.RFC3339),
		"status":     "confirmed",
	}
	docBytes, _ := json.Marshal(bookingData)
	writeResp, err := http.Post(fmt.Sprintf("%s/write", backbone), "application/json", bytes.NewReader(docBytes))
	if err != nil || writeResp.StatusCode != http.StatusOK {
		if writeResp != nil {
			writeResp.Body.Close()
		}
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "Could not save booking. Please call us directly.",
		}
	}
	writeResp.Body.Close()

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
	cachePayload, _ := json.Marshal(map[string]any{
		"key":   fmt.Sprintf("booking:%s", bookingID),
		"value": string(bookingJSON),
		"ttl":   7776000, // 90 days in seconds
	})
	cacheResp, _ := http.Post(fmt.Sprintf("%s/cache/set", backbone), "application/json", bytes.NewReader(cachePayload))
	if cacheResp != nil {
		cacheResp.Body.Close()
	}

	// ── Step 5: enqueue for confirmation email ────────────────────────────────
	queuePayload, _ := json.Marshal(map[string]any{
		"queue": "booking-queue",
		"body": map[string]any{
			"booking_id": bookingID,
			"name":       name,
			"email":      email,
			"date":       date,
			"time_slot":  timeSlot,
			"service":    service,
		},
	})
	pushResp, _ := http.Post(fmt.Sprintf("%s/queue/push", backbone), "application/json", bytes.NewReader(queuePayload))
	if pushResp != nil {
		pushResp.Body.Close()
	}

	return http.StatusOK, "Booking confirmed", map[string]string{
		"message":    fmt.Sprintf("Your %s is booked for %s at %s, %s. A confirmation email is on its way!", service, date, timeSlot, name),
		"booking_id": bookingID,
	}
}
