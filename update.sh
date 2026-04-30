#!/bin/bash

# ==========================================
# XMeta Partner API — Update / Redeploy
# ==========================================
# Pulls the latest commit from origin/main, rebuilds the backend image,
# and recreates ONLY the backend container — postgres + nginx keep their
# data and TLS state.
#
# Run this whenever you want to deploy new code without touching the DB.
# For first-time bring-up use ./setup.sh instead.

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

if [ "$LOCAL" = "$REMOTE" ]; then
    echo "✅ Already up to date with origin/main."
    read -p "🔁 Rebuild + recreate backend anyway? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "👋 Nothing to do."
        exit 0
    fi
else
    git pull --ff-only origin main
    NEW_COMMIT=$(git rev-parse --short HEAD)
    echo "✨ Updated to: $NEW_COMMIT ($(git log -1 --pretty=%s))"
    echo "📜 Changes:"
    git log --oneline "$CURRENT_COMMIT..$NEW_COMMIT" | head -20
fi

# Rebuild + recreate ONLY the backend.
echo "🔨 Rebuilding backend image..."
$COMPOSE build --pull backend

echo "♻️  Recreating backend container..."
$COMPOSE up -d --no-deps --force-recreate backend

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
echo "   git log -5 --oneline           # check what got deployed"
