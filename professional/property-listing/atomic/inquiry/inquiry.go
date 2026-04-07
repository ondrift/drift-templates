// @atomic route=post:inquiry auth=none env=.env

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
	Message   string `json:"message"`
}

func PostInquiry(req RequestBody) (int, string, interface{}) {
	if req.Name == "" || req.Email == "" || req.Message == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "name, email, and message are required",
		}
	}

	// Generate inquiry ID from timestamp.
	nano := fmt.Sprintf("%d", time.Now().UnixNano())
	inquiryID := "INQ-" + strings.ToUpper(nano[len(nano)-8:])

	doc := map[string]any{
		"id":         inquiryID,
		"listing_id": req.ListingID,
		"name":       req.Name,
		"email":      req.Email,
		"phone":      req.Phone,
		"message":    req.Message,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}
	_, err := drift.BackboneWrite("inquiries", doc)
	if err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save inquiry",
		}
	}

	// Enqueue agent notification.
	_ = drift.QueuePush("agent-queue", map[string]any{
		"type":       "inquiry",
		"inquiry_id": inquiryID,
		"listing_id": req.ListingID,
		"name":       req.Name,
		"email":      req.Email,
		"phone":      req.Phone,
		"message":    req.Message,
	})

	return http.StatusOK, "Inquiry received", map[string]any{
		"inquiry_id": inquiryID,
		"message":    fmt.Sprintf("Thanks %s! Your inquiry has been sent. Reference: %s. We'll be in touch shortly.", req.Name, inquiryID),
	}
}
