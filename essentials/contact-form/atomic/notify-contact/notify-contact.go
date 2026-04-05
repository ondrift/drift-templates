// @atomic route=post:notify-contact auth=none env=.env
// drift:trigger queue contact-queue poll=5000ms retry=3
//
// Consumes the contact-queue and forwards new lead details to the
// business owner via Resend.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"drift-sdk"
)

type RequestBody struct {
	LeadID  string `json:"lead_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func PostNotifyContact(req RequestBody) (int, string, interface{}) {
	apiKey := drift.Env("RESEND_API_KEY")
	sender := drift.Env("SENDER_EMAIL")
	owner := drift.Env("OWNER_EMAIL")

	if apiKey == "" || sender == "" || owner == "" {
		return http.StatusOK, "No email configured", map[string]string{
			"note": "Set RESEND_API_KEY, SENDER_EMAIL, and OWNER_EMAIL secrets to enable notifications",
		}
	}

	subject := fmt.Sprintf("[New Lead] %s — %s", req.Subject, req.Name)
	html := fmt.Sprintf(`
<p>A new message arrived via your contact form:</p>
<table style="border-collapse:collapse;width:100%%;max-width:560px">
  <tr><td style="padding:6px 12px;color:#555;width:120px"><strong>From</strong></td><td style="padding:6px 12px">%s &lt;%s&gt;</td></tr>
  <tr style="background:#f9fafb"><td style="padding:6px 12px;color:#555"><strong>Subject</strong></td><td style="padding:6px 12px">%s</td></tr>
  <tr><td style="padding:6px 12px;color:#555"><strong>Message</strong></td><td style="padding:6px 12px">%s</td></tr>
  <tr style="background:#f9fafb"><td style="padding:6px 12px;color:#555"><strong>Lead ID</strong></td><td style="padding:6px 12px;font-size:0.85em;color:#888">%s</td></tr>
</table>
<p style="margin-top:1.5rem">Reply directly to <a href="mailto:%s">%s</a> to respond.</p>`,
		req.Name, req.Email,
		req.Subject,
		req.Message,
		req.LeadID,
		req.Email, req.Email,
	)

	body, _ := json.Marshal(map[string]any{
		"from":     sender,
		"to":       []string{owner},
		"reply_to": req.Email,
		"subject":  subject,
		"html":     html,
	})

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
		return http.StatusInternalServerError, "Email error", map[string]string{
			"error": "Failed to send notification",
		}
	}
	if resp.Status >= 400 {
		return http.StatusInternalServerError, "Email error", map[string]string{
			"error": fmt.Sprintf("resend returned %d", resp.Status),
		}
	}

	drift.Log(fmt.Sprintf("[notify-contact] owner notified about lead from %s", req.Email))
	return http.StatusOK, "Notification sent", map[string]string{
		"message": fmt.Sprintf("Owner notified about lead from %s", req.Email),
	}
}
