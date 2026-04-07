// @atomic route=post:notify-inquiry auth=none env=.env
// drift:trigger queue inquiry-queue poll=5000ms retry=3

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	drift "github.com/ondrift/drift-sdk"
)

// RequestBody is the message popped from inquiry-queue by the trigger.
type RequestBody struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
	Budget  string `json:"budget"`
}

func PostNotifyInquiry(req RequestBody) (int, string, interface{}) {
	if req.Email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{"error": "email required"}
	}

	apiKey := drift.Env("RESEND_API_KEY")
	if apiKey == "" {
		drift.Log(fmt.Sprintf("[notify-inquiry] no RESEND_API_KEY — skipping email for inquiry %s", req.ID))
		return http.StatusOK, "Skipped", map[string]string{"reason": "no email key configured"}
	}

	senderEmail := drift.Env("SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "portfolio@yourdomain.com"
	}
	ownerName := drift.Env("OWNER_NAME")
	if ownerName == "" {
		ownerName = "Portfolio Owner"
	}
	ownerEmail := drift.Env("OWNER_EMAIL")
	if ownerEmail == "" {
		drift.Log(fmt.Sprintf("[notify-inquiry] no OWNER_EMAIL — skipping notification for inquiry %s", req.ID))
		return http.StatusOK, "Skipped", map[string]string{"reason": "no owner email configured"}
	}

	subject := req.Subject
	if subject == "" {
		subject = "New inquiry"
	}

	budget := req.Budget
	if budget == "" {
		budget = "Not specified"
	}

	payload := map[string]any{
		"from":     fmt.Sprintf("%s Portfolio <%s>", ownerName, senderEmail),
		"to":       []string{ownerEmail},
		"reply_to": req.Email,
		"subject":  fmt.Sprintf("New portfolio inquiry from %s — %s", req.Name, subject),
		"html": fmt.Sprintf(`
<h2>New Portfolio Inquiry</h2>
<table style="border-collapse:collapse;margin:1rem 0;font-family:sans-serif">
  <tr><td style="padding:6px 16px 6px 0;color:#666;vertical-align:top">From</td><td><strong>%s</strong> (%s)</td></tr>
  <tr><td style="padding:6px 16px 6px 0;color:#666;vertical-align:top">Subject</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:6px 16px 6px 0;color:#666;vertical-align:top">Budget</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:6px 16px 6px 0;color:#666;vertical-align:top">Inquiry ID</td><td>%s</td></tr>
</table>
<h3>Message</h3>
<p style="background:#f5f5f5;padding:16px;border-radius:6px;line-height:1.6">%s</p>
<p style="color:#999;font-size:0.85em">Reply directly to this email to respond to %s.</p>
`, req.Name, req.Email, subject, budget, req.ID, req.Message, req.Name),
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

	drift.Log(fmt.Sprintf("[notify-inquiry] notification sent to %s for inquiry %s from %s", ownerEmail, req.ID, req.Email))
	return http.StatusOK, "Sent", map[string]string{"inquiry_id": req.ID, "notified": ownerEmail}
}
