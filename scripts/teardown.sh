#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
docker compose --env-file .env.local exec -T clickhouse sh -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "DROP DATABASE IF EXISTS $CLICKHOUSE_DATABASE"'
echo "database dropped"
