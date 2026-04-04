#!/usr/bin/env bash
# event-page/up.sh — Deploy the event page template to your Drift slice.

set -euo pipefail
cd "$(dirname "$0")"

echo "🎟️  Deploying event-page template..."
echo ""

echo "▶  Setting up Backbone resources..."
bash backbone/setup.sh
echo ""

echo "▶  Deploying atomic functions..."
drift atomic deploy atomic/rsvp
drift atomic deploy atomic/get-rsvps
echo ""

echo "▶  Deploying canvas..."
drift canvas deploy canvas/
echo ""

echo "✅ Event page template deployed!"
echo ""
echo "   The RSVP counter on the page starts at zero."
echo "   Update it manually after RSVPs come in:"
echo "     drift backbone cache set rsvp:stats '{\"rsvp_count\":42,\"guest_count\":67}'"
echo ""
echo "   To protect the get-rsvps endpoint with an API key:"
echo "     drift atomic auth set --function get-rsvps --method GET --key <your-secret-key>"
