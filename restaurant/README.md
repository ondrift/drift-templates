# Restaurant Website — Drift Template

**Tier:** Hacker (free)

A full restaurant website with a dynamic menu, online reservation form, and automated confirmation emails.

## What you get

| Component | Description |
|-----------|-------------|
| `canvas/index.html` | Italian restaurant page with nav, hero, menu grid, reservation form, map embed |
| `atomic/get-menu` | Returns menu items from NoSQL; falls back to a hardcoded sample if not seeded |
| `atomic/submit-reservation` | Validates booking details, generates a confirmation code, saves to NoSQL, enqueues email |
| `atomic/confirm-reservation` | Queue trigger — sends HTML confirmation email via Resend |

## Prerequisites

- A [Resend](https://resend.com) account with a verified sender domain
- A Drift slice on the Hacker tier or higher

## Resource usage

| Resource | Used | Hacker limit |
|----------|------|-------------|
| Functions | 3 | 5 |
| NoSQL collections | 2 (`menu`, `reservations`) | 3 |
| Queues | 1 (`reservation-queue`) | 3 |
| Canvas sites | 1 | 1 |

## Deploy

```bash
# 1. Set up Backbone — seeds menu with 7 sample dishes (run once)
bash backbone/setup.sh

# 2. Set secrets
drift backbone secret set RESEND_API_KEY   <your-resend-api-key>
drift backbone secret set SENDER_EMAIL     reservations@yourrestaurant.com
drift backbone secret set RESTAURANT_NAME  "La Cucina"

# 3. Deploy functions
drift atomic deploy atomic/get-menu
drift atomic deploy atomic/submit-reservation
drift atomic deploy atomic/confirm-reservation

# 4. Deploy the site
drift canvas deploy canvas/
```

## Customise

- Update restaurant name, address, and contact details in `canvas/index.html`
- Swap the Google Maps embed `src` for your actual venue URL
- Update the menu by writing new documents to the `menu` collection:
  ```bash
  drift backbone nosql write --collection menu \
    --data '{"key":"menu","items":[...]}'
  ```
- Adjust opening hours and available time slots in `canvas/index.html`
- Edit the confirmation email template in `atomic/confirm-reservation/confirm-reservation.go`
