// @atomic route=post:signup auth=none env=.env

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
	Name       string `json:"name"`
	Email      string `json:"email"`
	ReferredBy string `json:"referred_by"`
}

func PostSignup(req RequestBody) (int, string, interface{}) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	name := strings.TrimSpace(req.Name)

	if email == "" || !strings.Contains(email, "@") {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "a valid email address is required",
		}
	}
	if name == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "name is required",
		}
	}

	// Check for duplicate signup.
	raw, err := drift.CacheGet("waitlist:" + email)
	if err == nil && len(raw) > 0 {
		var existing map[string]any
		if json.Unmarshal(raw, &existing) == nil {
			return http.StatusConflict, "Already signed up", map[string]any{
				"error":    "this email is already on the waitlist",
				"position": existing["position"],
			}
		}
	}

	// Increment the global counter to determine position.
	counterRaw, _ := drift.CacheGet("waitlist:counter")
	counter := 0
	if len(counterRaw) > 0 {
		counter, _ = strconv.Atoi(string(counterRaw))
	}
	counter++
	_ = drift.CacheSet("waitlist:counter", strconv.Itoa(counter), 0)

	position := counter
	referralCode := fmt.Sprintf("ref-%d", position)

	// Persist signup to NoSQL.
	doc := map[string]any{
		"email":         email,
		"name":          name,
		"position":      position,
		"referral_code": referralCode,
		"referred_by":   strings.TrimSpace(req.ReferredBy),
		"signed_up_at":  time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := drift.BackboneWrite("signups", doc); err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save signup",
		}
	}

	// Cache the signup for fast duplicate checks and position lookups.
	positionJSON, _ := json.Marshal(map[string]any{
		"position":      position,
		"referral_code": referralCode,
		"name":          name,
	})
	_ = drift.CacheSet("waitlist:"+email, string(positionJSON), 0)

	// Log referral if provided.
	if req.ReferredBy != "" {
		drift.Log(fmt.Sprintf("[signup] %s referred by code %s", email, req.ReferredBy))
	}

	return http.StatusOK, "Signed up", map[string]any{
		"position":      position,
		"referral_code": referralCode,
		"message":       fmt.Sprintf("You're #%d on the waitlist! Share your referral code: %s", position, referralCode),
	}
}
