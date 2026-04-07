// @atomic route=post:checkout auth=none env=.env
// Creates an order from the current cart, clears the cart, and enqueues
// the order for confirmation email and optional fulfilment webhook.
//
// Body:
//
//	{
//	  "session_id": "abc123",
//	  "name":       "Jane Smith",
//	  "email":      "jane@example.com",
//	  "address":    "123 Main St, City, Country"
//	}
//
// NOTE: This template does not process payments directly.
// Integrate a payment gateway (e.g. Stripe) by verifying a payment_intent_id
// before calling this endpoint, or use the order-webhook function for
// post-payment processing.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	drift "github.com/ondrift/drift-sdk"
)

type RequestBody struct {
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Address   string `json:"address"`
}

type CartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func PostCheckout(req RequestBody) (int, string, interface{}) {
	sid := strings.TrimSpace(req.SessionID)
	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(strings.ToLower(req.Email))
	address := strings.TrimSpace(req.Address)

	if sid == "" || name == "" || email == "" || address == "" || !strings.Contains(email, "@") {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "session_id, name, email, and address are required",
		}
	}

	// ── Load cart ────────────────────────────────────────────────────────────
	cacheKey := "cart:" + sid
	cart := loadCart(cacheKey)
	if len(cart) == 0 {
		return http.StatusBadRequest, "Empty cart", map[string]string{
			"error": "Your cart is empty.",
		}
	}

	// ── Create order in NoSQL ────────────────────────────────────────────────
	orderID := fmt.Sprintf("order-%d", time.Now().UnixNano())
	doc := map[string]any{
		"id":         orderID,
		"session_id": sid,
		"name":       name,
		"email":      email,
		"address":    address,
		"items":      cart,
		"status":     "pending_payment",
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := drift.BackboneWrite("orders", doc); err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "Could not create order. Please try again.",
		}
	}

	// ── Clear cart from cache ─────────────────────────────────────────────────
	_ = drift.CacheDel(cacheKey)

	// ── Enqueue for confirmation email ────────────────────────────────────────
	_ = drift.QueuePush("order-queue", map[string]any{
		"order_id": orderID,
		"name":     name,
		"email":    email,
		"address":  address,
		"items":    cart,
	})

	return http.StatusOK, "Order created", map[string]any{
		"order_id": orderID,
		"message":  fmt.Sprintf("Order placed, %s! Check your email for confirmation.", name),
	}
}

func loadCart(key string) []CartItem {
	raw, err := drift.CacheGet(key)
	if err != nil || len(raw) == 0 {
		return []CartItem{}
	}
	var items []CartItem
	json.Unmarshal(raw, &items)
	return items
}
