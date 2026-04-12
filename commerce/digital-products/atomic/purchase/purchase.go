// @atomic route=post:purchase auth=none env=.env
// Creates a purchase record and enqueues delivery.
// NOTE: This template does not handle payment processing.
// Integrate Stripe/Paddle before this endpoint to verify payment
// before allowing the purchase to proceed.

package main

import (
	"fmt"
	"net/http"
	"time"

	drift "github.com/ondrift/drift-sdk/go"
)

type RequestBody struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}

func PostPurchase(req RequestBody) (int, string, interface{}) {
	if req.ProductID == "" || req.Name == "" || req.Email == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "product_id, name, and email are required",
		}
	}

	purchaseID := fmt.Sprintf("purchase-%d", time.Now().UnixNano())
	accessToken := fmt.Sprintf("tok-%d", time.Now().UnixNano())

	doc := map[string]any{
		"id":           purchaseID,
		"product_id":   req.ProductID,
		"name":         req.Name,
		"email":        req.Email,
		"access_token": accessToken,
		"purchased_at": time.Now().UTC().Format(time.RFC3339),
		"status":       "pending_delivery",
	}
	_, err := drift.NoSQL.Collection("purchases").Insert(doc)
	if err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save purchase",
		}
	}

	_ = drift.Queue("delivery-queue").Push(map[string]any{
		"purchase_id":  purchaseID,
		"product_id":   req.ProductID,
		"name":         req.Name,
		"email":        req.Email,
		"access_token": accessToken,
	})

	return http.StatusOK, "Purchase received", map[string]any{
		"purchase_id": purchaseID,
		"message":     fmt.Sprintf("Thanks %s! Your purchase is confirmed. You will receive a delivery email at %s shortly.", req.Name, req.Email),
	}
}
