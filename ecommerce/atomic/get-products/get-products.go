// @atomic route=get:get-products auth=none env=.env
// Returns the full product catalogue from Backbone Cache.
// Products and catalogue index are stored as JSON strings in cache.
// Seed them with backbone/setup.sh.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func GetGetProducts() (int, string, interface{}) {
	backbone := os.Getenv("BACKBONE_URL")
	if backbone == "" {
		backbone = "http://backbone:8000"
	}

	// ── Load catalogue index ──────────────────────────────────────────────────
	idxResp, err := http.Get(fmt.Sprintf("%s/cache/get?key=products:catalogue", backbone))
	if err != nil || idxResp.StatusCode != http.StatusOK {
		if idxResp != nil {
			idxResp.Body.Close()
		}
		return http.StatusOK, "OK", map[string]any{
			"items": []any{},
			"note":  "Run backbone/setup.sh to seed the product catalogue",
		}
	}
	defer idxResp.Body.Close()

	idxRaw, err := io.ReadAll(idxResp.Body)
	if err != nil {
		return http.StatusInternalServerError, "Read error", map[string]string{"error": "could not read catalogue"}
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
		pResp, err := http.Get(fmt.Sprintf("%s/cache/get?key=products:%s", backbone, id))
		if err != nil || pResp.StatusCode != http.StatusOK {
			if pResp != nil {
				pResp.Body.Close()
			}
			continue
		}
		pRaw, _ := io.ReadAll(pResp.Body)
		pResp.Body.Close()

		var product map[string]any
		if json.Unmarshal(pRaw, &product) == nil {
			products = append(products, product)
		}
	}

	return http.StatusOK, "OK", map[string]any{"items": products}
}
