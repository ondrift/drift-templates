#!/usr/bin/env bash
# event-page/backbone/setup.sh

set -euo pipefail

echo "Setting up Backbone resources for event-page..."

drift backbone nosql write \
  --collection rsvps \
  --data '{"_setup":true}'
echo "✅ rsvps collection ready"

# Seed the RSVP stats cache (updated by your backend or a scheduled function).
drift backbone cache set rsvp:stats '{"rsvp_count":0,"guest_count":0}'
echo "✅ rsvp:stats cache key ready"

echo ""
echo "Deploy your functions:"
echo "    drift atomic deploy atomic/rsvp"
echo "    drift atomic deploy atomic/get-rsvps"
echo ""
echo "get-rsvps is protected by an API key. Set one after deploying:"
echo "    drift atomic auth set --function get-rsvps --method GET --key <your-secret-key>"
