// @atomic route=post:subscribe auth=none env=.env

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

	backbone := os.Getenv("BACKBONE_URL")
	if backbone == "" {
		backbone = "http://backbone:8000"
	}

	// Check for existing subscriber in cache (fast path).
	cacheResp, err := http.Get(fmt.Sprintf("%s/cache/get?key=sub:%s", backbone, email))
	if err == nil && cacheResp.StatusCode == http.StatusOK {
		cacheResp.Body.Close()
		return http.StatusConflict, "Already subscribed", map[string]string{
			"error": "this email is already subscribed",
		}
	}
	if cacheResp != nil {
		cacheResp.Body.Close()
	}

	// Write subscriber to NoSQL.
	doc := map[string]any{
		"collection": "subscribers",
		"email":      email,
		"name":       req.Name,
		"subscribed": true,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}
	docBytes, _ := json.Marshal(doc)
	writeResp, err := http.Post(fmt.Sprintf("%s/write", backbone), "application/json", bytes.NewReader(docBytes))
	if err != nil || writeResp.StatusCode != http.StatusOK {
		if writeResp != nil {
			writeResp.Body.Close()
		}
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save subscription",
		}
	}
	writeResp.Body.Close()

	// Mark in cache so duplicate checks are fast (24h TTL).
	cachePayload, _ := json.Marshal(map[string]any{
		"key":   fmt.Sprintf("sub:%s", email),
		"value": "1",
		"ttl":   86400,
	})
	cacheSet, _ := http.Post(fmt.Sprintf("%s/cache/set", backbone), "application/json", bytes.NewReader(cachePayload))
	if cacheSet != nil {
		cacheSet.Body.Close()
	}

	// Enqueue welcome email.
	queuePayload, _ := json.Marshal(map[string]any{
		"queue": "signup-queue",
		"body":  map[string]string{"email": email, "name": req.Name},
	})
	queueResp, _ := http.Post(fmt.Sprintf("%s/queue/push", backbone), "application/json", bytes.NewReader(queuePayload))
	if queueResp != nil {
		queueResp.Body.Close()
	}

	return http.StatusOK, "Subscribed", map[string]string{
		"message": "You're on the list! Check your inbox for a welcome email.",
	}
}
