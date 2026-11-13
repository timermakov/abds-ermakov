#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
docker compose --env-file .env.local exec -T clickhouse sh -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "CREATE DATABASE IF NOT EXISTS $CLICKHOUSE_DATABASE"'
docker compose --env-file .env.local exec -T clickhouse sh -c 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --database "$CLICKHOUSE_DATABASE" --multiquery' < schema/001_tables.sql
echo "schema applied"
