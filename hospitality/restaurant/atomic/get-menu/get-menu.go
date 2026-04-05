// @atomic route=get:get-menu auth=none env=.env
// Returns the restaurant menu.
// The menu is stored in Backbone Cache under key "menu" as a JSON string.
// Seed it with:
//   drift backbone cache set menu '{"items":[...]}'
// Update it the same way — the new value takes effect immediately.

package main

import (
	"encoding/json"
	"net/http"

	"drift-sdk"
)

func GetGetMenu() (int, string, interface{}) {
	raw, err := drift.CacheGet("menu")
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
			{"category": "Starters", "name": "Bruschetta al Pomodoro", "description": "Grilled sourdough, heritage tomatoes, fresh basil, aged balsamic", "price": 7.50, "image_url": "https://picsum.photos/seed/bruschetta/600/400"},
			{"category": "Starters", "name": "Burrata con Prosciutto", "description": "Creamy burrata, San Daniele prosciutto, fig compote, toasted walnuts", "price": 13.00, "image_url": "https://picsum.photos/seed/burrata/600/400"},
			{"category": "Starters", "name": "Zuppa del Giorno", "description": "Chef's soup of the day — ask your server", "price": 6.50, "image_url": "https://picsum.photos/seed/italiansoupe/600/400"},
			{"category": "Mains", "name": "Tagliatelle al Ragù", "description": "Slow-braised beef ragù, handmade egg tagliatelle, Parmigiano Reggiano", "price": 18.00, "image_url": "https://picsum.photos/seed/tagliatelle/600/400"},
			{"category": "Mains", "name": "Risotto ai Funghi", "description": "Carnaroli rice, wild porcini, black truffle shavings, aged butter", "price": 21.00, "image_url": "https://picsum.photos/seed/risotto/600/400"},
			{"category": "Mains", "name": "Branzino al Forno", "description": "Whole roasted sea bass, caperberries, olives, cherry tomatoes, white wine", "price": 27.00, "image_url": "https://picsum.photos/seed/seabass/600/400"},
			{"category": "Mains", "name": "Margherita al Forno", "description": "San Marzano tomato, buffalo mozzarella, fresh basil, stone-baked", "price": 15.00, "image_url": "https://picsum.photos/seed/pizzamargherita/600/400"},
			{"category": "Desserts", "name": "Tiramisù della Casa", "description": "House recipe — mascarpone, Savoiardi, espresso, cocoa", "price": 8.00, "image_url": "https://picsum.photos/seed/tiramisu/600/400"},
			{"category": "Desserts", "name": "Panna Cotta", "description": "Vanilla panna cotta, warm mixed berry coulis", "price": 7.00, "image_url": "https://picsum.photos/seed/pannacotta/600/400"},
			{"category": "Desserts", "name": "Gelato Artigianale", "description": "Three scoops of house-made gelato — ask your server for today's flavours", "price": 6.50, "image_url": "https://picsum.photos/seed/gelato/600/400"},
		},
	}
}
