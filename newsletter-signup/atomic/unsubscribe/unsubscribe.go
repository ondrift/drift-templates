// @atomic route=post:unsubscribe auth=none env=.env

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type RequestBody struct {
	Email string `json:"email"`
}

func PostUnsubscribe(req RequestBody) (int, string, interface{}) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "email is required",
		}
	}

	backbone := os.Getenv("BACKBONE_URL")
	if backbone == "" {
		backbone = "http://backbone:8000"
	}

	// Write a tombstone document — NoSQL collections are append-only; mark
	// unsubscribed=true so the send-welcome trigger skips this address.
	doc := map[string]any{
		"collection":   "subscribers",
		"email":        email,
		"subscribed":   false,
		"unsubscribed": true,
	}
	docBytes, _ := json.Marshal(doc)
	resp, err := http.Post(fmt.Sprintf("%s/write", backbone), "application/json", bytes.NewReader(docBytes))
	if err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to process unsubscribe request",
		}
	}
	resp.Body.Close()

	// Evict the cache entry so re-subscribe works immediately.
	evictPayload, _ := json.Marshal(map[string]string{
		"key": fmt.Sprintf("sub:%s", email),
	})
	delResp, _ := http.Post(fmt.Sprintf("%s/cache/del", backbone), "application/json", bytes.NewReader(evictPayload))
	if delResp != nil {
		delResp.Body.Close()
	}

	return http.StatusOK, "Unsubscribed", map[string]string{
		"message": "You have been unsubscribed. You won't receive any more emails from us.",
	}
}
