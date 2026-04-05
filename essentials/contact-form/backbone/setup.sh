#!/usr/bin/env bash
# small-business/backbone/setup.sh

set -euo pipefail

echo "Setting up Backbone resources for small-business..."

drift backbone nosql write \
  --collection leads \
  --data '{"_setup":true}'
echo "✅ leads collection ready"

# Prime the contact queue (auto-created on first push).
drift backbone queue push contact-queue '{"_setup":true}'
echo "✅ contact-queue ready"

echo ""
echo "Deploy your functions:"
echo "    drift atomic deploy atomic/contact"
echo "    drift atomic deploy atomic/notify-contact"
echo ""
echo "Set secrets to receive email notifications for new leads:"
echo "    drift backbone secret set RESEND_API_KEY=<your-resend-key>"
echo "    drift backbone secret set SENDER_EMAIL=notifications@yourdomain.com"
echo "    drift backbone secret set OWNER_EMAIL=you@yourdomain.com"
echo ""
echo "The notify-contact function registers a queue trigger on deploy."
echo "To register it manually:"
echo "    drift atomic trigger register queue notify-contact \\"
echo "        --queue contact-queue \\"
echo "        --target /api/notify-contact"
