// @atomic route=post:pre-order auth=none env=.env

package main

import (
	"fmt"
	"net/http"
	"time"

	drift "github.com/ondrift/drift-sdk/go"
)

type OrderItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type RequestBody struct {
	Name       string      `json:"name"`
	Email      string      `json:"email"`
	Items      []OrderItem `json:"items"`
	PickupTime string      `json:"pickup_time"` // e.g. "12:30"
}

func PostPreOrder(req RequestBody) (int, string, interface{}) {
	if req.Name == "" || req.Email == "" || req.PickupTime == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "name, email, and pickup_time are required",
		}
	}
	if len(req.Items) == 0 {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "at least one item is required",
		}
	}
	for _, item := range req.Items {
		if item.Name == "" || item.Quantity < 1 {
			return http.StatusBadRequest, "Bad Request", map[string]string{
				"error": "each item must have a name and quantity >= 1",
			}
		}
	}

	orderID := fmt.Sprintf("order-%d", time.Now().UnixNano())

	doc := map[string]any{
		"order_id":    orderID,
		"name":        req.Name,
		"email":       req.Email,
		"items":       req.Items,
		"pickup_time": req.PickupTime,
		"status":      "pending",
		"created_at":  time.Now().UTC().Format(time.RFC3339),
	}
	_, err := drift.NoSQL.Collection("orders").Insert(doc)
	if err != nil {
		return http.StatusInternalServerError, "Storage error", map[string]string{
			"error": "failed to save order",
		}
	}

	_ = drift.Queue("order-queue").Push(map[string]any{
		"order_id":    orderID,
		"name":        req.Name,
		"email":       req.Email,
		"items":       req.Items,
		"pickup_time": req.PickupTime,
	})

	return http.StatusOK, "Order received", map[string]any{
		"order_id": orderID,
		"message":  fmt.Sprintf("Thanks %s! Your order %s will be ready for pickup at %s.", req.Name, orderID, req.PickupTime),
	}
}
