from collections.abc import Iterator
from dataclasses import dataclass
from datetime import date, datetime, timedelta

INVOICE_HEADER = ("realty_id", "user_id", "tariff_id", "gauge_id", "day_kw", "night_kw", "price", "deadline", "created_at")
DATETIME_LAYOUT = "%Y-%m-%d %H:%M:%S"


@dataclass(frozen=True)
class InvoicePoint:
    realty_id: int
    user_id: int
    tariff_id: int
    gauge_id: int
    day_kw: int
    night_kw: int
    price: int
    deadline: date
    created_at: datetime


def month_end(year: int, month: int) -> date:
    next_month = date(year + month // 12, month % 12 + 1, 1)
    return next_month - timedelta(days=1)


def next_month_first(year: int, month: int) -> date:
    return date(year + month // 12, month % 12 + 1, 1)


def invoice_rows(config, consumption: dict, months: list, replacements: dict) -> Iterator[InvoicePoint]:
    for realty_id in range(1, config.realty_count + 1):
        for index, (year, month) in enumerate(months):
            day_kw, night_kw = consumption[(realty_id, index)]
            created_at = datetime.combine(
                date(year, month, config.invoice_created_day),
                config.invoice_created_time,
            )
            if created_at >= config.tariff_v2_start:
                day_price = config.tariff_v2_day_price
                night_price = config.tariff_v2_night_price
            else:
                day_price = config.tariff_v1_day_price
                night_price = config.tariff_v1_night_price
            period_end = month_end(year, month)
            replacement = replacements.get(realty_id)
            if replacement is not None and period_end >= replacement.date:
                gauge_id = replacement.gauge_id
            else:
                gauge_id = realty_id
            yield InvoicePoint(
                realty_id=realty_id,
                user_id=realty_id,
                tariff_id=1,
                gauge_id=gauge_id,
                day_kw=day_kw,
                night_kw=night_kw,
                price=day_kw * day_price + night_kw * night_price,
                deadline=next_month_first(year, month) + timedelta(days=config.invoice_deadline_day - 1),
                created_at=created_at,
            )


def csv_row(point: InvoicePoint) -> tuple:
    return (
        point.realty_id,
        point.user_id,
        point.tariff_id,
        point.gauge_id,
        point.day_kw,
        point.night_kw,
        point.price,
        point.deadline.isoformat(),
        point.created_at.strftime(DATETIME_LAYOUT),
    )
