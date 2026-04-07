// @atomic route=get:get-products auth=none env=.env
// Returns the digital product catalogue.
// The catalogue is stored in Backbone Cache under key "digital-catalogue" as a JSON string.
// Seed it with:
//   drift backbone cache set digital-catalogue '{"products":[...]}'
// Update it the same way — the new value takes effect immediately.

package main

import (
	"encoding/json"
	"net/http"

	drift "github.com/ondrift/drift-sdk"
)

func GetGetProducts() (int, string, interface{}) {
	raw, err := drift.CacheGet("digital-catalogue")
	if err != nil || len(raw) == 0 {
		return http.StatusOK, "OK", defaultCatalogue()
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return http.StatusOK, "OK", defaultCatalogue()
	}

	products, ok := doc["products"]
	if !ok {
		return http.StatusOK, "OK", defaultCatalogue()
	}
	return http.StatusOK, "OK", map[string]any{"products": products}
}

func defaultCatalogue() map[string]any {
	return map[string]any{
		"products": []map[string]any{
			{"id": "ebook-startup", "name": "The Startup Playbook", "description": "A practical guide to launching your first product. 180 pages of hard-won lessons.", "price": 19.99, "format": "PDF"},
			{"id": "course-go", "name": "Go from Zero to Production", "description": "Video course — 8 modules covering everything from basics to deploying production services.", "price": 49.99, "format": "Video"},
			{"id": "template-landing", "name": "Landing Page Template Pack", "description": "12 conversion-optimised landing page templates. HTML/CSS, ready to customise.", "price": 29.99, "format": "ZIP"},
		},
	}
}
