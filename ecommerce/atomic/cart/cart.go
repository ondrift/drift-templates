// @atomic route=post:cart auth=none env=.env
// Manages the shopping cart stored in Backbone Cache under key cart:{session_id}.
// Supports actions: "add", "remove", "clear", "get"
//
// Body:
//   {
//     "session_id": "abc123",       // required for all actions
//     "action":     "add",          // add | remove | clear | get
//     "product_id": "product-slug", // required for add/remove
//     "quantity":   2               // required for add (default 1)
//   }
//
// Cart TTL: 7 days (604800 seconds)

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const cartTTL = 604800 // 7 days in seconds

type RequestBody struct {
	SessionID string `json:"session_id"`
	Action    string `json:"action"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func PostCart(req RequestBody) (int, string, interface{}) {
	sid := strings.TrimSpace(req.SessionID)
	if sid == "" {
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "session_id is required",
		}
	}
	if req.Quantity == 0 {
		req.Quantity = 1
	}

	backbone := os.Getenv("BACKBONE_URL")
	if backbone == "" {
		backbone = "http://backbone:8000"
	}

	cacheKey := fmt.Sprintf("cart:%s", sid)
	cart := loadCart(backbone, cacheKey)

	switch req.Action {
	case "get", "":
		return http.StatusOK, "OK", map[string]any{"cart": cart}

	case "add":
		if req.ProductID == "" {
			return http.StatusBadRequest, "Bad Request", map[string]string{"error": "product_id required for add"}
		}
		found := false
		for i, item := range cart {
			if item.ProductID == req.ProductID {
				cart[i].Quantity += req.Quantity
				found = true
				break
			}
		}
		if !found {
			cart = append(cart, CartItem{ProductID: req.ProductID, Quantity: req.Quantity})
		}

	case "remove":
		if req.ProductID == "" {
			return http.StatusBadRequest, "Bad Request", map[string]string{"error": "product_id required for remove"}
		}
		updated := []CartItem{}
		for _, item := range cart {
			if item.ProductID != req.ProductID {
				updated = append(updated, item)
			}
		}
		cart = updated

	case "clear":
		cart = []CartItem{}

	default:
		return http.StatusBadRequest, "Bad Request", map[string]string{
			"error": "action must be one of: get, add, remove, clear",
		}
	}

	saveCart(backbone, cacheKey, cart)
	return http.StatusOK, "OK", map[string]any{"cart": cart}
}

func loadCart(backbone, key string) []CartItem {
	resp, err := http.Get(fmt.Sprintf("%s/cache/get?key=%s", backbone, key))
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return []CartItem{}
	}
	defer resp.Body.Close()
	// Cache returns the raw value string (application/octet-stream or text/plain).
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return []CartItem{}
	}
	var items []CartItem
	if json.Unmarshal(raw, &items) != nil {
		return []CartItem{}
	}
	return items
}

func saveCart(backbone, key string, cart []CartItem) {
	// cache/set value must be a string, so marshal cart to JSON first.
	cartJSON, _ := json.Marshal(cart)
	payload, _ := json.Marshal(map[string]any{
		"key":   key,
		"value": string(cartJSON),
		"ttl":   cartTTL,
	})
	resp, _ := http.Post(fmt.Sprintf("%s/cache/set", backbone), "application/json", bytes.NewReader(payload))
	if resp != nil {
		resp.Body.Close()
	}
}
