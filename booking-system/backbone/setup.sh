#!/usr/bin/env bash
# booking-system/backbone/setup.sh

set -euo pipefail

echo "Setting up Backbone resources for booking-system..."

drift backbone nosql write \
  --collection bookings \
  --data '{"_setup":true}'
echo "✅ bookings collection ready"

# Prime the queue (auto-created on first push).
drift backbone queue push booking-queue '{"_setup":true}'
echo "✅ booking-queue ready"

echo ""
echo "Deploy your functions:"
echo "    drift atomic deploy atomic/book-slot"
echo "    drift atomic deploy atomic/get-slots"
echo "    drift atomic deploy atomic/cancel-booking"
echo "    drift atomic deploy atomic/confirm-booking"
echo ""
echo "Set secrets for confirmation emails:"
echo "    drift backbone secret set RESEND_API_KEY=<your-resend-key>"
echo "    drift backbone secret set SENDER_EMAIL=bookings@yourbusiness.com"
echo "    drift backbone secret set BUSINESS_NAME=\"Your Business Name\""
echo ""
echo "The confirm-booking function registers a queue trigger on deploy."
echo "To register it manually:"
echo "    drift atomic trigger register queue confirm-booking \\"
echo "        --queue booking-queue \\"
echo "        --target /api/confirm-booking"
