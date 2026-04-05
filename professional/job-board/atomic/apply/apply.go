// @atomic route=post:apply auth=none env=.env

package main

import (
	"fmt"
	"net/http"
	"time"

	"drift-sdk"
)

type RequestBody struct {
	PositionID  string `json:"position_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	CoverLetter string `json:"cover_letter"`
	LinkedInURL string `json:"linkedin_url"`
}

func PostApply(req RequestBody) (int, string, interface{}) {
	if req.PositionID == "" || req.Name == "" || req.Email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "position_id, name, and email are required",
		}
	}

	applicationID := fmt.Sprintf("app-%d", time.Now().UnixNano())

	doc := map[string]any{
		"id":           applicationID,
		"position_id":  req.PositionID,
		"name":         req.Name,
		"email":        req.Email,
		"phone":        req.Phone,
		"cover_letter": req.CoverLetter,
		"linkedin_url": req.LinkedInURL,
		"applied_at":   time.Now().UTC().Format(time.RFC3339),
		"status":       "received",
	}
	_, err := drift.BackboneWrite("applications", doc)
	if err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save application",
		}
	}

	_ = drift.QueuePush("hiring-queue", map[string]any{
		"application_id": applicationID,
		"position_id":    req.PositionID,
		"name":           req.Name,
		"email":          req.Email,
		"phone":          req.Phone,
		"cover_letter":   req.CoverLetter,
		"linkedin_url":   req.LinkedInURL,
	})

	return http.StatusOK, "Application received", map[string]any{
		"application_id": applicationID,
		"message":        "Application received! We'll be in touch within 5 business days.",
	}
}
