# Booking System — Drift Template

**Tier:** Hacker (free)

A multi-step appointment booking flow with real-time slot availability, cancellation support, and confirmation emails.

## What you get

| Component | Description |
|-----------|-------------|
| `canvas/index.html` | 4-step booking wizard: service → slot → details → confirm |
| `atomic/get-slots` | Returns available time slots for a date (filters out cache-locked ones) |
| `atomic/book-slot` | Locks the slot in cache, saves booking to NoSQL, enqueues confirmation |
| `atomic/cancel-booking` | Validates ownership, marks cancelled, releases slot lock |
| `atomic/confirm-booking` | Queue trigger — sends HTML booking confirmation via Resend |

## Prerequisites

- A [Resend](https://resend.com) account with a verified sender domain
- A Drift slice on the Hacker tier or higher

## Resource usage

| Resource | Used | Hacker limit |
|----------|------|-------------|
| Functions | 4 | 5 |
| NoSQL collections | 1 (`bookings`) | 3 |
| Queues | 1 (`booking-queue`) | 3 |
| Canvas sites | 1 | 1 |

## Deploy

```bash
# 1. Set up Backbone (run once)
bash backbone/setup.sh

# 2. Set secrets
drift backbone secret set RESEND_API_KEY  <your-resend-api-key>
drift backbone secret set SENDER_EMAIL    bookings@yourbusiness.com
drift backbone secret set BUSINESS_NAME   "Your Business Name"

# 3. Deploy functions
drift atomic deploy atomic/get-slots
drift atomic deploy atomic/book-slot
drift atomic deploy atomic/cancel-booking
drift atomic deploy atomic/confirm-booking

# 4. Deploy the site
drift canvas deploy canvas/
```

## Customise

- Edit the available time slots in `atomic/get-slots/get-slots.go` (`allSlots` variable)
- Edit the service list in `canvas/index.html` (`<select id="svc-select">`)
- Slot locks expire after 48 hours by default; change `cartTTLms` in `book-slot.go`
- Edit the confirmation email in `atomic/confirm-booking/confirm-booking.go`

## Cancellations

Customers need their booking ID (returned at booking time and sent in the confirmation
email) and the email address used to book. The cancel endpoint is at `POST /api/cancel-booking`.
