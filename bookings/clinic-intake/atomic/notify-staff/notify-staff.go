// @atomic route=post:notify-staff auth=none env=.env
// drift:trigger queue staff-queue poll=5000ms retry=3

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	drift "github.com/ondrift/drift-sdk"
)

// RequestBody is the message popped from staff-queue by the trigger.
type RequestBody struct {
	AppointmentID   string `json:"appointment_id"`
	PatientEmail    string `json:"patient_email"`
	PreferredDate   string `json:"preferred_date"`
	PreferredTime   string `json:"preferred_time"`
	AppointmentType string `json:"appointment_type"`
	Notes           string `json:"notes"`
}

func PostNotifyStaff(req RequestBody) (int, string, interface{}) {
	if req.PatientEmail == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{"error": "patient_email required"}
	}

	apiKey := drift.Env("RESEND_API_KEY")
	if apiKey == "" {
		drift.Log(fmt.Sprintf("[notify-staff] no RESEND_API_KEY — skipping email for appointment %s", req.AppointmentID))
		return http.StatusOK, "Skipped", map[string]string{"reason": "no email key configured"}
	}

	senderEmail := drift.Env("SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "appointments@yourdomain.com"
	}
	clinicName := drift.Env("CLINIC_NAME")
	if clinicName == "" {
		clinicName = "Our Clinic"
	}
	staffEmail := drift.Env("STAFF_EMAIL")
	if staffEmail == "" {
		drift.Log("[notify-staff] no STAFF_EMAIL configured — skipping notification")
		return http.StatusOK, "Skipped", map[string]string{"reason": "no staff email configured"}
	}

	notes := req.Notes
	if notes == "" {
		notes = "None"
	}

	payload := map[string]any{
		"from":    fmt.Sprintf("%s <%s>", clinicName, senderEmail),
		"to":      []string{staffEmail},
		"subject": fmt.Sprintf("New appointment request — %s on %s at %s", req.AppointmentType, req.PreferredDate, req.PreferredTime),
		"html": fmt.Sprintf(`
<p>A new appointment request has been submitted.</p>
<table style="border-collapse:collapse;margin:1rem 0">
  <tr><td style="padding:4px 12px 4px 0;color:#666">Appointment ID</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Patient Email</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Preferred Date</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Preferred Time</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Type</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Notes</td><td><strong>%s</strong></td></tr>
</table>
<p>Please review and confirm this appointment in the system.</p>
<p>— %s</p>
`, req.AppointmentID, req.PatientEmail, req.PreferredDate, req.PreferredTime, req.AppointmentType, notes, clinicName),
	}

	body, _ := json.Marshal(payload)
	resp, err := drift.HTTPRequest(
		http.MethodPost,
		"https://api.resend.com/emails",
		map[string]string{
			"Authorization": "Bearer " + apiKey,
			"Content-Type":  "application/json",
		},
		body,
	)
	if err != nil {
		return http.StatusInternalServerError, "Email error", map[string]string{"error": err.Error()}
	}
	if resp.Status >= 400 {
		return http.StatusInternalServerError, "Email error", map[string]string{"error": fmt.Sprintf("resend returned %d", resp.Status)}
	}

	drift.Log(fmt.Sprintf("[notify-staff] notification sent to %s for appointment %s", staffEmail, req.AppointmentID))
	return http.StatusOK, "Sent", map[string]string{"staff_email": staffEmail, "appointment_id": req.AppointmentID}
}
