// @atomic route=post:notify-order auth=none env=.env
// drift:trigger queue order-queue poll=3000ms retry=3

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	drift "github.com/ondrift/drift-sdk"
)

type OrderItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

// RequestBody is the message popped from order-queue by the trigger.
type RequestBody struct {
	OrderID    string      `json:"order_id"`
	Name       string      `json:"name"`
	Email      string      `json:"email"`
	Items      []OrderItem `json:"items"`
	PickupTime string      `json:"pickup_time"`
}

func PostNotifyOrder(req RequestBody) (int, string, interface{}) {
	if req.Email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{"error": "email required"}
	}

	apiKey := drift.Env("RESEND_API_KEY")
	if apiKey == "" {
		drift.Log(fmt.Sprintf("[notify-order] no RESEND_API_KEY — skipping email for %s", req.Email))
		return http.StatusOK, "Skipped", map[string]string{"reason": "no email key configured"}
	}

	senderEmail := drift.Env("SENDER_EMAIL")
	if senderEmail == "" {
		senderEmail = "orders@yourdomain.com"
	}
	truckName := drift.Env("TRUCK_NAME")
	if truckName == "" {
		truckName = "My Food Truck"
	}

	// Build items table rows.
	var rows strings.Builder
	for _, item := range req.Items {
		rows.WriteString(fmt.Sprintf(
			`<tr><td style="padding:6px 16px 6px 0;border-bottom:1px solid #eee">%s</td><td style="padding:6px 16px;border-bottom:1px solid #eee;text-align:center">x%d</td></tr>`,
			item.Name, item.Quantity,
		))
	}

	payload := map[string]any{
		"from":    fmt.Sprintf("%s <%s>", truckName, senderEmail),
		"to":      []string{req.Email},
		"subject": fmt.Sprintf("Order confirmed — %s (pickup %s)", req.OrderID, req.PickupTime),
		"html": fmt.Sprintf(`
<p>Hi %s,</p>
<p>Your order at <strong>%s</strong> has been received!</p>
<table style="border-collapse:collapse;margin:1rem 0;width:100%%;max-width:400px">
  <tr style="background:#f97316;color:#fff">
    <th style="padding:8px 16px 8px 0;text-align:left">Item</th>
    <th style="padding:8px 16px;text-align:center">Qty</th>
  </tr>
  %s
</table>
<p><strong>Pickup time:</strong> %s</p>
<p><strong>Order ID:</strong> %s</p>
<p>We'll have it hot and ready for you!</p>
<p>— %s</p>
`, req.Name, truckName, rows.String(), req.PickupTime, req.OrderID, truckName),
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

	drift.Log(fmt.Sprintf("[notify-order] confirmation sent to %s (order %s)", req.Email, req.OrderID))
	return http.StatusOK, "Sent", map[string]string{"email": req.Email, "order_id": req.OrderID}
}
