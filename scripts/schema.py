import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
COMPOSE = ["docker", "compose", "--env-file", ".env.local"]


def exec_clickhouse(query: str, stdin=None) -> int:
    command = COMPOSE + ["exec", "-T", "clickhouse", "sh", "-c", query]
    result = subprocess.run(command, stdin=stdin, cwd=ROOT)
    return result.returncode


CLIENT = 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD"'
DROP = CLIENT + ' --query "DROP DATABASE IF EXISTS $CLICKHOUSE_DATABASE"'
CREATE = CLIENT + ' --query "CREATE DATABASE IF NOT EXISTS $CLICKHOUSE_DATABASE"'
TABLES = CLIENT + ' --database "$CLICKHOUSE_DATABASE" --multiquery'
INSERT = CLIENT + ' --database "$CLICKHOUSE_DATABASE" --query "INSERT INTO {table} FORMAT CSVWithNames"'


def main() -> None:
    schema = ROOT / "schema" / "001_tables.sql"
    if not schema.exists():
        sys.exit(f"schema file {schema} is missing")
    for query in (DROP, CREATE):
        if exec_clickhouse(query) != 0:
            sys.exit("schema queries failed")
    with schema.open("rb") as handle:
        if exec_clickhouse(TABLES, stdin=handle) != 0:
            sys.exit("schema tables failed")
    print("schema applied")


if __name__ == "__main__":
    main()
