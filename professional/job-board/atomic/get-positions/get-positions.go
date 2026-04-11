// @atomic route=get:get-positions auth=none env=.env
// Returns open positions from the careers board.
// Positions are stored in Backbone Cache under key "positions" as a JSON string.
// Seed it with:
//   drift backbone cache set positions '{"positions":[...]}'
// Update it the same way — the new value takes effect immediately.

package main

import (
	"encoding/json"
	"net/http"

	drift "github.com/ondrift/drift-sdk"
)

func GetGetPositions() (int, string, interface{}) {
	raw, err := drift.Cache.Get("positions")
	if err != nil || len(raw) == 0 {
		return http.StatusOK, "OK", defaultPositions()
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return http.StatusOK, "OK", defaultPositions()
	}

	positions, ok := doc["positions"]
	if !ok {
		return http.StatusOK, "OK", defaultPositions()
	}
	return http.StatusOK, "OK", map[string]any{"positions": positions}
}

func defaultPositions() map[string]any {
	return map[string]any{
		"positions": []map[string]any{
			{"id": "eng-backend", "title": "Backend Engineer", "department": "Engineering", "location": "Remote", "type": "Full-time", "description": "Build and scale our core platform services. You'll work on API design, database optimization, and infrastructure.", "requirements": []string{"3+ years backend experience", "Proficiency in Go or Python", "Experience with PostgreSQL and Redis", "Comfortable with cloud infrastructure (AWS/GCP)"}},
			{"id": "eng-frontend", "title": "Frontend Engineer", "department": "Engineering", "location": "Remote", "type": "Full-time", "description": "Create delightful user experiences for our web application. Component architecture, performance, and accessibility.", "requirements": []string{"3+ years frontend experience", "Expert in React and TypeScript", "Eye for design and UX", "Experience with testing frameworks"}},
			{"id": "design-product", "title": "Product Designer", "department": "Design", "location": "Hybrid — London", "type": "Full-time", "description": "Own the end-to-end design process from research to polished UI. Work closely with engineering and product.", "requirements": []string{"Portfolio demonstrating product thinking", "Proficiency in Figma", "Experience with user research", "Understanding of design systems"}},
			{"id": "ops-devrel", "title": "Developer Relations", "department": "Marketing", "location": "Remote", "type": "Contract", "description": "Be the bridge between our product and the developer community. Create content, speak at events, gather feedback.", "requirements": []string{"Strong technical background", "Excellent writing and speaking skills", "Active in developer communities", "Experience creating technical content"}},
		},
	}
}
