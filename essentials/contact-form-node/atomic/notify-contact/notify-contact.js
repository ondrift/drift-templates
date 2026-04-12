// @atomic route=post:notify-contact auth=none env=.env
// drift:trigger queue contact-queue poll=5000ms retry=3
//
// Consumes the contact-queue and forwards new lead details to the
// business owner via Resend.

const drift = require('./drift-sdk');

async function postNotifyContact(body) {
    const apiKey = process.env.RESEND_API_KEY || '';
    const sender = process.env.SENDER_EMAIL || '';
    const owner = process.env.OWNER_EMAIL || '';

    if (!apiKey || !sender || !owner) {
        return [200, 'No email configured', {
            note: 'Set RESEND_API_KEY, SENDER_EMAIL, and OWNER_EMAIL secrets to enable notifications',
        }];
    }

    const { lead_id: leadId, name, email, subject, message } = body;
    const emailSubject = `[New Lead] ${subject} — ${name}`;
    const html = `
<p>A new message arrived via your contact form:</p>
<table style="border-collapse:collapse;width:100%;max-width:560px">
  <tr><td style="padding:6px 12px;color:#555;width:120px"><strong>From</strong></td><td style="padding:6px 12px">${name} &lt;${email}&gt;</td></tr>
  <tr style="background:#f9fafb"><td style="padding:6px 12px;color:#555"><strong>Subject</strong></td><td style="padding:6px 12px">${subject}</td></tr>
  <tr><td style="padding:6px 12px;color:#555"><strong>Message</strong></td><td style="padding:6px 12px">${message}</td></tr>
  <tr style="background:#f9fafb"><td style="padding:6px 12px;color:#555"><strong>Lead ID</strong></td><td style="padding:6px 12px;font-size:0.85em;color:#888">${leadId}</td></tr>
</table>
<p style="margin-top:1.5rem">Reply directly to <a href="mailto:${email}">${email}</a> to respond.</p>`;

    const resp = await drift.httpRequest(
        'POST',
        'https://api.resend.com/emails',
        { Authorization: `Bearer ${apiKey}`, 'Content-Type': 'application/json' },
        JSON.stringify({
            from: sender,
            to: [owner],
            reply_to: email,
            subject: emailSubject,
            html,
        }),
    );

    if (resp.status >= 400) {
        return [500, 'Email error', { error: `resend returned ${resp.status}` }];
    }

    drift.log(`[notify-contact] owner notified about lead from ${email}`);
    return [200, 'Notification sent', { message: `Owner notified about lead from ${email}` }];
}

module.exports = { postNotifyContact };
