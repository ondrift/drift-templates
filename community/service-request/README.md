# Service Request Portal — Drift Template

**Tier:** Hacker (free)

A civic service request portal for reporting city issues like potholes, broken streetlights, graffiti, and more. Includes a public submission form, status tracker, and automated department notifications.

## What you get

| Component | Description |
|-----------|-------------|
| `canvas/index.html` | Government-styled page with category cards, report form, and ticket status checker |
| `atomic/submit-request` | Validates issue details, generates a ticket number, saves to NoSQL, enqueues notification |
| `atomic/check-status` | Looks up a ticket by number from cache and returns current status |
| `atomic/notify-department` | Queue trigger — sends service request notification email to the department via Resend |

## Prerequisites

- A [Resend](https://resend.com) account with a verified sender domain
- A Drift slice on the Hacker tier or higher

## Resource usage

| Resource | Used | Hacker limit |
|----------|------|-------------|
| Functions | 3 | 5 |
| NoSQL collections | 1 (`requests`) | 3 |
| Queues | 1 (`department-queue`) | 3 |
| Canvas sites | 1 | 1 |

## Deploy

```bash
# 1. Set secrets
drift backbone secret set RESEND_API_KEY   <your-resend-api-key>
drift backbone secret set SENDER_EMAIL     noreply@yourcity.gov
drift backbone secret set DEPARTMENT_EMAIL publicworks@yourcity.gov
drift backbone secret set CITY_NAME        "My City"

# 2. Deploy functions
drift atomic deploy atomic/submit-request
drift atomic deploy atomic/check-status
drift atomic deploy atomic/notify-department

# 3. Deploy the site
drift canvas deploy canvas/
```

## Customise

- Update city name and branding in `canvas/index.html`
- Change the `CITY_NAME` secret to update email sender name
- Edit categories in both `canvas/index.html` and `atomic/submit-request/submit-request.go`
- Modify the notification email template in `atomic/notify-department/notify-department.go`
- Add additional status values (e.g. `in_progress`, `resolved`) by updating the cache entry
