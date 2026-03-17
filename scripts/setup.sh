#!/usr/bin/env bash
set -euo pipefail

# Unified setup script for Compliance Assistant API
# - Loads .env
# - Applies SQL migrations
# - Provisions tenant and admin
# - Seeds sample data (optional)
# - Starts API (optional) and ngrok (optional)

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
cd "$ROOT_DIR"

usage() {
  cat <<USAGE
Usage: scripts/setup.sh [options]

Options:
  --apply-migrations[=true|false]  Apply SQL migrations (default: true)
  --provision[=true|false]         Create tenant and admin (default: true)
  --seed[=true|false]              Seed dev data (default: false)
  --start[=true|false]             Start API with uvicorn (default: false)
  --ngrok[=true|false]             Start ngrok (default: false)
  --tenant-name=NAME               Tenant name (default: DEV)
  --admin-email=EMAIL              Admin email (default: admin@example.com)
  --admin-password=PASS            Admin password (default: admin123!)

Environment:
  DATABASE_URL (required for migrations)
  ALLOWED_ORIGINS, OKTA_ISSUER, OKTA_AUDIENCE, OPENAI_API_KEY, etc.
USAGE
}

# Defaults
APPLY_MIGRATIONS=true
PROVISION=true
SEED=false
START_API=false
START_NGROK=false
TENANT_NAME="DEV"
ADMIN_EMAIL="admin@example.com"
ADMIN_PASSWORD="admin123!"

for arg in "$@"; do
  case $arg in
    --apply-migrations|--apply-migrations=true) APPLY_MIGRATIONS=true ;;
    --apply-migrations=false) APPLY_MIGRATIONS=false ;;
    --provision|--provision=true) PROVISION=true ;;
    --provision=false) PROVISION=false ;;
    --seed|--seed=true) SEED=true ;;
    --seed=false) SEED=false ;;
    --start|--start=true) START_API=true ;;
    --start=false) START_API=false ;;
    --ngrok|--ngrok=true) START_NGROK=true ;;
    --ngrok=false) START_NGROK=false ;;
    --tenant-name=*) TENANT_NAME="${arg#*=}" ;;
    --admin-email=*) ADMIN_EMAIL="${arg#*=}" ;;
    --admin-password=*) ADMIN_PASSWORD="${arg#*=}" ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $arg"; usage; exit 1 ;;
  esac
done

if [[ -f .env ]]; then
  echo "Loading .env"
  set -a; source .env; set +a
fi

need_cmd() { command -v "$1" >/dev/null 2>&1 || { echo "Missing dependency: $1"; exit 1; }; }

if $APPLY_MIGRATIONS; then
  need_cmd psql
  if [[ -z "${DATABASE_URL:-}" ]]; then
    echo "DATABASE_URL is required to run migrations"; exit 1
  fi
  echo "Applying SQL migrations..."
  for file in $(ls -1 migrations/*.sql | sort); do
    echo "- $file"
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$file"
  done
fi

if $PROVISION; then
  echo "Provisioning tenant and admin via scripts/provision.py"
  if [[ -x venv/bin/python ]]; then
    TENANT_NAME="$TENANT_NAME" ADMIN_EMAIL="$ADMIN_EMAIL" ADMIN_PASSWORD="$ADMIN_PASSWORD" venv/bin/python scripts/provision.py
  else
    TENANT_NAME="$TENANT_NAME" ADMIN_EMAIL="$ADMIN_EMAIL" ADMIN_PASSWORD="$ADMIN_PASSWORD" python3 scripts/provision.py
  fi
fi

if $SEED; then
  if [[ -f scripts/seed_dev.py ]]; then
    echo "Seeding dev data..."
    if [[ -x venv/bin/python ]]; then
      venv/bin/python scripts/seed_dev.py || true
    else
      python3 scripts/seed_dev.py || true
    fi
  else
    echo "Skipping seeding (scripts/seed_dev.py not found)"
  fi
fi

if $START_API; then
  need_cmd uvicorn
  echo "Starting API on :8000"
  uvicorn app.main:app --host 0.0.0.0 --port 8000 &
  API_PID=$!
  echo "API PID: $API_PID"
fi

if $START_NGROK; then
  if command -v ngrok >/dev/null 2>&1; then
    echo "Starting ngrok tunnel on :8000"
    ngrok http 8000 &
  else
    echo "ngrok not installed; skipping"
  fi
fi

echo "Setup complete."
if $START_API; then
  echo "API running at http://localhost:8000 (Swagger if enabled: /docs)"
fi


