// @atomic route=post:subscribe auth=none env=.env

package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	drift "github.com/ondrift/drift-sdk/go"
)

type RequestBody struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func PostSubscribe(req RequestBody) (int, string, interface{}) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "a valid email address is required",
		}
	}

	// Check for existing subscriber in cache (fast path).
	raw, err := drift.Cache.Get(fmt.Sprintf("sub:%s", email))
	if err == nil && len(raw) > 0 {
		return http.StatusConflict, "Already subscribed", map[string]string{
			"error": "this email is already subscribed",
		}
	}

	// Write subscriber to NoSQL.
	doc := map[string]any{
		"email":      email,
		"name":       req.Name,
		"subscribed": true,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := drift.NoSQL.Collection("subscribers").Insert(doc); err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save subscription",
		}
	}

	// Mark in cache so duplicate checks are fast (24h TTL).
	_ = drift.Cache.Set(fmt.Sprintf("sub:%s", email), "1", 86400)

	// Enqueue welcome email.
	_ = drift.Queue("signup-queue").Push(map[string]string{"email": email, "name": req.Name})

	return http.StatusOK, "Subscribed", map[string]string{
		"message": "You're on the list! Check your inbox for a welcome email.",
	}
}
