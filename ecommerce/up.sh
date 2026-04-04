#!/usr/bin/env bash
# ecommerce/up.sh — Deploy the e-commerce template to your Drift slice.
#
# ⚠️  Requires the Prototyper tier:
#     drift slice upgrade --tier prototyper

set -euo pipefail
cd "$(dirname "$0")"

echo "🛒  Deploying ecommerce template..."
echo ""

echo "▶  Setting up Backbone resources..."
bash backbone/setup.sh
echo ""

echo "▶  Deploying atomic functions..."
drift atomic deploy atomic/get-products
drift atomic deploy atomic/cart
drift atomic deploy atomic/checkout
drift atomic deploy atomic/order-confirm
echo ""

echo "▶  Deploying canvas..."
drift canvas deploy canvas/
echo ""

echo "✅ E-commerce template deployed!"
echo ""
echo "   To enable order confirmation emails, set:"
echo "     drift backbone secret set RESEND_API_KEY=re_..."
echo "     drift backbone secret set SENDER_EMAIL=orders@yourstore.com"
echo "     drift backbone secret set STORE_NAME=\"Your Store Name\""
echo ""
echo "   To add or update products:"
echo "     drift backbone cache set products:<id> '{\"id\":\"<id>\",\"name\":\"...\",\"price\":9.99,...}'"
echo "     drift backbone cache set products:catalogue '{\"product_ids\":[\"<id>\",...]}'"
