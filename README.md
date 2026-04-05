# drift-templates

Starter templates for building on Drift. Each template is a complete, deployable application using the three Drift service layers: **atomic** (serverless functions), **backbone** (data and state), and **canvas** (static hosting).

## Templates

| Template | Tier | Description |
|----------|------|-------------|
| [restaurant](restaurant/) | Hacker (free) | Restaurant site with dynamic menu, reservations, and confirmation emails |
| [booking-system](booking-system/) | Hacker (free) | Appointment booking with time slots and calendar management |
| [ecommerce](ecommerce/) | Hacker (free) | Online store with product catalog, cart, checkout, and order webhooks |
| [event-page](event-page/) | Hacker (free) | Event landing page with RSVP form and attendee tracking |
| [newsletter-signup](newsletter-signup/) | Hacker (free) | Email subscription with welcome emails and unsubscribe flow |
| [small-business](small-business/) | Hacker (free) | Business site with contact form and email notifications |

## Structure

Every template follows the same layout:

```
template-name/
├── drift.yaml          Deployment manifest
├── README.md           Setup guide and resource usage
├── atomic/             Serverless functions (one directory per function)
│   └── function-name/
│       ├── main.go
│       └── go.mod
├── backbone/           Data layer setup scripts and seed data
└── canvas/             Static HTML/CSS/JS site files
```

## Deploying a template

```bash
drift slice create myproject
drift slice use myproject
drift deploy drift.yaml
```

The `drift.yaml` manifest declares all functions, backbone resources (NoSQL collections, queues, secrets, cache), and canvas sites. `drift deploy` provisions everything in one step.
