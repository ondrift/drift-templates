# E-commerce — Drift Template

**Tier:** Prototyper (upgrade required)

A lightweight online shop with a product catalogue, session-based cart, and order checkout flow. Confirmation emails are sent via Resend when orders are placed.

> **⚠️ Requires Prototyper tier.**
> This template uses 4 functions and 2 NoSQL collections, which exceeds the Hacker tier's limits.
> Upgrade your slice before deploying: `drift slice upgrade --tier prototyper`

## What you get

| Component | Description |
|-----------|-------------|
| `canvas/index.html` | Product grid, cart drawer, checkout modal with localStorage session |
| `atomic/get-products` | Reads product catalogue index and fetches each product from NoSQL |
| `atomic/cart` | Manages cart in Backbone Cache (TTL 7 days); supports add/remove/clear/get |
| `atomic/checkout` | Creates order in NoSQL, clears cart, enqueues confirmation |
| `atomic/order-confirm` | Queue trigger — sends HTML order confirmation email via Resend |

## Prerequisites

- A [Resend](https://resend.com) account with a verified sender domain
- A Drift slice on the **Prototyper tier** or higher

## Resource usage

| Resource | Used | Prototyper limit |
|----------|------|-----------------|
| Functions | 4 | 20 |
| NoSQL collections | 2 (`products`, `orders`) | 15 |
| Queues | 1 (`order-queue`) | 10 |
| Canvas sites | 1 | 5 |

## Deploy

```bash
# 1. Upgrade slice to Prototyper
drift slice upgrade --tier prototyper

# 2. Set up Backbone — seeds 3 sample products (run once)
bash backbone/setup.sh

# 3. Set secrets
drift backbone secret set RESEND_API_KEY  <your-resend-api-key>
drift backbone secret set SENDER_EMAIL    orders@yourstore.com
drift backbone secret set STORE_NAME      "Your Store Name"

# 4. Deploy functions
drift atomic deploy atomic/get-products
drift atomic deploy atomic/cart
drift atomic deploy atomic/checkout
drift atomic deploy atomic/order-confirm

# 5. Deploy the site
drift canvas deploy canvas/
```

## Adding products

Write product documents directly to the `products` collection, then update the catalogue index:

```bash
drift backbone nosql write --collection products --id product-new \
  --data '{"id":"product-new","name":"New Item","description":"...","price":19.99,"category":"Accessories","in_stock":true}'

# Update catalogue index to include the new product
drift backbone nosql write --collection products --id catalogue \
  --data '{"id":"catalogue","product_ids":["product-tshirt","product-hoodie","product-mug","product-new"]}'
```

## Product images

Upload images to Backbone Blobs and set the `image` field on each product:

```bash
drift backbone blobs upload product-photo.jpg
# Returns a blob URL — set it on the product document
drift backbone nosql write --collection products --id product-tshirt \
  --data '{"id":"product-tshirt",...,"image":"<blob-url>"}'
```

## Payment integration

This template does **not** process payments. To accept payments:

1. Integrate [Stripe](https://stripe.com) — verify a `payment_intent_id` before calling `/api/checkout`
2. Or use Stripe webhooks: update order status in NoSQL after `payment_intent.succeeded`
3. The `order-confirm` function sends confirmation regardless — gate it on `status=paid` for production use
