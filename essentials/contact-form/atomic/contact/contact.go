// @atomic route=post:contact auth=none env=.env
// Stores a contact form submission in the leads NoSQL collection and
// pushes it to the contact-queue so the owner gets an email notification.

package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	drift "github.com/ondrift/drift-sdk/go"
)

type RequestBody struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func PostContact(req RequestBody) (int, string, interface{}) {
	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(strings.ToLower(req.Email))
	subject := strings.TrimSpace(req.Subject)
	message := strings.TrimSpace(req.Message)

	if name == "" || email == "" || message == "" || !strings.Contains(email, "@") {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "name, email, and message are required",
		}
	}
	if subject == "" {
		subject = "New enquiry"
	}

	leadID := fmt.Sprintf("lead-%d", time.Now().UnixNano())
	doc := map[string]any{
		"id":          leadID,
		"name":        name,
		"email":       email,
		"subject":     subject,
		"message":     message,
		"received_at": time.Now().UTC().Format(time.RFC3339),
		"status":      "new",
	}
	if _, err := drift.NoSQL.Collection("leads").Insert(doc); err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "Could not save your message. Please try again.",
		}
	}

	_ = drift.Queue("contact-queue").Push(map[string]any{
		"lead_id": leadID,
		"name":    name,
		"email":   email,
		"subject": subject,
		"message": message,
	})

	return http.StatusOK, "Message received", map[string]string{
		"message": fmt.Sprintf("Thanks, %s! We'll be in touch soon.", name),
	}
}
