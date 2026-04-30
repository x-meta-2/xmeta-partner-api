#!/bin/bash
# ============================================================================
# scripts/setup-ssl.sh — one-shot Let's Encrypt setup for a fresh server
# ============================================================================
# Idempotent: re-running on a server that already has a cert and hook is
# safe (skips the cert step, refreshes the hook).
#
# Prerequisites
# - This compose stack is up and serving HTTP on port 80 (run ./setup.sh
#   first). Webroot mode requires nginx to be answering /.well-known/...
# - DNS for $DOMAIN already points at this server's public IP.
# - apt-based system (Ubuntu / Debian).
#
# Usage:
#   sudo ./scripts/setup-ssl.sh <domain> <email>
#   sudo ./scripts/setup-ssl.sh partners-api.x-meta.com admin@x-meta.com
#
# What it does:
#   1. apt install certbot
#   2. certbot certonly --webroot ...   (or skips if cert is recent)
#   3. Copies fullchain.pem + privkey.pem into ./nginx/ssl/
#   4. Installs the renewal deploy hook so future certbot.timer auto-renewals
#      copy the fresh cert into nginx/ssl/ and restart nginx.
#   5. Runs `certbot renew --dry-run` to verify the whole pipeline.

set -e

if [ "$EUID" -ne 0 ]; then
    echo "❌ Run as root: sudo $0 $@"
    exit 1
fi

DOMAIN="${1:-}"
EMAIL="${2:-}"

if [ -z "$DOMAIN" ] || [ -z "$EMAIL" ]; then
    echo "Usage: sudo $0 <domain> <email>"
    echo "Example: sudo $0 partners-api.x-meta.com admin@x-meta.com"
    exit 1
fi

APP_DIR="$(cd "$(dirname "$0")/.." && pwd)"
WEBROOT="$APP_DIR/nginx/certbot"
SSL_DIR="$APP_DIR/nginx/ssl"
HOOK_SRC="$APP_DIR/scripts/letsencrypt-deploy-hook.sh"
HOOK_DST="/etc/letsencrypt/renewal-hooks/deploy/copy-to-partner-api.sh"

echo "📂 App dir:  $APP_DIR"
echo "🌐 Domain:   $DOMAIN"
echo "📧 Email:    $EMAIL"
echo "🪣 Webroot:  $WEBROOT"

# ─── 1. Install certbot if missing ──────────────────────────────────────
if ! command -v certbot >/dev/null 2>&1; then
    echo "📦 Installing certbot via apt..."
    apt update
    apt install -y certbot
fi
echo "✓ certbot $(certbot --version | awk '{print $2}')"

# ─── 2. Ensure webroot folder exists ─────────────────────────────────────
mkdir -p "$WEBROOT" "$SSL_DIR"
chown -R ubuntu:ubuntu "$WEBROOT" "$SSL_DIR"

# ─── 3. Obtain or renew the cert via webroot ─────────────────────────────
if [ -d "/etc/letsencrypt/live/$DOMAIN" ]; then
    echo "🔄 Cert already exists for $DOMAIN — renewal will be handled by certbot.timer"
else
    echo "📜 Issuing initial cert via webroot..."
    certbot certonly --webroot \
        -w "$WEBROOT" \
        -d "$DOMAIN" \
        --email "$EMAIL" \
        --agree-tos \
        --no-eff-email \
        --non-interactive
fi

# ─── 4. Copy live cert into nginx/ssl/ ───────────────────────────────────
cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" "$SSL_DIR/"
cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem"   "$SSL_DIR/"
chown ubuntu:ubuntu "$SSL_DIR/"*.pem
echo "✓ Cert copied to $SSL_DIR"

# ─── 5. Install the renewal deploy hook ──────────────────────────────────
mkdir -p "$(dirname "$HOOK_DST")"
cp "$HOOK_SRC" "$HOOK_DST"
chmod +x "$HOOK_DST"
echo "✓ Deploy hook installed at $HOOK_DST"

# ─── 6. Make sure renewal config uses webroot, not standalone ────────────
RENEWAL_CONF="/etc/letsencrypt/renewal/$DOMAIN.conf"
if [ -f "$RENEWAL_CONF" ] && grep -q "^authenticator = standalone" "$RENEWAL_CONF"; then
    echo "🔧 Switching renewal config from standalone → webroot..."
    sed -i "s|^authenticator = standalone|authenticator = webroot\nwebroot_path = $WEBROOT|" "$RENEWAL_CONF"
fi

# ─── 7. Verify the whole pipeline ────────────────────────────────────────
echo "🧪 Running certbot renew --dry-run..."
certbot renew --dry-run

echo ""
echo "✅ SSL setup complete."
echo ""
echo "📌 Next renewal will run automatically via certbot.timer:"
systemctl list-timers --all 2>/dev/null | grep certbot || echo "(timer not visible — apt's certbot installs one in /etc/cron.d/certbot)"
