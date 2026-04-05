// @atomic route=post:order-confirm auth=none env=.env
// drift:trigger queue order-queue poll=3000ms retry=3
//
// Consumes the order-queue and sends an order confirmation email via Resend.
// Required env vars (set as Backbone Secrets):
//   RESEND_API_KEY  - Resend API key
//   SENDER_EMAIL    - verified sender, e.g. orders@yourstore.com
//   STORE_NAME      - displayed in the email (default: "Our store")

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"drift-sdk"
)

type CartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type RequestBody struct {
	OrderID string     `json:"order_id"`
	Name    string     `json:"name"`
	Email   string     `json:"email"`
	Address string     `json:"address"`
	Items   []CartItem `json:"items"`
}

func PostOrderConfirm(req RequestBody) (int, string, interface{}) {
	apiKey := drift.Env("RESEND_API_KEY")
	sender := drift.Env("SENDER_EMAIL")
	store := drift.Env("STORE_NAME")
	if store == "" {
		store = "Our store"
	}

	if apiKey == "" || sender == "" {
		return http.StatusOK, "No email configured", map[string]string{
			"note": "Set RESEND_API_KEY and SENDER_EMAIL secrets to enable order confirmations",
		}
	}

	// Build item rows
	rows := ""
	for _, item := range req.Items {
		rows += fmt.Sprintf(
			`<tr><td style="padding:6px 12px">%s</td><td style="padding:6px 12px;text-align:right">×%d</td></tr>`,
			item.ProductID, item.Quantity,
		)
	}

	html := fmt.Sprintf(`
<p>Hi %s,</p>
<p>Thanks for your order from <strong>%s</strong>! Here's a summary:</p>
<table style="border-collapse:collapse;width:100%%;max-width:520px;margin:1rem 0">
  <thead>
    <tr style="background:#f9fafb">
      <th style="padding:8px 12px;text-align:left;font-size:0.8rem;text-transform:uppercase;letter-spacing:0.05em">Product</th>
      <th style="padding:8px 12px;text-align:right;font-size:0.8rem;text-transform:uppercase;letter-spacing:0.05em">Qty</th>
    </tr>
  </thead>
  <tbody>%s</tbody>
</table>
<p><strong>Shipping to:</strong><br>%s</p>
<p style="color:#6b7280;font-size:0.9em">Order ID: %s</p>
<p>We'll send another email when your order ships.</p>
<p>— %s</p>`,
		req.Name, store,
		rows,
		strings.ReplaceAll(req.Address, "\n", "<br>"),
		req.OrderID,
		store,
	)

	body, _ := json.Marshal(map[string]any{
		"from":    sender,
		"to":      []string{req.Email},
		"subject": fmt.Sprintf("Order confirmed — %s", req.OrderID),
		"html":    html,
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
			"error": "Failed to send order confirmation",
		}
	}
	if resp.Status >= 400 {
		return http.StatusInternalServerError, "Email error", map[string]string{
			"error": fmt.Sprintf("resend returned %d", resp.Status),
		}
	}

	drift.Log(fmt.Sprintf("[order-confirm] confirmation sent to %s for %s", req.Email, req.OrderID))
	return http.StatusOK, "Email sent", map[string]string{
		"message": fmt.Sprintf("Confirmation sent to %s", req.Email),
	}
}
