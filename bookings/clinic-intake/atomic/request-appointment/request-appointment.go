// @atomic route=post:request-appointment auth=none env=.env

package main

import (
	"fmt"
	"net/http"
	"time"

	"drift-sdk"
)

type RequestBody struct {
	PatientEmail    string `json:"patient_email"`
	PreferredDate   string `json:"preferred_date"`
	PreferredTime   string `json:"preferred_time"`
	AppointmentType string `json:"appointment_type"`
	Notes           string `json:"notes"`
}

func PostRequestAppointment(req RequestBody) (int, string, interface{}) {
	if req.PatientEmail == "" || req.PreferredDate == "" || req.PreferredTime == "" || req.AppointmentType == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "patient_email, preferred_date, preferred_time, and appointment_type are required",
		}
	}

	// Look up patient by email from cache — must exist.
	patientID, err := drift.CacheGet("patient:" + req.PatientEmail)
	if err != nil || len(patientID) == 0 {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "no patient record found for this email — please complete the intake form first",
		}
	}

	appointmentID := fmt.Sprintf("appt-%d", time.Now().UnixNano())

	doc := map[string]any{
		"id":               appointmentID,
		"patient_id":       string(patientID),
		"patient_email":    req.PatientEmail,
		"preferred_date":   req.PreferredDate,
		"preferred_time":   req.PreferredTime,
		"appointment_type": req.AppointmentType,
		"notes":            req.Notes,
		"status":           "requested",
		"requested_at":     time.Now().UTC().Format(time.RFC3339),
	}
	_, err = drift.BackboneWrite("appointments", doc)
	if err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save appointment request",
		}
	}

	// Enqueue staff notification.
	_ = drift.QueuePush("staff-queue", map[string]any{
		"appointment_id":   appointmentID,
		"patient_email":    req.PatientEmail,
		"preferred_date":   req.PreferredDate,
		"preferred_time":   req.PreferredTime,
		"appointment_type": req.AppointmentType,
		"notes":            req.Notes,
	})

	return http.StatusOK, "Appointment requested", map[string]any{
		"appointment_id": appointmentID,
		"message":        fmt.Sprintf("Your appointment request has been submitted. We'll confirm within 24 hours. Reference: %s", appointmentID),
	}
}
