// @atomic route=get:get-products auth=none env=.env
// Returns the full product catalogue from Backbone Cache.
// Products and catalogue index are stored as JSON strings in cache.
// Seed them via drift.yaml cache entries.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	drift "github.com/ondrift/drift-sdk"
)

func GetGetProducts() (int, string, interface{}) {
	// ── Load catalogue index ──────────────────────────────────────────────────
	idxRaw, err := drift.CacheGet("products:catalogue")
	if err != nil || len(idxRaw) == 0 {
		return http.StatusOK, "OK", map[string]any{
			"items": []any{},
			"note":  "Product catalogue not yet seeded",
		}
	}

	var catalogue map[string]any
	if err := json.Unmarshal(idxRaw, &catalogue); err != nil {
		return http.StatusInternalServerError, "Parse error", map[string]string{"error": "could not parse catalogue"}
	}

	ids, _ := catalogue["product_ids"].([]any)
	products := []any{}

	for _, rawID := range ids {
		id, ok := rawID.(string)
		if !ok {
			continue
		}
		pRaw, err := drift.CacheGet(fmt.Sprintf("products:%s", id))
		if err != nil || len(pRaw) == 0 {
			continue
		}
		var product map[string]any
		if json.Unmarshal(pRaw, &product) == nil {
			products = append(products, product)
		}
	}

	return http.StatusOK, "OK", map[string]any{"items": products}
}
