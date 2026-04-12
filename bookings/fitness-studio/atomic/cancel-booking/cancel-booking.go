// @atomic route=post:cancel-booking auth=none env=.env

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	drift "github.com/ondrift/drift-sdk/go"
)

type RequestBody struct {
	BookingID string `json:"booking_id"`
	Email     string `json:"email"`
}

type booking struct {
	ID      string `json:"id"`
	ClassID string `json:"class_id"`
	Date    string `json:"date"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Status  string `json:"status"`
}

func PostCancelBooking(req RequestBody) (int, string, interface{}) {
	if req.BookingID == "" || req.Email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "booking_id and email are required",
		}
	}

	// Look up booking from cache.
	raw, err := drift.Cache.Get("booking:" + req.BookingID)
	if err != nil || len(raw) == 0 {
		return http.StatusNotFound, "Not Found", map[string]string{
			"error": "booking not found",
		}
	}

	var bk booking
	if err := json.Unmarshal(raw, &bk); err != nil {
		return http.StatusInternalServerError, "Parse error", map[string]string{
			"error": "could not parse booking",
		}
	}

	// Verify email matches.
	if bk.Email != req.Email {
		return http.StatusForbidden, "Forbidden", map[string]string{
			"error": "email does not match booking",
		}
	}

	if bk.Status == "cancelled" {
		return http.StatusConflict, "Conflict", map[string]string{
			"error": "booking is already cancelled",
		}
	}

	// Decrement class spot count.
	countKey := "class:" + bk.ClassID + ":" + bk.Date + ":count"
	currentCount := 0
	countRaw, err := drift.Cache.Get(countKey)
	if err == nil && len(countRaw) > 0 {
		currentCount, _ = strconv.Atoi(string(countRaw))
	}
	newCount := currentCount - 1
	if newCount < 0 {
		newCount = 0
	}
	_ = drift.Cache.Set(countKey, []byte(strconv.Itoa(newCount)), 172800)

	// Write cancellation to NoSQL.
	cancelDoc := map[string]any{
		"id":           bk.ID,
		"class_id":     bk.ClassID,
		"date":         bk.Date,
		"name":         bk.Name,
		"email":        bk.Email,
		"status":       "cancelled",
		"cancelled_at": time.Now().UTC().Format(time.RFC3339),
	}
	_, _ = drift.NoSQL.Collection("bookings").Insert(cancelDoc)

	// Delete booking from cache.
	_ = drift.Cache.Del("booking:" + req.BookingID)

	drift.Log(fmt.Sprintf("[cancel-booking] cancelled %s for %s", req.BookingID, bk.Email))

	return http.StatusOK, "Booking cancelled", map[string]any{
		"booking_id": req.BookingID,
		"message":    fmt.Sprintf("Booking %s has been cancelled. Your spot has been released.", req.BookingID),
	}
}
