// @atomic route=post:deliver auth=none env=.env
// drift:trigger queue delivery-queue poll=3000ms retry=3

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	drift "github.com/ondrift/drift-sdk/go"
)

// RequestBody is the message popped from delivery-queue by the trigger.
type RequestBody struct {
	PurchaseID  string `json:"purchase_id"`
	ProductID   string `json:"product_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	AccessToken string `json:"access_token"`
}

func PostDeliver(req RequestBody) (int, string, interface{}) {
	if req.Email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{"error": "email required"}
	}

	apiKey := drift.Env("RESEND_API_KEY")
	if apiKey == "" {
		drift.Log(fmt.Sprintf("[deliver] no RESEND_API_KEY — skipping email for %s", req.Email))
		return http.StatusOK, "Skipped", map[string]string{"reason": "no email key configured"}
	}

	senderEmail := drift.Env("SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "store@yourdomain.com"
	}
	storeName := drift.Env("STORE_NAME")
	if storeName == "" {
		storeName = "My Digital Store"
	}

	// Build a download link using the access token.
	// Replace this URL with your actual file hosting / signed-URL logic.
	downloadLink := fmt.Sprintf("https://yourdomain.com/download?token=%s", req.AccessToken)

	payload := map[string]any{
		"from":    fmt.Sprintf("%s <%s>", storeName, senderEmail),
		"to":      []string{req.Email},
		"subject": fmt.Sprintf("Your purchase is ready — %s", req.ProductID),
		"html": fmt.Sprintf(`
<p>Hi %s,</p>
<p>Thank you for your purchase from <strong>%s</strong>!</p>
<table style="border-collapse:collapse;margin:1rem 0">
  <tr><td style="padding:4px 12px 4px 0;color:#666">Order</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Product</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#666">Access Token</td><td><strong>%s</strong></td></tr>
</table>
<p>Download your product here:<br>
<a href="%s" style="color:#4f46e5;font-weight:bold">%s</a></p>
<p>This link is unique to your purchase. Do not share it.</p>
<p>Questions? Reply to this email.<br>— %s</p>
`, req.Name, storeName, req.PurchaseID, req.ProductID, req.AccessToken, downloadLink, downloadLink, storeName),
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

	drift.Log(fmt.Sprintf("[deliver] delivery email sent to %s (purchase %s)", req.Email, req.PurchaseID))
	return http.StatusOK, "Sent", map[string]string{"email": req.Email, "purchase_id": req.PurchaseID}
}
