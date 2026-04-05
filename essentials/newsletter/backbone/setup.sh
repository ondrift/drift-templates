#!/usr/bin/env bash
# newsletter-signup/backbone/setup.sh
# Run this once after deploying your functions to initialize Backbone resources.
# Usage: bash backbone/setup.sh

set -euo pipefail

echo "Setting up Backbone resources for newsletter-signup..."

# Create the subscribers collection by writing a seed document.
# (Collections are created implicitly on first write.)
drift backbone nosql write \
  --collection subscribers \
  --data '{"_setup":true,"note":"seed document — safe to ignore"}'

echo "✅ subscribers collection ready"

# Push a throwaway message to create the signup-queue.
# The send-welcome trigger will drain it immediately.
drift backbone queue push signup-queue '{"_setup":true}'

echo "✅ signup-queue ready"

# Store your Resend API key and sender address as Backbone secrets.
# The send-welcome function reads these at runtime.
echo ""
echo "⚠️  You still need to set your email secrets:"
echo "    drift backbone secret set RESEND_API_KEY=re_..."
echo "    drift backbone secret set SENDER_EMAIL=hello@yourdomain.com"
echo ""
echo "Setup complete. Deploy your functions:"
echo "    drift atomic deploy atomic/subscribe"
echo "    drift atomic deploy atomic/unsubscribe"
echo "    drift atomic deploy atomic/send-welcome"
