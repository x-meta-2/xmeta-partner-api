#!/bin/bash

# ==========================================
# XMeta Partner API — Update / Redeploy
# ==========================================
# Pulls the latest commit from origin/main and recreates only what changed:
#  - Backend always rebuilds + restarts (most common code path).
#  - Nginx restarts ONLY if the new commit touched nginx/* (so its TLS
#    state and rate-limit counters stay intact otherwise).
#
# Postgres is not part of this stack — partner-api shares xmeta-admin's
# database. See README + docker-compose.yml.
#
# For first-time bring-up use ./setup.sh.

set -e

# Locate compose binary — prefer the v2 plugin, fall back to legacy.
if docker compose version >/dev/null 2>&1; then
    COMPOSE="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
    COMPOSE="docker-compose"
else
    echo "❌ Error: neither 'docker compose' nor 'docker-compose' is installed."
    exit 1
fi

# Ensure .env exists.
if [ ! -f .env ]; then
    echo "❌ Error: .env file not found. Run ./setup.sh first."
    exit 1
fi

# Show current state.
CURRENT_COMMIT=$(git rev-parse --short HEAD)
echo "🔎 Current commit: $CURRENT_COMMIT ($(git log -1 --pretty=%s))"

# Fetch + fast-forward main.
echo "📥 Pulling origin/main..."
git fetch origin main --quiet
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse origin/main)

# Track whether nginx config moved between LOCAL and REMOTE so we know
# whether to restart nginx after the pull.
NGINX_CHANGED=false

if [ "$LOCAL" = "$REMOTE" ]; then
    echo "✅ Already up to date with origin/main."
    read -p "🔁 Rebuild + recreate backend anyway? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "👋 Nothing to do."
        exit 0
    fi
else
    if git diff --name-only "$LOCAL" "$REMOTE" | grep -q '^nginx/'; then
        NGINX_CHANGED=true
    fi
    git pull --ff-only origin main
    NEW_COMMIT=$(git rev-parse --short HEAD)
    echo "✨ Updated to: $NEW_COMMIT ($(git log -1 --pretty=%s))"
    echo "📜 Changes:"
    git log --oneline "$CURRENT_COMMIT..$NEW_COMMIT" | head -20
fi

# Rebuild + recreate the backend.
echo "🔨 Rebuilding backend image..."
$COMPOSE build --pull backend

echo "♻️  Recreating backend container..."
$COMPOSE up -d --force-recreate backend

# If the pull touched nginx/, restart it so the new config / certs land.
if [ "$NGINX_CHANGED" = "true" ]; then
    echo "🌐 nginx/* changed — restarting nginx so the new config takes effect..."
    $COMPOSE restart nginx
fi

# Brief grace period for the new container to bind ports.
echo "⏳ Waiting 5s for backend to come up..."
sleep 5

# Status + tail.
echo "📊 Container status:"
$COMPOSE ps

echo "📝 Recent backend logs:"
$COMPOSE logs --tail=40 backend

echo ""
echo "✅ Update finished."
echo ""
echo "💡 If something looks wrong:"
echo "   $COMPOSE logs -f backend       # follow live"
echo "   $COMPOSE restart backend       # quick restart"
echo "   $COMPOSE restart nginx         # reload nginx config"
echo "   git log -5 --oneline           # check what got deployed"
