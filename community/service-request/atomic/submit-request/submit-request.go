// @atomic route=post:submit-request auth=none env=.env

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"drift-sdk"
)

type RequestBody struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Category    string `json:"category"`    // pothole, streetlight, graffiti, noise, sidewalk, other
	Location    string `json:"location"`
	Description string `json:"description"`
}

func PostSubmitRequest(req RequestBody) (int, string, interface{}) {
	if req.Name == "" || req.Email == "" || req.Category == "" || req.Description == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "name, email, category, and description are required",
		}
	}

	validCategories := map[string]bool{
		"pothole": true, "streetlight": true, "graffiti": true,
		"noise": true, "sidewalk": true, "other": true,
	}
	if !validCategories[req.Category] {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "category must be one of: pothole, streetlight, graffiti, noise, sidewalk, other",
		}
	}

	// Generate a short, memorable ticket number.
	ticket := fmt.Sprintf("SR-%d", time.Now().UnixNano()%1000000)

	submittedAt := time.Now().UTC().Format(time.RFC3339)

	doc := map[string]any{
		"ticket":       ticket,
		"name":         req.Name,
		"email":        req.Email,
		"category":     req.Category,
		"location":     req.Location,
		"description":  req.Description,
		"status":       "open",
		"submitted_at": submittedAt,
	}
	_, err := drift.BackboneWrite("requests", doc)
	if err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save request",
		}
	}

	// Cache for quick status lookups.
	statusJSON, _ := json.Marshal(map[string]string{
		"ticket":       ticket,
		"category":     req.Category,
		"status":       "open",
		"submitted_at": submittedAt,
	})
	_ = drift.CacheSet("ticket:"+ticket, string(statusJSON), 0)

	// Enqueue department notification.
	_ = drift.QueuePush("department-queue", map[string]any{
		"ticket":      ticket,
		"name":        req.Name,
		"email":       req.Email,
		"category":    req.Category,
		"location":    req.Location,
		"description": req.Description,
	})

	return http.StatusOK, "Request received", map[string]any{
		"ticket":  ticket,
		"message": fmt.Sprintf("Your request has been logged. Track it with ticket number %s", ticket),
	}
}
