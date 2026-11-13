#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
for file in data/*.csv; do
  table=$(basename "$file" .csv)
  echo "loading $table"
  docker compose --env-file .env.local exec -T clickhouse sh -c "clickhouse-client --user \"\$CLICKHOUSE_USER\" --password \"\$CLICKHOUSE_PASSWORD\" --database \"\$CLICKHOUSE_DATABASE\" --query \"INSERT INTO $table FORMAT CSVWithNames\"" < "$file"
done
echo "datasets loaded"
