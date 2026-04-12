// @atomic route=post:contact auth=none env=.env
// Stores a contact form submission in the leads NoSQL collection and
// pushes it to the contact-queue so the owner gets an email notification.

const drift = require('./drift-sdk');

async function postContact(body) {
    const name = (body.name || '').trim();
    const email = (body.email || '').trim().toLowerCase();
    const subject = (body.subject || '').trim() || 'New enquiry';
    const message = (body.message || '').trim();

    if (!name || !email || !message || !email.includes('@')) {
        return [400, 'Bad Request', { error: 'name, email, and message are required' }];
    }

    const leadId = `lead-${Date.now()}${Math.floor(Math.random() * 1e6)}`;
    await drift.nosql.collection('leads').insert({
        id: leadId,
        name,
        email,
        subject,
        message,
        received_at: new Date().toISOString(),
        status: 'new',
    });

    await drift.queue('contact-queue').push({
        lead_id: leadId,
        name,
        email,
        subject,
        message,
    });

    return [200, 'Message received', { message: `Thanks, ${name}! We'll be in touch soon.` }];
}

module.exports = { postContact };
