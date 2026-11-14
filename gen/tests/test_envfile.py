import os
import tempfile
import unittest
from pathlib import Path

from gen.config import load_env_file


class LoadEnvFileTest(unittest.TestCase):
    def setUp(self):
        self.saved = dict(os.environ)

    def tearDown(self):
        os.environ.clear()
        os.environ.update(self.saved)

    def write(self, text):
        handle, path = tempfile.mkstemp(suffix=".env")
        with os.fdopen(handle, "w", encoding="utf-8", newline="\n") as file:
            file.write(text)
        return Path(path)

    def test_loads_plain_quoted_and_skips_comments(self):
        path = self.write(
            "TEST_DATASET_PLAIN=42\n"
            "\n"
            "# comment line\n"
            "TEST_DATASET_QUOTED=\"2023-11-01 00:00:00\"\n"
            "TEST_DATASET_SINGLE='energosbyt'\n"
        )
        try:
            load_env_file(path)
            self.assertEqual(os.environ["TEST_DATASET_PLAIN"], "42")
            self.assertEqual(os.environ["TEST_DATASET_QUOTED"], "2023-11-01 00:00:00")
            self.assertEqual(os.environ["TEST_DATASET_SINGLE"], "energosbyt")
        finally:
            path.unlink()

    def test_does_not_override_existing_env(self):
        os.environ["TEST_DATASET_EXISTING"] = "from-shell"
        path = self.write("TEST_DATASET_EXISTING=from-file\n")
        try:
            load_env_file(path)
            self.assertEqual(os.environ["TEST_DATASET_EXISTING"], "from-shell")
        finally:
            path.unlink()

    def test_missing_file_is_not_an_error(self):
        load_env_file(Path(tempfile.gettempdir()) / "missing-env-file")

    def test_real_env_file_values_load(self):
        load_env_file()
        self.assertEqual(os.environ["DATASET_SEED"], "42")
        self.assertEqual(os.environ["DATASET_TARIFF_V1_START"], "2023-11-01 00:00:00")
        self.assertEqual(os.environ["CLICKHOUSE_DATABASE"], "energosbyt")


if __name__ == "__main__":
    unittest.main()
