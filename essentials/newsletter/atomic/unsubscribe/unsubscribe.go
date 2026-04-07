// @atomic route=post:unsubscribe auth=none env=.env

package main

import (
	"fmt"
	"net/http"
	"strings"

	drift "github.com/ondrift/drift-sdk"
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

	// Write a tombstone document — NoSQL collections are append-only; mark
	// unsubscribed=true so the send-welcome trigger skips this address.
	doc := map[string]any{
		"email":        email,
		"subscribed":   false,
		"unsubscribed": true,
	}
	if _, err := drift.BackboneWrite("subscribers", doc); err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to process unsubscribe request",
		}
	}

	// Evict the cache entry so re-subscribe works immediately.
	_ = drift.CacheDel(fmt.Sprintf("sub:%s", email))

	return http.StatusOK, "Unsubscribed", map[string]string{
		"message": "You have been unsubscribed. You won't receive any more emails from us.",
	}
}
