// @atomic route=post:confirm-reservation auth=none env=.env
// drift:trigger queue reservation-queue poll=2000ms retry=3

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// RequestBody is the message popped from reservation-queue by the trigger.
type RequestBody struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	PartySize   int    `json:"party_size"`
	ConfirmCode string `json:"confirm_code"`
}

func PostConfirmReservation(req RequestBody) (int, string, interface{}) {
	if req.Email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{"error": "email required"}
	}

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		fmt.Printf("[confirm-reservation] no RESEND_API_KEY — skipping email for %s\n", req.Email)
		return http.StatusOK, "Skipped", map[string]string{"reason": "no email key configured"}
	}

	senderEmail := os.Getenv("SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "reservations@yourdomain.com"
	}
	restaurantName := os.Getenv("RESTAURANT_NAME")
	if restaurantName == "" {
		restaurantName = "Our Restaurant"
	}

	payload := map[string]any{
		"from":    fmt.Sprintf("%s <%s>", restaurantName, senderEmail),
		"to":      []string{req.Email},
		"subject": fmt.Sprintf("Reservation confirmed — %s at %s", req.Date, req.Time),
		"html": fmt.Sprintf(`
<p>Hi %s,</p>
<p>Your table is confirmed at <strong>%s</strong>.</p>
<table style="border-collapse:collapse;margin:1rem 0">
  <tr><td style="padding:4px 12px 4px 0;color:#666">Date</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Time</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Party</td><td><strong>%d guest(s)</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Code</td><td><strong>%s</strong></td></tr>
</table>
<p>Need to change or cancel? Reply to this email.</p>
<p>We look forward to seeing you!<br>– %s</p>
`, req.Name, restaurantName, req.Date, req.Time, req.PartySize, req.ConfirmCode, restaurantName),
	}

	body, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return http.StatusInternalServerError, "Email error", map[string]string{"error": err.Error()}
	}
	defer resp.Body.Close()

	fmt.Printf("[confirm-reservation] confirmation sent to %s (code %s)\n", req.Email, req.ConfirmCode)
	return http.StatusOK, "Sent", map[string]string{"email": req.Email, "code": req.ConfirmCode}
}
