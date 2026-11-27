from datetime import datetime
import random
import unittest

from gen.entities import replaced_realties
from gen.invoices import month_end
from gen.readings import points, replacements

from gen.tests.test_entities import make_config


class ReadingsPointsTest(unittest.TestCase):
    def setUp(self) -> None:
        self.config = make_config()
        replaced = replaced_realties(self.config)
        self.replacements = replacements(self.config, replaced)
        self.rng = random.Random(self.config.seed)
        self.points = list(points(self.config, self.rng, self.replacements))

    def test_base_rows_cover_every_realty_and_day(self) -> None:
        base = [point for point in self.points if not point.corrected]
        self.assertEqual(len(base), 12 * 30)
        keys = {(point.realty_id, point.reading_date) for point in base}
        self.assertEqual(len(keys), 360)

    def test_correction_share_is_positive_and_limited(self) -> None:
        corrected = [point for point in self.points if point.corrected]
        self.assertGreater(len(corrected), 0)
        self.assertLess(len(corrected), 0.25 * len(self.points))

    def test_correction_wins_over_base_row(self) -> None:
        by_key = {}
        for point in self.points:
            by_key.setdefault((point.realty_id, point.reading_date), []).append(point)
        for group in by_key.values():
            latest = max(group, key=lambda point: point.created_at)
            if latest.corrected:
                base = min(group, key=lambda point: point.created_at)
                self.assertGreater(latest.created_at, base.created_at)
                self.assertGreater(latest.day_kw, base.day_kw)
                self.assertGreater(latest.night_kw, base.night_kw)

    def test_counters_are_monotonic_per_realty(self) -> None:
        seen = {}
        for point in self.points:
            last = seen.get(point.realty_id)
            if last is not None:
                self.assertGreaterEqual(point.day_kw, last[0])
                self.assertGreaterEqual(point.night_kw, last[1])
            seen[point.realty_id] = (point.day_kw, point.night_kw)

    def test_first_realty_keeps_first_gauge(self) -> None:
        for point in self.points:
            if point.realty_id == 1:
                self.assertEqual(point.gauge_id, 1)

    def test_replaced_realty_switches_gauge(self) -> None:
        replaced = replaced_realties(self.config)
        target = replaced[0]
        replacement = self.replacements[target]
        for point in self.points:
            if point.realty_id != target:
                continue
            if point.reading_date >= replacement.date:
                self.assertEqual(point.gauge_id, replacement.gauge_id)
            else:
                self.assertEqual(point.gauge_id, target)

    def test_all_references_are_positive(self) -> None:
        for point in self.points:
            self.assertEqual(point.user_id, point.realty_id)
            self.assertEqual(point.tariff_id, 1)
            self.assertGreater(point.day_kw, 0)
            self.assertGreater(point.night_kw, 0)


class MonthEndTest(unittest.TestCase):
    def test_month_end_handles_year_rollover(self) -> None:
        self.assertEqual(month_end(2023, 12).isoformat(), "2023-12-31")
        self.assertEqual(month_end(2024, 2).isoformat(), "2024-02-29")


if __name__ == "__main__":
    unittest.main()
