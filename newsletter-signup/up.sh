#!/usr/bin/env bash
# newsletter-signup/up.sh — Deploy the newsletter signup template to your Drift slice.

set -euo pipefail
cd "$(dirname "$0")"

echo "📧  Deploying newsletter-signup template..."
echo ""

echo "▶  Setting up Backbone resources..."
bash backbone/setup.sh
echo ""

echo "▶  Deploying atomic functions..."
drift atomic deploy atomic/subscribe
drift atomic deploy atomic/unsubscribe
drift atomic deploy atomic/send-welcome
echo ""

echo "▶  Deploying canvas..."
drift canvas deploy canvas/
echo ""

echo "✅ Newsletter signup template deployed!"
echo ""
echo "   To enable welcome emails, set these secrets:"
echo "     drift backbone secret set RESEND_API_KEY=re_..."
echo "     drift backbone secret set SENDER_EMAIL=hello@yourdomain.com"
