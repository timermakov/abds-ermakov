from datetime import date, datetime, time
import unittest

from gen.config import DatasetConfig
from gen.entities import (
    binding_rows,
    gauge_rows,
    realty_rows,
    replaced_realties,
    tariff_rows,
    user_rows,
)
from gen.readings import replacements


def make_config(**overrides) -> DatasetConfig:
    defaults = {
        "seed": 7,
        "realty_count": 12,
        "units_per_building": 4,
        "non_residential_units": 1,
        "readings_start": date(2023, 11, 1),
        "readings_end": date(2023, 11, 30),
        "correction_ratio": 0.05,
        "replacement_ratio": 0.25,
        "replacement_date_fraction": 0.5,
        "tariff_v1_start": datetime(2023, 11, 1, 0, 0, 0),
        "tariff_v2_start": datetime(2024, 1, 1, 0, 0, 0),
        "tariff_open_end": datetime(2099, 1, 1, 0, 0, 0),
        "tariff_v1_day_price": 620,
        "tariff_v1_night_price": 310,
        "tariff_v2_day_price": 700,
        "tariff_v2_night_price": 350,
        "dim_created_at": datetime(2023, 10, 1, 9, 0, 0),
        "binding_created_at": datetime(2023, 10, 20, 10, 0, 0),
        "day_kw_delta_min": 2,
        "day_kw_delta_max": 12,
        "night_kw_delta_min": 1,
        "night_kw_delta_max": 6,
        "counter_start_min": 100,
        "counter_start_max": 1000,
        "correction_increment_min": 1,
        "correction_increment_max": 5,
        "reading_created_time": time(21, 0, 0),
        "correction_created_time": time(23, 0, 0),
        "invoice_created_day": 27,
        "invoice_created_time": time(10, 0, 0),
        "invoice_deadline_day": 10,
        "output_dir": "data",
    }
    defaults.update(overrides)
    return DatasetConfig(**defaults)


class RealtyRowsTest(unittest.TestCase):
    def test_rows_match_realty_count(self) -> None:
        rows = list(realty_rows(make_config()))
        self.assertEqual(len(rows), 12)

    def test_ids_start_at_one(self) -> None:
        rows = list(realty_rows(make_config()))
        self.assertEqual([row[0] for row in rows], list(range(1, 13)))

    def test_positive_physical_fields(self) -> None:
        rows = list(realty_rows(make_config()))
        for row in rows:
            self.assertGreater(row[2], 0)
            self.assertGreater(row[6], 0)
            self.assertGreater(row[7], 0)

    def test_types_are_valid(self) -> None:
        rows = list(realty_rows(make_config()))
        for row in rows:
            self.assertIn(row[3], ("flat", "nrcp"))

    def test_non_residential_has_no_number_and_rooms(self) -> None:
        rows = list(realty_rows(make_config()))
        for row in rows:
            if row[3] == "nrcp":
                self.assertEqual(row[4], "")
                self.assertEqual(row[5], 0)
            else:
                self.assertNotEqual(row[4], "")
                self.assertGreater(row[5], 0)

    def test_addresses_are_not_empty(self) -> None:
        rows = list(realty_rows(make_config()))
        for row in rows:
            self.assertNotEqual(row[1], "")


class GaugeRowsTest(unittest.TestCase):
    def test_rows_match_realty_count_plus_replacements(self) -> None:
        config = make_config()
        replaced = replaced_realties(config)
        rows = list(gauge_rows(config, replaced))
        self.assertEqual(len(rows), 12 + len(replaced))

    def test_ids_are_unique(self) -> None:
        config = make_config()
        rows = list(gauge_rows(config, replaced_realties(config)))
        ids = [row[0] for row in rows]
        self.assertEqual(len(ids), len(set(ids)))


class TariffRowsTest(unittest.TestCase):
    def test_two_versions_of_one_tariff(self) -> None:
        rows = list(tariff_rows(make_config()))
        self.assertEqual(len(rows), 2)
        self.assertEqual([row[0] for row in rows], [1, 1])

    def test_versions_chain(self) -> None:
        rows = list(tariff_rows(make_config()))
        self.assertEqual(rows[0][4], "2023-11-01 00:00:00")
        self.assertEqual(rows[0][5], "2024-01-01 00:00:00")
        self.assertEqual(rows[1][4], "2024-01-01 00:00:00")
        self.assertEqual(rows[1][5], "2099-01-01 00:00:00")

    def test_prices_are_positive(self) -> None:
        rows = list(tariff_rows(make_config()))
        for row in rows:
            self.assertGreater(row[2], 0)
            self.assertGreater(row[3], 0)


class UserRowsTest(unittest.TestCase):
    def test_rows_match_realty_count(self) -> None:
        rows = list(user_rows(make_config()))
        self.assertEqual(len(rows), 12)

    def test_phones_are_unique_and_prefixed(self) -> None:
        rows = list(user_rows(make_config()))
        phones = [row[2] for row in rows]
        self.assertEqual(len(phones), len(set(phones)))
        for phone in phones:
            self.assertTrue(phone.startswith("+7900"))

    def test_phone_prev_is_empty_initially(self) -> None:
        rows = list(user_rows(make_config()))
        for row in rows:
            self.assertEqual(row[3], "")

    def test_fio_has_three_words(self) -> None:
        rows = list(user_rows(make_config()))
        for row in rows:
            self.assertEqual(len(row[1].split(" ")), 3)


class BindingRowsTest(unittest.TestCase):
    def test_rows_match_realty_count_plus_two_per_replacement(self) -> None:
        config = make_config()
        replaced = replaced_realties(config)
        replacements_map = replacements(config, replaced)
        rows = list(binding_rows(config, replacements_map))
        self.assertEqual(len(rows), 12 + 2 * len(replaced))

    def test_first_realty_has_single_install(self) -> None:
        config = make_config()
        replacements_map = replacements(config, replaced_realties(config))
        rows = list(binding_rows(config, replacements_map))
        first = [row for row in rows if row[1] == 1]
        self.assertEqual(first, [(1, 1, "2023-10-20 10:00:00", 1, 1)])

    def test_replaced_realty_has_cancel_and_new_install(self) -> None:
        config = make_config()
        replaced = replaced_realties(config)
        replacements_map = replacements(config, replaced)
        rows = list(binding_rows(config, replacements_map))
        target = replaced[0]
        rows_for_realty = [row for row in rows if row[1] == target]
        self.assertEqual(len(rows_for_realty), 3)
        self.assertEqual(rows_for_realty[1][0], target)
        self.assertEqual(rows_for_realty[1][3], -1)
        self.assertEqual(rows_for_realty[1][4], 1)
        self.assertEqual(rows_for_realty[2][0], 12 + 1)
        self.assertEqual(rows_for_realty[2][3], 1)
        self.assertEqual(rows_for_realty[2][4], 2)

    def test_cancel_and_install_share_version_and_gauge(self) -> None:
        config = make_config()
        replaced = replaced_realties(config)
        replacements_map = replacements(config, replaced)
        rows = list(binding_rows(config, replacements_map))
        for realty_id in replaced:
            initial = [row for row in rows if row[1] == realty_id and row[4] == 1]
            self.assertEqual(len(initial), 2)
            self.assertEqual(initial[0][0], initial[1][0])
            self.assertEqual(initial[0][3] + initial[1][3], 0)

    def test_first_realty_never_replaced(self) -> None:
        self.assertNotIn(1, replaced_realties(make_config()))


if __name__ == "__main__":
    unittest.main()
