#!/bin/bash

# ==========================================
# XMeta Partner API — Deployment Script
# ==========================================
# Brings up postgres + backend + nginx via docker-compose.
# Frontend (xmeta-partner-web) is deployed separately on Vercel.

set -e

echo "🚀 Starting XMeta Partner API deployment..."

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
    echo "❌ Error: .env file not found."
    echo "   Run: cp env.example .env && nano .env"
    exit 1
fi

# Load + validate required environment.
set -a
source .env
set +a

REQUIRED_VARS=(
    "DB_USER"
    "DB_PASSWORD"
    "DB_NAME"
    "AWS_ACCESS_KEY_ID"
    "AWS_SECRET_ACCESS_KEY"
    "AWS_REGION"
    "COGNITO_USER_POOL_ID"
    "COGNITO_CLIENT_ID"
    "PARTNER_COGNITO_USER_POOL_ID"
    "PARTNER_COGNITO_CLIENT_ID"
    "INTERNAL_API_KEY"
    "ALLOWED_ORIGINS"
)

echo "🔍 Validating environment variables..."
MISSING=0
for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ]; then
        echo "   ❌ $var is not set"
        MISSING=1
    fi
done
if [ "$MISSING" = "1" ]; then
    echo "Fix the missing values in .env and re-run."
    exit 1
fi
echo "✅ All required variables are set."

# Pre-flight: Internal API key shouldn't be the placeholder.
if [ "$INTERNAL_API_KEY" = "your-internal-api-key" ] || [ "$INTERNAL_API_KEY" = "dev-internal-key-2026" ]; then
    echo "❌ INTERNAL_API_KEY is still set to a default placeholder. Rotate it."
    exit 1
fi

# Bring up the stack.
echo "🛑 Stopping any existing containers..."
$COMPOSE down

echo "🔨 Building images (no cache)..."
$COMPOSE build --no-cache

echo "🚀 Starting containers..."
$COMPOSE up -d

echo "⏳ Waiting 10s for the database to settle..."
sleep 10

echo "📊 Container status:"
$COMPOSE ps

echo "📝 Recent logs:"
$COMPOSE logs --tail=50

echo ""
echo "✅ Deployment finished."
echo ""
echo "📌 Backend API:    http://<server-ip>/api/v1/..."
echo "📌 Swagger docs:   http://<server-ip>/swagger/index.html"
echo "📌 Healthcheck:    http://<server-ip>/health"
echo ""
echo "💡 Useful commands:"
echo "   $COMPOSE logs -f backend         # follow backend logs"
echo "   $COMPOSE logs -f nginx           # follow nginx logs"
echo "   $COMPOSE restart backend         # restart only the API"
echo "   docker exec -it xmeta_partner_postgres psql -U \$DB_USER -d \$DB_NAME"
echo ""
echo "⚠️  Reminders:"
echo "   1. After first successful boot, set DB_AUTO_MIGRATE=false and DB_RUN_MIGRATIONS=false in .env"
echo "   2. Open ports 80 (HTTP) / 443 (HTTPS) in the firewall (ufw allow 80,443/tcp)"
echo "   3. Point partner-api.x-meta.com DNS at this server, then run certbot for TLS"
