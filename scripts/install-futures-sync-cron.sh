#!/bin/sh

set -eu

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
FUTURES_LOG_FILE="${FUTURES_SYNC_LOG_FILE:-/var/log/xmeta-partner-futures-sync.log}"
TIER_LOG_FILE="${MONTHLY_TIER_REVIEW_LOG_FILE:-/var/log/xmeta-partner-monthly-tier-review.log}"

if [ "$(date +%z)" = "+0800" ]; then
  FUTURES_CRON_TIME="0 8 * * *"
  TIER_CRON_TIME="30 8 1 * *"
  TZ_NOTE="server timezone is UTC+8, scheduling daily 08:00 and monthly day 1 08:30 local time"
else
  FUTURES_CRON_TIME="0 0 * * *"
  TIER_CRON_TIME="30 0 1 * *"
  TZ_NOTE="server timezone is not UTC+8, scheduling daily 00:00 and monthly day 1 00:30 server time (08:00/08:30 UTC+8)"
fi

FUTURES_MARKER="xmeta-partner-futures-sync"
TIER_MARKER="xmeta-partner-monthly-tier-review"
FUTURES_CRON_CMD="cd ${PROJECT_DIR} && docker compose --profile jobs run --rm futures-commission-sync >> ${FUTURES_LOG_FILE} 2>&1"
TIER_CRON_CMD="cd ${PROJECT_DIR} && docker compose --profile jobs run --rm monthly-tier-review >> ${TIER_LOG_FILE} 2>&1"
FUTURES_CRON_LINE="${FUTURES_CRON_TIME} ${FUTURES_CRON_CMD}"
TIER_CRON_LINE="${TIER_CRON_TIME} ${TIER_CRON_CMD}"

echo "Installing partner cron jobs"
echo "${TZ_NOTE}"
echo "${FUTURES_CRON_LINE}"
echo "${TIER_CRON_LINE}"

TMP_FILE="$(mktemp)"
trap 'rm -f "$TMP_FILE"' EXIT

crontab -l 2>/dev/null | grep -v "${FUTURES_MARKER}" | grep -v "${TIER_MARKER}" > "$TMP_FILE" || true
{
  cat "$TMP_FILE"
  echo "# ${FUTURES_MARKER} - runs once daily at 08:00 UTC+8"
  echo "${FUTURES_CRON_LINE} # ${FUTURES_MARKER}"
  echo "# ${TIER_MARKER} - runs monthly on day 1 at 08:30 UTC+8"
  echo "${TIER_CRON_LINE} # ${TIER_MARKER}"
} | crontab -

echo "Cron installed."
crontab -l | grep -E "${FUTURES_MARKER}|${TIER_MARKER}"
