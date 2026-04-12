// @atomic route=get:get-menu auth=none env=.env
// Returns the food truck menu.
// The menu is stored in Backbone Cache under key "truck-menu" as a JSON string.
// Seed it with:
//   drift backbone cache set truck-menu '{"items":[...]}'
// Update it the same way — the new value takes effect immediately.

package main

import (
	"encoding/json"
	"net/http"

	drift "github.com/ondrift/drift-sdk/go"
)

func GetGetMenu() (int, string, interface{}) {
	raw, err := drift.Cache.Get("truck-menu")
	if err != nil || len(raw) == 0 {
		return http.StatusOK, "OK", defaultMenu()
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return http.StatusOK, "OK", defaultMenu()
	}

	items, ok := doc["items"]
	if !ok {
		return http.StatusOK, "OK", defaultMenu()
	}
	return http.StatusOK, "OK", map[string]any{"items": items}
}

func defaultMenu() map[string]any {
	return map[string]any{
		"items": []map[string]any{
			{"category": "Mains", "name": "Classic Burger", "description": "Smashed beef patty, American cheese, pickles, special sauce", "price": 9.50},
			{"category": "Sides", "name": "Loaded Fries", "description": "Crispy fries, cheese sauce, jalapeños, bacon bits", "price": 6.50},
			{"category": "Mains", "name": "Chicken Tacos", "description": "Grilled chicken, slaw, chipotle mayo, corn tortillas (x2)", "price": 8.00},
			{"category": "Drinks", "name": "Lemonade", "description": "Fresh-squeezed, sweetened or unsweetened", "price": 3.50},
			{"category": "Desserts", "name": "Cookie", "description": "Chocolate chip, baked fresh daily", "price": 2.50},
		},
	}
}
