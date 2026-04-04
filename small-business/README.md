# Small Business Landing Page — Drift Template

**Tier:** Hacker (free)

A professional, fully responsive landing page for a small business or agency. Includes a services section, about section, static blog/updates, and a contact form that notifies the owner by email.

## What you get

| Component | Description |
|-----------|-------------|
| `canvas/index.html` | Full landing page: sticky nav, hero, services grid, about + stats, blog section, contact form, footer |
| `atomic/contact` | Validates and saves contact form submissions to NoSQL, enqueues notification |
| `atomic/notify-contact` | Queue trigger — forwards lead details to the owner's inbox via Resend |

## Prerequisites

- A [Resend](https://resend.com) account with a verified sender domain (for notifications)
- A Drift slice on the Hacker tier or higher

## Resource usage

| Resource | Used | Hacker limit |
|----------|------|-------------|
| Functions | 2 | 5 |
| NoSQL collections | 1 (`leads`) | 3 |
| Queues | 1 (`contact-queue`) | 3 |
| Canvas sites | 1 | 1 |

## Deploy

```bash
# 1. Set up Backbone (run once)
bash backbone/setup.sh

# 2. Set secrets (required for email notifications)
drift backbone secret set RESEND_API_KEY  <your-resend-api-key>
drift backbone secret set SENDER_EMAIL    notifications@yourdomain.com
drift backbone secret set OWNER_EMAIL     you@yourdomain.com

# 3. Deploy functions
drift atomic deploy atomic/contact
drift atomic deploy atomic/notify-contact

# 4. Deploy the site
drift canvas deploy canvas/
```

## Customise

- Replace "Apex Studio" branding throughout `canvas/index.html`
- Update the services grid with your actual offerings
- Update the about section stats and copy
- Replace the blog section with your real articles (or remove it)
- Update contact details (email, phone, location) in the Contact section
- The contact form works without email secrets — submissions are always saved to the `leads` NoSQL collection
