// @atomic route=post:confirm-booking auth=none env=.env
// drift:trigger queue booking-queue poll=2000ms retry=3
//
// Consumes the booking-queue and sends a confirmation email via Resend.
// Required env vars (set as Backbone Secrets):
//   RESEND_API_KEY   - Resend API key
//   SENDER_EMAIL     - verified sender address, e.g. bookings@yourdomain.com
//   BUSINESS_NAME    - displayed in the email subject/body

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type RequestBody struct {
	BookingID string `json:"booking_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Date      string `json:"date"`
	TimeSlot  string `json:"time_slot"`
	Service   string `json:"service"`
}

func PostConfirmBooking(req RequestBody) (int, string, interface{}) {
	apiKey := os.Getenv("RESEND_API_KEY")
	sender := os.Getenv("SENDER_EMAIL")
	business := os.Getenv("BUSINESS_NAME")
	if business == "" {
		business = "Our business"
	}

	if apiKey == "" || sender == "" {
		// No email configured — silently succeed so the queue item is consumed
		return http.StatusOK, "No email configured", map[string]string{
			"note": "Set RESEND_API_KEY and SENDER_EMAIL secrets to enable confirmation emails",
		}
	}

	subject := fmt.Sprintf("Booking confirmed: %s on %s at %s", req.Service, req.Date, req.TimeSlot)
	html := fmt.Sprintf(`
<p>Hi %s,</p>
<p>Your booking with <strong>%s</strong> is confirmed!</p>
<ul>
  <li><strong>Service:</strong> %s</li>
  <li><strong>Date:</strong> %s</li>
  <li><strong>Time:</strong> %s</li>
  <li><strong>Booking ID:</strong> %s</li>
</ul>
<p>Need to cancel? Reply to this email with your booking ID.</p>
<p>— %s</p>`,
		req.Name, business,
		req.Service, req.Date, req.TimeSlot, req.BookingID,
		business,
	)

	payload, _ := json.Marshal(map[string]any{
		"from":    sender,
		"to":      []string{req.Email},
		"subject": subject,
		"html":    html,
	})

	httpReq, _ := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(payload))
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil || resp.StatusCode >= 400 {
		if resp != nil {
			resp.Body.Close()
		}
		return http.StatusInternalServerError, "Email error", map[string]string{
			"error": "Failed to send confirmation email",
		}
	}
	resp.Body.Close()

	return http.StatusOK, "Email sent", map[string]string{
		"message": fmt.Sprintf("Confirmation sent to %s", req.Email),
	}
}
