from collections.abc import Iterator
from dataclasses import dataclass
from datetime import date, datetime, timedelta
import random

READING_HEADER = ("realty_id", "user_id", "tariff_id", "gauge_id", "reading_date", "day_kw", "night_kw", "created_at")
DATETIME_LAYOUT = "%Y-%m-%d %H:%M:%S"
TARIFF_ID = 1


@dataclass(frozen=True)
class Replacement:
    gauge_id: int
    date: date


@dataclass(frozen=True)
class ReadingPoint:
    realty_id: int
    user_id: int
    tariff_id: int
    gauge_id: int
    reading_date: date
    day_kw: int
    night_kw: int
    created_at: datetime
    day_delta: int
    night_delta: int
    corrected: bool


def period_days(config) -> list[date]:
    days = []
    current = config.readings_start
    while current <= config.readings_end:
        days.append(current)
        current += timedelta(days=1)
    return days


def months(config) -> list[tuple[int, int]]:
    result = []
    year, month = config.readings_start.year, config.readings_start.month
    while (year, month) <= (config.readings_end.year, config.readings_end.month):
        result.append((year, month))
        month += 1
        if month > 12:
            month = 1
            year += 1
    return result


def month_index(day: date, start: date) -> int:
    return (day.year - start.year) * 12 + day.month - start.month


def replacements(config, replaced: tuple[int, ...]) -> dict[int, Replacement]:
    period = len(period_days(config))
    offset = int((period - 1) * config.replacement_date_fraction)
    base = config.readings_start + timedelta(days=offset)
    return {
        realty_id: Replacement(
            gauge_id=config.realty_count + ordinal,
            date=base,
        )
        for ordinal, realty_id in enumerate(replaced, start=1)
    }


def active_gauge(realty_id: int, day: date, replacements: dict) -> int:
    replacement = replacements.get(realty_id)
    if replacement is not None and day >= replacement.date:
        return replacement.gauge_id
    return realty_id


def points(config, rng: random.Random, replacements: dict) -> Iterator[ReadingPoint]:
    days = period_days(config)
    for realty_id in range(1, config.realty_count + 1):
        day_kw = rng.randint(config.counter_start_min, config.counter_start_max)
        night_kw = rng.randint(config.counter_start_min, config.counter_start_max)
        for day in days:
            day_delta = rng.randint(config.day_kw_delta_min, config.day_kw_delta_max)
            night_delta = rng.randint(config.night_kw_delta_min, config.night_kw_delta_max)
            day_kw += day_delta
            night_kw += night_delta
            gauge_id = active_gauge(realty_id, day, replacements)
            yield ReadingPoint(
                realty_id=realty_id,
                user_id=realty_id,
                tariff_id=TARIFF_ID,
                gauge_id=gauge_id,
                reading_date=day,
                day_kw=day_kw,
                night_kw=night_kw,
                created_at=datetime.combine(day, config.reading_created_time),
                day_delta=day_delta,
                night_delta=night_delta,
                corrected=False,
            )
            if rng.random() < config.correction_ratio:
                day_kw += rng.randint(config.correction_increment_min, config.correction_increment_max)
                night_kw += rng.randint(config.correction_increment_min, config.correction_increment_max)
                yield ReadingPoint(
                    realty_id=realty_id,
                    user_id=realty_id,
                    tariff_id=TARIFF_ID,
                    gauge_id=gauge_id,
                    reading_date=day,
                    day_kw=day_kw,
                    night_kw=night_kw,
                    created_at=datetime.combine(day, config.correction_created_time),
                    day_delta=0,
                    night_delta=0,
                    corrected=True,
                )


def csv_row(point: ReadingPoint) -> tuple:
    return (
        point.realty_id,
        point.user_id,
        point.tariff_id,
        point.gauge_id,
        point.reading_date.isoformat(),
        point.day_kw,
        point.night_kw,
        point.created_at.strftime(DATETIME_LAYOUT),
    )
