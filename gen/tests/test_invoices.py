from datetime import date, datetime
import random
import unittest

from gen.entities import replaced_realties
from gen.invoices import invoice_rows
from gen.readings import month_index, points, replacements

from gen.tests.test_entities import make_config


def consumption_of(config, rng, replacements_map, months_list):
    consumption = {}
    for point in points(config, rng, replacements_map):
        if point.corrected:
            continue
        key = (point.realty_id, month_index(point.reading_date, config.readings_start))
        day_kw, night_kw = consumption.get(key, (0, 0))
        consumption[key] = (day_kw + point.day_delta, night_kw + point.night_delta)
    return consumption


class InvoiceRowsTest(unittest.TestCase):
    def test_one_invoice_per_realty_per_month(self) -> None:
        config = make_config()
        replacements_map = replacements(config, replaced_realties(config))
        months_list = [(2023, 11)]
        consumption = consumption_of(config, random.Random(config.seed), replacements_map, months_list)
        rows = list(invoice_rows(config, consumption, months_list, replacements_map))
        self.assertEqual(len(rows), 12)
        self.assertEqual(len({row.realty_id for row in rows}), 12)

    def test_price_uses_version_one_before_change(self) -> None:
        config = make_config()
        replacements_map = replacements(config, replaced_realties(config))
        months_list = [(2023, 11)]
        rng = random.Random(config.seed)
        consumption = consumption_of(config, rng, replacements_map, months_list)
        rows = list(invoice_rows(config, consumption, months_list, replacements_map))
        for row in rows:
            day_kw, night_kw = consumption[(row.realty_id, 0)]
            self.assertEqual(row.price, day_kw * 620 + night_kw * 310)

    def test_price_uses_version_two_after_change(self) -> None:
        config = make_config(
            readings_end=date(2024, 1, 15),
            tariff_v2_start=datetime(2024, 1, 1, 0, 0, 0),
        )
        replacements_map = replacements(config, replaced_realties(config))
        months_list = [(2023, 11), (2023, 12), (2024, 1)]
        rng = random.Random(config.seed)
        consumption = consumption_of(config, rng, replacements_map, months_list)
        rows = list(invoice_rows(config, consumption, months_list, replacements_map))
        self.assertEqual(len(rows), 36)
        for row in rows:
            if row.created_at.month == 1 and row.created_at.year == 2024:
                day_kw, night_kw = consumption[(row.realty_id, 2)]
                self.assertEqual(row.price, day_kw * 700 + night_kw * 350)

    def test_deadline_is_tenth_of_next_month(self) -> None:
        config = make_config()
        replacements_map = replacements(config, replaced_realties(config))
        months_list = [(2023, 11)]
        rng = random.Random(config.seed)
        consumption = consumption_of(config, rng, replacements_map, months_list)
        rows = list(invoice_rows(config, consumption, months_list, replacements_map))
        for row in rows:
            self.assertEqual(row.deadline, date(2023, 12, 10))

    def test_created_at_is_invoice_day(self) -> None:
        config = make_config()
        replacements_map = replacements(config, replaced_realties(config))
        months_list = [(2023, 11)]
        rng = random.Random(config.seed)
        consumption = consumption_of(config, rng, replacements_map, months_list)
        rows = list(invoice_rows(config, consumption, months_list, replacements_map))
        for row in rows:
            self.assertEqual(row.created_at, datetime(2023, 11, 27, 10, 0, 0))

    def test_replaced_realty_invoices_use_new_gauge(self) -> None:
        config = make_config()
        replaced = replaced_realties(config)
        replacements_map = replacements(config, replaced)
        months_list = [(2023, 11)]
        rng = random.Random(config.seed)
        consumption = consumption_of(config, rng, replacements_map, months_list)
        rows = list(invoice_rows(config, consumption, months_list, replacements_map))
        target = replaced[0]
        replacement = replacements_map[target]
        for row in rows:
            if row.realty_id == target:
                self.assertEqual(row.gauge_id, replacement.gauge_id)

    def test_invoices_are_positive(self) -> None:
        config = make_config()
        replacements_map = replacements(config, replaced_realties(config))
        months_list = [(2023, 11)]
        rng = random.Random(config.seed)
        consumption = consumption_of(config, rng, replacements_map, months_list)
        rows = list(invoice_rows(config, consumption, months_list, replacements_map))
        for row in rows:
            self.assertGreater(row.day_kw, 0)
            self.assertGreater(row.night_kw, 0)
            self.assertGreater(row.price, 0)


if __name__ == "__main__":
    unittest.main()
