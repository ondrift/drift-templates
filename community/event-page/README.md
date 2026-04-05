# Event Page — Drift Template

**Tier:** Hacker (free)

A slick event landing page with a live RSVP counter, session schedule, and RSVP form.

## What you get

| Component | Description |
|-----------|-------------|
| `canvas/index.html` | Dark-themed event page with live counter, schedule, RSVP form, map embed |
| `atomic/rsvp` | Validates and saves RSVPs to NoSQL |
| `atomic/get-rsvps` | API-key protected — returns RSVP and guest counts from cache |

## Prerequisites

- A Drift slice on the Hacker tier or higher

## Resource usage

| Resource | Used | Hacker limit |
|----------|------|-------------|
| Functions | 2 | 5 |
| NoSQL collections | 1 (`rsvps`) | 3 |
| Cache keys | 1 (`rsvp:stats`) | unlimited |
| Canvas sites | 1 | 1 |

## Deploy

```bash
# 1. Set up Backbone (run once)
bash backbone/setup.sh

# 2. Deploy functions
drift atomic deploy atomic/rsvp
drift atomic deploy atomic/get-rsvps

# 3. Protect the get-rsvps function with an API key
drift atomic auth set --function get-rsvps --method GET --key <your-secret-key>

# 4. Deploy the site
drift canvas deploy canvas/
```

## Keeping the counter up to date

The `rsvp:stats` cache key is seeded at `{"rsvp_count": 0, "guest_count": 0}`.
Update it after each RSVP using a scheduled function or by querying the `rsvps`
collection and writing updated totals:

```bash
drift backbone cache set rsvp:stats '{"rsvp_count":42,"guest_count":67}'
```

## Customise

- Update event name, date, time, and location in `canvas/index.html`
- Swap the Google Maps embed `src` for your venue URL
- Edit the session schedule in `canvas/index.html`
- Change the max guests per RSVP in `atomic/rsvp/rsvp.go`
