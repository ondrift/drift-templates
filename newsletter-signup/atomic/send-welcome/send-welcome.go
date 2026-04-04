// @atomic route=post:send-welcome auth=none env=.env
// drift:trigger queue signup-queue poll=1000ms retry=3

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		fmt.Println("[send-welcome] RESEND_API_KEY not set — skipping email")
		return http.StatusOK, "Skipped", map[string]string{"reason": "no email key configured"}
	}

	senderEmail := os.Getenv("SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "hello@yourdomain.com"
	}

	name := req.Name
	if name == "" {
		name = "there"
	}

	// Send via Resend (https://resend.com/docs/api-reference/emails/send-email).
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

	httpReq, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return http.StatusInternalServerError, "Email error", map[string]string{"error": err.Error()}
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return http.StatusInternalServerError, "Email error", map[string]string{"error": err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return http.StatusInternalServerError, "Email delivery failed", map[string]string{
			"resend_status": fmt.Sprintf("%d", resp.StatusCode),
		}
	}

	fmt.Printf("[send-welcome] welcome email sent to %s\n", req.Email)
	return http.StatusOK, "Sent", map[string]string{"email": req.Email}
}
