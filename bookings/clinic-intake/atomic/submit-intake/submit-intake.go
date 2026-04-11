// @atomic route=post:submit-intake auth=none env=.env

package main

import (
	"fmt"
	"net/http"
	"time"

	drift "github.com/ondrift/drift-sdk"
)

type RequestBody struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	DateOfBirth       string `json:"date_of_birth"`
	Reason            string `json:"reason"`
	Allergies         string `json:"allergies"`
	Medications       string `json:"medications"`
	InsuranceProvider string `json:"insurance_provider"`
}

func PostSubmitIntake(req RequestBody) (int, string, interface{}) {
	if req.Name == "" || req.Email == "" || req.Phone == "" || req.DateOfBirth == "" || req.Reason == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "name, email, phone, date_of_birth, and reason are required",
		}
	}

	patientID := fmt.Sprintf("patient-%d", time.Now().UnixNano())

	doc := map[string]any{
		"id":                 patientID,
		"name":               req.Name,
		"email":              req.Email,
		"phone":              req.Phone,
		"date_of_birth":      req.DateOfBirth,
		"reason":             req.Reason,
		"allergies":          req.Allergies,
		"medications":        req.Medications,
		"insurance_provider": req.InsuranceProvider,
		"created_at":         time.Now().UTC().Format(time.RFC3339),
	}
	_, err := drift.NoSQL.Collection("patients").Insert(doc)
	if err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save patient record",
		}
	}

	// Store patient ID in cache for quick lookup by email.
	_ = drift.Cache.Set("patient:"+req.Email, patientID, 0)

	return http.StatusOK, "Patient registered", map[string]any{
		"patient_id": patientID,
		"message":    fmt.Sprintf("Thank you %s! Your intake form has been submitted. Your patient ID is %s.", req.Name, patientID),
	}
}
