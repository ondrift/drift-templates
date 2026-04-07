// @atomic route=post:confirm-class auth=none env=.env
// drift:trigger queue class-queue poll=2000ms retry=3

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	drift "github.com/ondrift/drift-sdk"
)

// RequestBody is the message popped from class-queue by the trigger.
type RequestBody struct {
	BookingID  string `json:"booking_id"`
	ClassID    string `json:"class_id"`
	ClassName  string `json:"class_name"`
	Date       string `json:"date"`
	Time       string `json:"time"`
	Duration   string `json:"duration"`
	Location   string `json:"location"`
	Instructor string `json:"instructor"`
	Name       string `json:"name"`
	Email      string `json:"email"`
}

func PostConfirmClass(req RequestBody) (int, string, interface{}) {
	if req.Email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{"error": "email required"}
	}

	apiKey := drift.Env("RESEND_API_KEY")
	if apiKey == "" {
		drift.Log(fmt.Sprintf("[confirm-class] no RESEND_API_KEY — skipping email for %s", req.Email))
		return http.StatusOK, "Skipped", map[string]string{"reason": "no email key configured"}
	}

	senderEmail := drift.Env("SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "bookings@yourdomain.com"
	}
	studioName := drift.Env("STUDIO_NAME")
	if studioName == "" {
		studioName = "My Studio"
	}

	payload := map[string]any{
		"from":    fmt.Sprintf("%s <%s>", studioName, senderEmail),
		"to":      []string{req.Email},
		"subject": fmt.Sprintf("Class confirmed — %s on %s", req.ClassName, req.Date),
		"html": fmt.Sprintf(`
<p>Hi %s,</p>
<p>You're all set for your class at <strong>%s</strong>!</p>
<table style="border-collapse:collapse;margin:1rem 0">
  <tr><td style="padding:4px 12px 4px 0;color:#666">Class</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Date</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Time</td><td><strong>%s (%s)</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Location</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Instructor</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Booking ID</td><td><strong>%s</strong></td></tr>
</table>
<p>Please arrive 10 minutes early. Bring water and a towel.</p>
<p>See you there!<br>– %s</p>
`, req.Name, studioName, req.ClassName, req.Date, req.Time, req.Duration, req.Location, req.Instructor, req.BookingID, studioName),
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

	drift.Log(fmt.Sprintf("[confirm-class] confirmation sent to %s (booking %s)", req.Email, req.BookingID))
	return http.StatusOK, "Sent", map[string]string{"email": req.Email, "booking_id": req.BookingID}
}
