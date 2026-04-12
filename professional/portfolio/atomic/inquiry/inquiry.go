// @atomic route=post:inquiry auth=none env=.env

package main

import (
	"fmt"
	"net/http"
	"time"

	drift "github.com/ondrift/drift-sdk/go"
)

type RequestBody struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
	Budget  string `json:"budget"`
}

func PostInquiry(req RequestBody) (int, string, interface{}) {
	if req.Name == "" || req.Email == "" || req.Message == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "name, email, and message are required",
		}
	}

	inquiryID := fmt.Sprintf("inq-%d", time.Now().UnixNano())

	doc := map[string]any{
		"id":          inquiryID,
		"name":        req.Name,
		"email":       req.Email,
		"subject":     req.Subject,
		"message":     req.Message,
		"budget":      req.Budget,
		"received_at": time.Now().UTC().Format(time.RFC3339),
		"status":      "new",
	}
	_, err := drift.NoSQL.Collection("inquiries").Insert(doc)
	if err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save inquiry",
		}
	}

	// Enqueue notification email.
	_ = drift.Queue("inquiry-queue").Push(map[string]any{
		"id":      inquiryID,
		"name":    req.Name,
		"email":   req.Email,
		"subject": req.Subject,
		"message": req.Message,
		"budget":  req.Budget,
	})

	return http.StatusOK, "Inquiry received", map[string]any{
		"inquiry_id": inquiryID,
		"message":    fmt.Sprintf("Thanks %s! Your inquiry has been received. I'll get back to you shortly.", req.Name),
	}
}
