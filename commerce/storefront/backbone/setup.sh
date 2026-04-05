#!/usr/bin/env bash
# ecommerce/backbone/setup.sh
#
# ⚠️  This template requires the Prototyper tier or higher.
#     Hacker tier limits (5 functions, 3 collections, 3 queues) are exceeded.
#     Upgrade your slice before deploying:
#         drift slice upgrade --tier prototyper

set -euo pipefail

echo "Setting up Backbone resources for ecommerce..."

# ── NoSQL collections ────────────────────────────────────────────────────────
drift backbone nosql write \
  --collection orders \
  --data '{"_setup":true}'
echo "✅ orders collection ready"

# ── Product catalogue in cache (no TTL = persists until manually updated) ───
echo "Seeding sample products into cache..."

drift backbone cache set products:product-tshirt \
  '{"id":"product-tshirt","name":"Classic T-Shirt","description":"100% organic cotton, available in 6 colours.","price":29.99,"category":"Apparel","image":"","in_stock":true}'

drift backbone cache set products:product-hoodie \
  '{"id":"product-hoodie","name":"Heavyweight Hoodie","description":"Brushed fleece interior, kangaroo pocket, unisex fit.","price":64.99,"category":"Apparel","image":"","in_stock":true}'

drift backbone cache set products:product-mug \
  '{"id":"product-mug","name":"Ceramic Mug","description":"330 ml, dishwasher safe, available in white and black.","price":14.99,"category":"Accessories","image":"","in_stock":true}'

# Catalogue index — tells get-products which product IDs to fetch.
drift backbone cache set products:catalogue \
  '{"product_ids":["product-tshirt","product-hoodie","product-mug"]}'

echo "✅ product catalogue seeded in cache"

# ── Order queue (auto-created on first push) ─────────────────────────────────
drift backbone queue push order-queue '{"_setup":true}'
echo "✅ order-queue ready"

echo ""
echo "Deploy your functions:"
echo "    drift atomic deploy atomic/get-products"
echo "    drift atomic deploy atomic/cart"
echo "    drift atomic deploy atomic/checkout"
echo "    drift atomic deploy atomic/order-confirm"
echo ""
echo "Set secrets for order confirmation emails:"
echo "    drift backbone secret set RESEND_API_KEY=<your-resend-key>"
echo "    drift backbone secret set SENDER_EMAIL=orders@yourstore.com"
echo "    drift backbone secret set STORE_NAME=\"Your Store Name\""
echo ""
echo "The order-confirm function registers a queue trigger on deploy."
echo "To register it manually:"
echo "    drift atomic trigger register queue order-confirm \\"
echo "        --queue order-queue \\"
echo "        --target /api/order-confirm"
echo ""
echo "To add product images, set the 'image' field to a public URL:"
echo "    drift backbone cache set products:product-tshirt '{\"id\":\"product-tshirt\",...,\"image\":\"https://...\"}'"
