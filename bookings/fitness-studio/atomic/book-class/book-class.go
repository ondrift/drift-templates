// @atomic route=post:book-class auth=none env=.env

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	drift "github.com/ondrift/drift-sdk"
)

type RequestBody struct {
	ClassID string `json:"class_id"`
	Date    string `json:"date"`    // "2026-04-07"
	Name    string `json:"name"`
	Email   string `json:"email"`
}

type classInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Instructor  string   `json:"instructor"`
	Time        string   `json:"time"`
	Duration    string   `json:"duration"`
	MaxCapacity int      `json:"max_capacity"`
	Location    string   `json:"location"`
	Days        []string `json:"days"`
}

type schedule struct {
	Classes []classInfo `json:"classes"`
}

func PostBookClass(req RequestBody) (int, string, interface{}) {
	if req.ClassID == "" || req.Date == "" || req.Name == "" || req.Email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "class_id, date, name, and email are required",
		}
	}

	// Load schedule to find the class and its max_capacity.
	raw, err := drift.Cache.Get("class-schedule")
	if err != nil || len(raw) == 0 {
		return http.StatusInternalServerError, "Schedule error", map[string]string{
			"error": "could not load class schedule",
		}
	}

	var sched schedule
	if err := json.Unmarshal(raw, &sched); err != nil {
		return http.StatusInternalServerError, "Schedule error", map[string]string{
			"error": "could not parse class schedule",
		}
	}

	var found *classInfo
	for i := range sched.Classes {
		if sched.Classes[i].ID == req.ClassID {
			found = &sched.Classes[i]
			break
		}
	}
	if found == nil {
		return http.StatusNotFound, "Not Found", map[string]string{
			"error": "class not found",
		}
	}

	// Check current count for this class + date.
	countKey := "class:" + req.ClassID + ":" + req.Date + ":count"
	currentCount := 0
	countRaw, err := drift.Cache.Get(countKey)
	if err == nil && len(countRaw) > 0 {
		currentCount, _ = strconv.Atoi(string(countRaw))
	}

	if currentCount >= found.MaxCapacity {
		return http.StatusConflict, "Conflict", map[string]string{
			"error": "Class is full",
		}
	}

	// Increment spot count (48h TTL = 172800 seconds).
	newCount := currentCount + 1
	_ = drift.Cache.Set(countKey, []byte(strconv.Itoa(newCount)), 172800)

	// Generate booking ID.
	nano := fmt.Sprintf("%d", time.Now().UnixNano())
	bookingID := "BK-" + strings.ToUpper(nano[len(nano)-8:])

	doc := map[string]any{
		"id":        bookingID,
		"class_id":  req.ClassID,
		"date":      req.Date,
		"name":      req.Name,
		"email":     req.Email,
		"booked_at": time.Now().UTC().Format(time.RFC3339),
		"status":    "confirmed",
	}

	// Store booking in NoSQL.
	_, err = drift.NoSQL.Collection("bookings").Insert(doc)
	if err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save booking",
		}
	}

	// Cache the booking for quick lookup by cancel-booking.
	bookingJSON, _ := json.Marshal(doc)
	_ = drift.Cache.Set("booking:"+bookingID, bookingJSON, 172800)

	// Enqueue confirmation email.
	_ = drift.Queue("class-queue").Push(map[string]any{
		"booking_id": bookingID,
		"class_id":   req.ClassID,
		"class_name": found.Name,
		"date":       req.Date,
		"time":       found.Time,
		"duration":   found.Duration,
		"location":   found.Location,
		"instructor": found.Instructor,
		"name":       req.Name,
		"email":      req.Email,
	})

	return http.StatusOK, "Booking confirmed", map[string]any{
		"booking_id":      bookingID,
		"class_name":      found.Name,
		"date":            req.Date,
		"time":            found.Time,
		"spots_remaining": found.MaxCapacity - newCount,
		"message":         fmt.Sprintf("You're booked for %s on %s at %s. Booking ID: %s", found.Name, req.Date, found.Time, bookingID),
	}
}
