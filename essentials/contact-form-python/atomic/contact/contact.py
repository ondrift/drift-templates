# @atomic route=post:contact auth=none env=.env
# Stores a contact form submission in the leads NoSQL collection and
# pushes it to the contact-queue so the owner gets an email notification.

import drift
from datetime import datetime, timezone

def post_contact(body):
    name = (body.get("name") or "").strip()
    email = (body.get("email") or "").strip().lower()
    subject = (body.get("subject") or "").strip() or "New enquiry"
    message = (body.get("message") or "").strip()

    if not name or not email or not message or "@" not in email:
        return 400, "Bad Request", {"error": "name, email, and message are required"}

    lead_id = f"lead-{int(datetime.now(timezone.utc).timestamp() * 1e9)}"
    drift.nosql.collection("leads").insert({
        "id": lead_id,
        "name": name,
        "email": email,
        "subject": subject,
        "message": message,
        "received_at": datetime.now(timezone.utc).isoformat(),
        "status": "new",
    })

    drift.queue("contact-queue").push({
        "lead_id": lead_id,
        "name": name,
        "email": email,
        "subject": subject,
        "message": message,
    })

    return 200, "Message received", {"message": f"Thanks, {name}! We'll be in touch soon."}
