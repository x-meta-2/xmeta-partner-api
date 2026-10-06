#!/bin/sh

set -eu

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
LOG_DIR="${PARTNER_JOB_LOG_DIR:-${PROJECT_DIR}/logs}"
FUTURES_LOG_FILE="${FUTURES_SYNC_LOG_FILE:-${LOG_DIR}/futures-sync.log}"
TIER_LOG_FILE="${MONTHLY_TIER_REVIEW_LOG_FILE:-${LOG_DIR}/monthly-tier-review.log}"

mkdir -p "$LOG_DIR"

if [ "$(date +%z)" = "+0800" ]; then
  FUTURES_CRON_TIME="5 0 * * *"
  TIER_CRON_TIME="10 0 1 * *"
  TZ_NOTE="server timezone is UTC+8, scheduling daily 00:05 and monthly day 1 00:10 local time"
  FUTURES_CRON_GUARD=""
  TIER_CRON_GUARD=""
else
  FUTURES_CRON_TIME="5 16 * * *"
  TIER_CRON_TIME="10 16 28-31 * *"
  TZ_NOTE="server timezone is not UTC+8, scheduling daily 00:05 and monthly day 1 00:10 Asia/Ulaanbaatar time"
  FUTURES_CRON_GUARD='[ "$(TZ=Asia/Ulaanbaatar date +\%H:\%M)" = "00:05" ] && '
  TIER_CRON_GUARD='[ "$(TZ=Asia/Ulaanbaatar date +\%d\ \%H:\%M)" = "01 00:10" ] && '
fi

FUTURES_MARKER="xmeta-partner-futures-sync"
TIER_MARKER="xmeta-partner-monthly-tier-review"
FUTURES_CRON_CMD="cd ${PROJECT_DIR} && docker compose --profile jobs run --rm futures-commission-sync >> ${FUTURES_LOG_FILE} 2>&1"
TIER_CRON_CMD="cd ${PROJECT_DIR} && docker compose --profile jobs run --rm monthly-tier-review >> ${TIER_LOG_FILE} 2>&1"
FUTURES_CRON_LINE="${FUTURES_CRON_TIME} ${FUTURES_CRON_GUARD}${FUTURES_CRON_CMD}"
TIER_CRON_LINE="${TIER_CRON_TIME} ${TIER_CRON_GUARD}${TIER_CRON_CMD}"

echo "Installing partner cron jobs"
echo "${TZ_NOTE}"
echo "${FUTURES_CRON_LINE}"
echo "${TIER_CRON_LINE}"

TMP_FILE="$(mktemp)"
trap 'rm -f "$TMP_FILE"' EXIT

crontab -l 2>/dev/null | grep -v "${FUTURES_MARKER}" | grep -v "${TIER_MARKER}" > "$TMP_FILE" || true
{
  cat "$TMP_FILE"
  echo "# ${FUTURES_MARKER} - runs once daily at 00:05 UTC+8"
  echo "${FUTURES_CRON_LINE} # ${FUTURES_MARKER}"
  echo "# ${TIER_MARKER} - runs monthly on day 1 at 00:10 UTC+8"
  echo "${TIER_CRON_LINE} # ${TIER_MARKER}"
} | crontab -

echo "Cron installed."
crontab -l | grep -E "${FUTURES_MARKER}|${TIER_MARKER}"
