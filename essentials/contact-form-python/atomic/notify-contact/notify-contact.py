# @atomic route=post:notify-contact auth=none env=.env
# drift:trigger queue contact-queue poll=5000ms retry=3
#
# Consumes the contact-queue and forwards new lead details to the
# business owner via Resend.

import os
import json
import drift

def post_notify_contact(body):
    api_key = os.environ.get("RESEND_API_KEY", "")
    sender = os.environ.get("SENDER_EMAIL", "")
    owner = os.environ.get("OWNER_EMAIL", "")

    if not api_key or not sender or not owner:
        return 200, "No email configured", {
            "note": "Set RESEND_API_KEY, SENDER_EMAIL, and OWNER_EMAIL secrets to enable notifications",
        }

    lead_id = body.get("lead_id", "")
    name = body.get("name", "")
    email = body.get("email", "")
    subject = body.get("subject", "")
    message = body.get("message", "")

    email_subject = f"[New Lead] {subject} — {name}"
    html = f"""
<p>A new message arrived via your contact form:</p>
<table style="border-collapse:collapse;width:100%;max-width:560px">
  <tr><td style="padding:6px 12px;color:#555;width:120px"><strong>From</strong></td><td style="padding:6px 12px">{name} &lt;{email}&gt;</td></tr>
  <tr style="background:#f9fafb"><td style="padding:6px 12px;color:#555"><strong>Subject</strong></td><td style="padding:6px 12px">{subject}</td></tr>
  <tr><td style="padding:6px 12px;color:#555"><strong>Message</strong></td><td style="padding:6px 12px">{message}</td></tr>
  <tr style="background:#f9fafb"><td style="padding:6px 12px;color:#555"><strong>Lead ID</strong></td><td style="padding:6px 12px;font-size:0.85em;color:#888">{lead_id}</td></tr>
</table>
<p style="margin-top:1.5rem">Reply directly to <a href="mailto:{email}">{email}</a> to respond.</p>"""

    resp = drift.http_request(
        "POST",
        "https://api.resend.com/emails",
        {"Authorization": f"Bearer {api_key}", "Content-Type": "application/json"},
        json.dumps({
            "from": sender,
            "to": [owner],
            "reply_to": email,
            "subject": email_subject,
            "html": html,
        }).encode(),
    )

    if resp.get("status", 0) >= 400:
        return 500, "Email error", {"error": f"resend returned {resp.get('status')}"}

    drift.log(f"[notify-contact] owner notified about lead from {email}")
    return 200, "Notification sent", {"message": f"Owner notified about lead from {email}"}
