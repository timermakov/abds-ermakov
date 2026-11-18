import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
COMPOSE = ["docker", "compose", "--env-file", ".env.local"]
DROP = 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --query "DROP DATABASE IF EXISTS $CLICKHOUSE_DATABASE"'


def main() -> None:
    command = COMPOSE + ["exec", "-T", "clickhouse", "sh", "-c", DROP]
    result = subprocess.run(command, cwd=ROOT)
    if result.returncode != 0:
        sys.exit("teardown failed")
    print("database dropped")


if __name__ == "__main__":
    main()
