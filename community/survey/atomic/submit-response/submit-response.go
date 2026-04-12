// @atomic route=post:submit-response auth=none env=.env

package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	drift "github.com/ondrift/drift-sdk/go"
)

type Answer struct {
	QuestionID string `json:"question_id"`
	Value      string `json:"value"`
}

type RequestBody struct {
	RespondentEmail string   `json:"respondent_email"`
	Answers         []Answer `json:"answers"`
}

func PostSubmitResponse(req RequestBody) (int, string, interface{}) {
	email := strings.TrimSpace(strings.ToLower(req.RespondentEmail))

	if email == "" || !strings.Contains(email, "@") {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "a valid respondent_email is required",
		}
	}
	if len(req.Answers) == 0 {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "at least one answer is required",
		}
	}

	// Check for duplicate submission.
	existing, err := drift.Cache.Get("survey:resp:" + email)
	if err == nil && len(existing) > 0 {
		return http.StatusConflict, "Conflict", map[string]string{
			"error": "you have already submitted a response",
		}
	}

	// Generate a short response ID.
	nano := fmt.Sprintf("%d", time.Now().UnixNano())
	responseID := "resp-" + nano[len(nano)-8:]

	// Build answer list.
	answers := make([]map[string]any, len(req.Answers))
	for i, a := range req.Answers {
		answers[i] = map[string]any{
			"question_id": a.QuestionID,
			"value":       a.Value,
		}
	}

	doc := map[string]any{
		"id":               responseID,
		"respondent_email": email,
		"answers":          answers,
		"submitted_at":     time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := drift.NoSQL.Collection("responses").Insert(doc); err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save response",
		}
	}

	// Mark as submitted in cache to prevent duplicates.
	_ = drift.Cache.Set("survey:resp:"+email, []byte("1"), 0)

	// Update response count in cache.
	countRaw, err := drift.Cache.Get("survey:count")
	count := 0
	if err == nil && len(countRaw) > 0 {
		count, _ = strconv.Atoi(string(countRaw))
	}
	count++
	_ = drift.Cache.Set("survey:count", []byte(strconv.Itoa(count)), 0)

	drift.Log(fmt.Sprintf("[submit-response] response %s saved from %s (total: %d)", responseID, email, count))

	return http.StatusOK, "Response recorded", map[string]any{
		"response_id": responseID,
		"message":     "Thank you for your feedback! Your response has been recorded.",
		"total_count": count,
	}
}
