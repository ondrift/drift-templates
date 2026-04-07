// @atomic route=post:cart auth=none env=.env
// Manages the shopping cart stored in Backbone Cache under key cart:{session_id}.
// Supports actions: "add", "remove", "clear", "get"
//
// Body:
//
//	{
//	  "session_id": "abc123",       // required for all actions
//	  "action":     "add",          // add | remove | clear | get
//	  "product_id": "product-slug", // required for add/remove
//	  "quantity":   2               // required for add (default 1)
//	}
//
// Cart TTL: 7 days (604800 seconds)

package main

import (
	"encoding/json"
	"net/http"
	"strings"

	drift "github.com/ondrift/drift-sdk"
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

	cacheKey := "cart:" + sid
	cart := loadCart(cacheKey)

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

	saveCart(cacheKey, cart)
	return http.StatusOK, "OK", map[string]any{"cart": cart}
}

func loadCart(key string) []CartItem {
	raw, err := drift.CacheGet(key)
	if err != nil || len(raw) == 0 {
		return []CartItem{}
	}
	var items []CartItem
	if json.Unmarshal(raw, &items) != nil {
		return []CartItem{}
	}
	return items
}

func saveCart(key string, cart []CartItem) {
	cartJSON, _ := json.Marshal(cart)
	drift.CacheSet(key, string(cartJSON), cartTTL)
}
