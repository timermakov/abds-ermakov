import csv
import os
from typing import Iterable, Sequence


class CsvWriter:
    def __init__(self, path: str, header: Sequence[str]) -> None:
        self.path = path
        self.rows = 0
        self._file = open(path, "w", newline="", encoding="utf-8")
        self._writer = csv.writer(self._file, lineterminator="\n")
        self._writer.writerow(header)

    def write(self, values: Iterable[object]) -> None:
        self._writer.writerow(values)
        self.rows += 1

    def close(self) -> None:
        self._file.close()

    def __enter__(self) -> "CsvWriter":
        return self

    def __exit__(self, exc_type, exc, tb) -> None:
        self.close()
