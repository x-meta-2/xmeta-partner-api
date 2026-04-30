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

# NOTE: AWS_ACCESS_KEY_ID + AWS_SECRET_ACCESS_KEY are intentionally NOT
# required. On EC2 we attach an IAM Instance Profile and the SDK falls
# through to the default credential chain (IMDS). Leave both empty in
# .env for production; set them only for local dev outside AWS.
REQUIRED_VARS=(
    "DB_USER"
    "DB_PASSWORD"
    "DB_NAME"
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

echo "⏳ Waiting 10s for the backend to settle..."
sleep 10

echo "📊 Container status:"
$COMPOSE ps

echo "📝 Recent logs:"
$COMPOSE logs --tail=50

echo ""
echo "✅ Deployment finished."
echo ""
echo "📌 Backend API:    https://<your-domain>/api/v1/..."
echo "📌 Swagger docs:   https://<your-domain>/swagger/index.html"
echo "📌 Healthcheck:    https://<your-domain>/health"
echo ""
echo "💡 Useful commands:"
echo "   $COMPOSE logs -f backend         # follow backend logs"
echo "   $COMPOSE logs -f nginx           # follow nginx logs"
echo "   $COMPOSE restart backend         # restart only the API"
echo "   ./update.sh                       # pull + redeploy after a code change"
echo ""
echo "⚠️  Next steps:"
echo "   1. After first successful boot, set DB_AUTO_MIGRATE=false and DB_RUN_MIGRATIONS=false in .env"
echo "   2. Open ports 80 (HTTP) / 443 (HTTPS) in the firewall + AWS Security Group"
echo "   3. Point your DNS at this server, then run:"
echo "        sudo ./scripts/setup-ssl.sh <domain> <email>"
