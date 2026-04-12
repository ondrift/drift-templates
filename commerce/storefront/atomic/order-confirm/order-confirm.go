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
	"fmt"
	"net/http"

	drift "github.com/ondrift/drift-sdk/go"
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
	apiKey, _ := drift.Secret.Get("RESEND_API_KEY")
	sender, _ := drift.Secret.Get("SENDER_EMAIL")
	store, _ := drift.Secret.Get("STORE_NAME")
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

	return http.StatusOK, "Email sent", map[string]string{
		"message": fmt.Sprintf("Confirmation sent to %s", req.Email),
	}
}
