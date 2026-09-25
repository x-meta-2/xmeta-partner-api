#!/bin/sh

set -eu

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
LOG_FILE="${FUTURES_SYNC_LOG_FILE:-/var/log/xmeta-partner-futures-sync.log}"

if [ "$(date +%z)" = "+0800" ]; then
  CRON_TIME="0 8 * * *"
  TZ_NOTE="server timezone is UTC+8, scheduling at 08:00 local time"
else
  CRON_TIME="0 0 * * *"
  TZ_NOTE="server timezone is not UTC+8, scheduling at 00:00 server time (08:00 UTC+8)"
fi

CRON_CMD="cd ${PROJECT_DIR} && docker compose --profile jobs run --rm futures-commission-sync >> ${LOG_FILE} 2>&1"
CRON_LINE="${CRON_TIME} ${CRON_CMD}"
MARKER="xmeta-partner-futures-sync"

echo "Installing ${MARKER} cron"
echo "${TZ_NOTE}"
echo "${CRON_LINE}"

TMP_FILE="$(mktemp)"
trap 'rm -f "$TMP_FILE"' EXIT

crontab -l 2>/dev/null | grep -v "${MARKER}" > "$TMP_FILE" || true
{
  cat "$TMP_FILE"
  echo "# ${MARKER} - runs once daily at 08:00 UTC+8"
  echo "${CRON_LINE} # ${MARKER}"
} | crontab -

echo "Cron installed."
crontab -l | grep "${MARKER}"
