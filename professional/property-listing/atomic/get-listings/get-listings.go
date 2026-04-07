// @atomic route=get:get-listings auth=none env=.env
// Returns the property listings.
// Listings are stored in Backbone Cache under key "listings" as a JSON string.
// Seed it with:
//   drift backbone cache set listings '{"listings":[...]}'
// Update it the same way — the new value takes effect immediately.

package main

import (
	"encoding/json"
	"net/http"

	drift "github.com/ondrift/drift-sdk"
)

func GetGetListings() (int, string, interface{}) {
	raw, err := drift.CacheGet("listings")
	if err != nil || len(raw) == 0 {
		return http.StatusOK, "OK", defaultListings()
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return http.StatusOK, "OK", defaultListings()
	}

	listings, ok := doc["listings"]
	if !ok {
		return http.StatusOK, "OK", defaultListings()
	}
	return http.StatusOK, "OK", map[string]any{"listings": listings}
}

func defaultListings() map[string]any {
	return map[string]any{
		"listings": []map[string]any{
			{"id": "prop-oak-lane", "address": "42 Oak Lane, Riverside", "price": 485000, "bedrooms": 3, "bathrooms": 2, "sqft": 1850, "description": "Charming craftsman home with original hardwood floors, updated kitchen, and a landscaped backyard with mature trees.", "image_url": "https://picsum.photos/seed/house1/800/500", "status": "active", "open_house_date": "2026-04-12 14:00-16:00"},
			{"id": "prop-maple-drive", "address": "118 Maple Drive, Hillcrest", "price": 725000, "bedrooms": 4, "bathrooms": 3, "sqft": 2600, "description": "Spacious modern home with open floor plan, chef's kitchen, home office, and a two-car garage. Mountain views from the deck.", "image_url": "https://picsum.photos/seed/house2/800/500", "status": "active", "open_house_date": "2026-04-13 11:00-13:00"},
			{"id": "prop-elm-court", "address": "7 Elm Court, Downtown", "price": 340000, "bedrooms": 2, "bathrooms": 1, "sqft": 1100, "description": "Updated downtown condo with exposed brick, in-unit laundry, and rooftop access. Walking distance to restaurants and transit.", "image_url": "https://picsum.photos/seed/condo1/800/500", "status": "active", "open_house_date": ""},
			{"id": "prop-cedar-ridge", "address": "201 Cedar Ridge Road", "price": 890000, "bedrooms": 5, "bathrooms": 4, "sqft": 3400, "description": "Executive home on a half-acre lot. Pool, three-car garage, gourmet kitchen, and a finished basement with home theatre.", "image_url": "https://picsum.photos/seed/house3/800/500", "status": "pending", "open_house_date": ""},
		},
	}
}
