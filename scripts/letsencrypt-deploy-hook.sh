#!/bin/bash
# ============================================================================
# Let's Encrypt deploy hook — runs after a successful certbot renewal
# ============================================================================
# Installed by scripts/setup-ssl.sh into:
#   /etc/letsencrypt/renewal-hooks/deploy/copy-to-partner-api.sh
#
# Certbot calls this whenever it finishes renewing a cert. The renewed
# fullchain.pem and privkey.pem live under /etc/letsencrypt/live/<domain>/
# (root-owned). Nginx (in our compose stack) reads them from
# <APP_DIR>/nginx/ssl/ via a docker volume mount, so the hook copies the
# new files into that path, fixes ownership, and bounces nginx.

set -e

DOMAIN="${RENEWED_DOMAINS:-partners-api.x-meta.com}"
APP_DIR="${APP_DIR:-/home/ubuntu/apps/xmeta-partner-api}"

echo "[letsencrypt-deploy-hook] domain=$DOMAIN app_dir=$APP_DIR"

cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" "$APP_DIR/nginx/ssl/"
cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem"   "$APP_DIR/nginx/ssl/"
chown ubuntu:ubuntu "$APP_DIR/nginx/ssl/"*.pem

cd "$APP_DIR" && docker compose restart nginx

echo "[letsencrypt-deploy-hook] done"
