// @atomic route=get:get-projects auth=none env=.env
// Returns the portfolio project list.
// Projects are stored in Backbone Cache under key "projects" as a JSON string.
// Seed it with:
//   drift backbone cache set projects '{"projects":[...]}'
// Update it the same way — the new value takes effect immediately.

package main

import (
	"encoding/json"
	"net/http"

	"drift-sdk"
)

func GetGetProjects() (int, string, interface{}) {
	raw, err := drift.CacheGet("projects")
	if err != nil || len(raw) == 0 {
		return http.StatusOK, "OK", defaultProjects()
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return http.StatusOK, "OK", defaultProjects()
	}

	projects, ok := doc["projects"]
	if !ok {
		return http.StatusOK, "OK", defaultProjects()
	}
	return http.StatusOK, "OK", map[string]any{"projects": projects}
}

func defaultProjects() map[string]any {
	return map[string]any{
		"projects": []map[string]any{
			{"id": "project-saas", "title": "SaaS Dashboard Redesign", "description": "Redesigned the analytics dashboard for a B2B SaaS product. Improved user retention by 23%.", "tags": []string{"UI/UX", "React", "TypeScript"}, "url": "https://example.com", "image_url": "https://picsum.photos/seed/saas/600/400"},
			{"id": "project-mobile", "title": "Fitness Tracking App", "description": "Built a cross-platform mobile app for tracking workouts and nutrition with social features.", "tags": []string{"React Native", "Node.js", "PostgreSQL"}, "url": "https://example.com", "image_url": "https://picsum.photos/seed/fitness/600/400"},
			{"id": "project-ecommerce", "title": "Artisan Marketplace", "description": "Full-stack e-commerce platform for local artisans. Stripe integration, real-time inventory.", "tags": []string{"Next.js", "Stripe", "Tailwind"}, "url": "https://example.com", "image_url": "https://picsum.photos/seed/market/600/400"},
			{"id": "project-api", "title": "IoT Data Pipeline", "description": "High-throughput data pipeline processing 2M events/day from IoT sensors. 99.99% uptime.", "tags": []string{"Go", "Kafka", "TimescaleDB"}, "url": "https://example.com", "image_url": "https://picsum.photos/seed/iot/600/400"},
		},
	}
}
