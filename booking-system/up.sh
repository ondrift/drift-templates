#!/usr/bin/env bash
# booking-system/up.sh — Deploy the booking system template to your Drift slice.

set -euo pipefail
cd "$(dirname "$0")"

echo "📅  Deploying booking-system template..."
echo ""

echo "▶  Setting up Backbone resources..."
bash backbone/setup.sh
echo ""

echo "▶  Deploying atomic functions..."
drift atomic deploy atomic/get-slots
drift atomic deploy atomic/book-slot
drift atomic deploy atomic/cancel-booking
drift atomic deploy atomic/confirm-booking
echo ""

echo "▶  Deploying canvas..."
drift canvas deploy canvas/
echo ""

echo "✅ Booking system template deployed!"
echo ""
echo "   To enable confirmation emails, set these secrets:"
echo "     drift backbone secret set RESEND_API_KEY=re_..."
echo "     drift backbone secret set SENDER_EMAIL=bookings@yourbusiness.com"
echo "     drift backbone secret set BUSINESS_NAME=\"Your Business Name\""
echo ""
echo "   To customise available time slots, edit:"
echo "     atomic/get-slots/get-slots.go  (allSlots variable)"
