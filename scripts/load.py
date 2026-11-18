import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
COMPOSE = ["docker", "compose", "--env-file", ".env.local"]
INSERT = 'clickhouse-client --user "$CLICKHOUSE_USER" --password "$CLICKHOUSE_PASSWORD" --database "$CLICKHOUSE_DATABASE" --query "INSERT INTO {table} FORMAT CSVWithNames"'


def main() -> None:
    files = sorted((ROOT / "data").glob("*.csv"))
    if not files:
        sys.exit("no csv files in data/: run make gen first")
    for file in files:
        table = file.stem
        print(f"loading {table}")
        with file.open("rb") as handle:
            command = COMPOSE + ["exec", "-T", "clickhouse", "sh", "-c", INSERT.format(table=table)]
            result = subprocess.run(command, stdin=handle, cwd=ROOT)
        if result.returncode != 0:
            sys.exit(f"loading {table} failed")
    print("datasets loaded")


if __name__ == "__main__":
    main()
