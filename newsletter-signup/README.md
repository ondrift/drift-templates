# Newsletter Signup — Drift Template

**Tier:** Hacker (free)

A simple email newsletter signup page with double-send prevention, unsubscribe support, and automated welcome emails.

## What you get

| Component | Description |
|-----------|-------------|
| `canvas/index.html` | Dark-themed signup page with inline unsubscribe form |
| `atomic/subscribe` | Validates email, deduplicates via cache, saves to NoSQL, enqueues welcome email |
| `atomic/unsubscribe` | Marks email as unsubscribed in NoSQL, evicts cache entry |
| `atomic/send-welcome` | Queue trigger — sends a welcome email via Resend |

## Prerequisites

- A [Resend](https://resend.com) account with a verified sender domain
- A Drift slice on the Hacker tier or higher

## Resource usage

| Resource | Used | Hacker limit |
|----------|------|-------------|
| Functions | 3 | 5 |
| NoSQL collections | 1 (`subscribers`) | 3 |
| Queues | 1 (`signup-queue`) | 3 |
| Canvas sites | 1 | 1 |

## Deploy

```bash
# 1. Set up Backbone (run once)
bash backbone/setup.sh

# 2. Set secrets
drift backbone secret set RESEND_API_KEY  <your-resend-api-key>
drift backbone secret set SENDER_EMAIL    hello@yourdomain.com

# 3. Deploy functions
drift atomic deploy atomic/subscribe
drift atomic deploy atomic/unsubscribe
drift atomic deploy atomic/send-welcome

# 4. Deploy the site
drift canvas deploy canvas/
```

## Customise

- Edit the hero copy and colour scheme in `canvas/index.html`
- Edit the welcome email HTML in `atomic/send-welcome/send-welcome.go`
- The `signup-queue` trigger is auto-registered on deploy of `send-welcome`
