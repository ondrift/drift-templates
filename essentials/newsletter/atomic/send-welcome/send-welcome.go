// @atomic route=post:send-welcome auth=none env=.env
// drift:trigger queue signup-queue poll=1000ms retry=3

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	drift "github.com/ondrift/drift-sdk"
)

// RequestBody is the message popped from the signup-queue by the trigger.
type RequestBody struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func PostSendWelcome(req RequestBody) (int, string, interface{}) {
	if req.Email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "email is required",
		}
	}

	apiKey := drift.Env("RESEND_API_KEY")
	if apiKey == "" {
		drift.Log("[send-welcome] RESEND_API_KEY not set — skipping email")
		return http.StatusOK, "Skipped", map[string]string{"reason": "no email key configured"}
	}

	senderEmail := drift.Env("SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "hello@yourdomain.com"
	}

	name := req.Name
	if name == "" {
		name = "there"
	}

	payload := map[string]any{
		"from":    fmt.Sprintf("Your Newsletter <%s>", senderEmail),
		"to":      []string{req.Email},
		"subject": "Welcome — you're on the list!",
		"html": fmt.Sprintf(`
<p>Hi %s,</p>
<p>Thanks for subscribing! You'll hear from us when we have something worth saying.</p>
<p>– The Team</p>
<p style="font-size:12px;color:#888"><a href="https://yourdomain.com/api/unsubscribe?email=%s">Unsubscribe</a></p>
`, name, req.Email),
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
		return http.StatusInternalServerError, "Email delivery failed", map[string]string{
			"resend_status": fmt.Sprintf("%d", resp.Status),
		}
	}

	drift.Log(fmt.Sprintf("[send-welcome] welcome email sent to %s", req.Email))
	return http.StatusOK, "Sent", map[string]string{"email": req.Email}
}
