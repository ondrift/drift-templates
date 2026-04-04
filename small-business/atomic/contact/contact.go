// @atomic route=post:contact auth=none env=.env
// Stores a contact form submission in the leads NoSQL collection and
// pushes it to the contact-queue so the owner gets an email notification.

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

	backbone := os.Getenv("BACKBONE_URL")
	if backbone == "" {
		backbone = "http://backbone:8000"
	}

	leadID := fmt.Sprintf("lead-%d", time.Now().UnixNano())
	doc := map[string]any{
		"collection":  "leads",
		"id":          leadID,
		"name":        name,
		"email":       email,
		"subject":     subject,
		"message":     message,
		"received_at": time.Now().UTC().Format(time.RFC3339),
		"status":      "new",
	}
	docBytes, _ := json.Marshal(doc)
	writeResp, err := http.Post(fmt.Sprintf("%s/write", backbone), "application/json", bytes.NewReader(docBytes))
	if err != nil || writeResp.StatusCode != http.StatusOK {
		if writeResp != nil {
			writeResp.Body.Close()
		}
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "Could not save your message. Please try again.",
		}
	}
	writeResp.Body.Close()

	// Push notification to queue for owner email
	queuePayload, _ := json.Marshal(map[string]any{
		"queue": "contact-queue",
		"body": map[string]any{
			"lead_id": leadID,
			"name":    name,
			"email":   email,
			"subject": subject,
			"message": message,
		},
	})
	pushResp, _ := http.Post(fmt.Sprintf("%s/queue/push", backbone), "application/json", bytes.NewReader(queuePayload))
	if pushResp != nil {
		pushResp.Body.Close()
	}

	return http.StatusOK, "Message received", map[string]string{
		"message": fmt.Sprintf("Thanks, %s! We'll be in touch soon.", name),
	}
}
