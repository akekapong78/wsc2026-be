#!/usr/bin/env bash
# Hits the local keepalive endpoint once — meant to run from crontab daily
# so the Supabase project stays active. Requires the server to be running
# (e.g. via a LaunchAgent/pm2/systemd, or start it before this fires).
#
# Add to crontab (runs every day at 09:00):
#   crontab -e
#   0 9 * * * /path/to/wsc2026-be/scripts/keepalive.sh >> /tmp/wsc2026-keepalive.log 2>&1

set -euo pipefail

URL="${KEEPALIVE_URL:-http://localhost:8080/api/v1/ops/keepalive}"

echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] pinging $URL"
curl -sf -X POST "$URL"
echo
