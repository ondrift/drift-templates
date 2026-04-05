#!/usr/bin/env bash
# small-business/up.sh — Deploy the small business landing page to your Drift slice.

set -euo pipefail
cd "$(dirname "$0")"

echo "🏢  Deploying small-business template..."
echo ""

echo "▶  Setting up Backbone resources..."
bash backbone/setup.sh
echo ""

echo "▶  Deploying atomic functions..."
drift atomic deploy atomic/contact
drift atomic deploy atomic/notify-contact
echo ""

echo "▶  Deploying canvas..."
drift canvas deploy canvas/
echo ""

echo "✅ Small business landing page deployed!"
echo ""
echo "   Contact form submissions are stored in the 'leads' collection."
echo "   To also receive email notifications for new leads, set:"
echo "     drift backbone secret set RESEND_API_KEY=re_..."
echo "     drift backbone secret set SENDER_EMAIL=notifications@yourdomain.com"
echo "     drift backbone secret set OWNER_EMAIL=you@yourdomain.com"
