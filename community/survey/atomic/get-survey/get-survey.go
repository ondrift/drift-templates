// @atomic route=get:get-survey auth=none env=.env
// Returns the survey definition (title, description, questions).
// The survey is stored in Backbone Cache under key "survey-definition" as a JSON string.
// Seed it with:
//   drift backbone cache set survey-definition '{"title":"...","questions":[...]}'
// Update it the same way — the new value takes effect immediately.

package main

import (
	"encoding/json"
	"net/http"

	drift "github.com/ondrift/drift-sdk"
)

func GetGetSurvey() (int, string, interface{}) {
	raw, err := drift.Cache.Get("survey-definition")
	if err != nil || len(raw) == 0 {
		return http.StatusOK, "OK", defaultSurvey()
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return http.StatusOK, "OK", defaultSurvey()
	}

	return http.StatusOK, "OK", doc
}

func defaultSurvey() map[string]any {
	return map[string]any{
		"title":       "Customer Feedback Survey",
		"description": "Help us improve — takes less than 2 minutes.",
		"questions": []map[string]any{
			{"id": "q1", "text": "How satisfied are you with our service?", "type": "rating", "options": []string{"1", "2", "3", "4", "5"}},
			{"id": "q2", "text": "What do you like most about our product?", "type": "choice", "options": []string{"Ease of use", "Performance", "Price", "Support", "Features"}},
			{"id": "q3", "text": "What could we improve?", "type": "text", "options": []string{}},
			{"id": "q4", "text": "How likely are you to recommend us to a friend?", "type": "rating", "options": []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}},
			{"id": "q5", "text": "Any additional comments?", "type": "text", "options": []string{}},
		},
	}
}
