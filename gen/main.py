import os
import random
import sys
from datetime import date

from gen.config import DatasetConfig
from gen.entities import (
    BINDING_HEADER,
    GAUGE_HEADER,
    REALTY_HEADER,
    TARIFF_HEADER,
    USER_HEADER,
    binding_rows,
    gauge_rows,
    realty_rows,
    replaced_realties,
    tariff_rows,
    user_rows,
)
from gen.invoices import INVOICE_HEADER, csv_row as invoice_csv_row, invoice_rows
from gen.readings import (
    READING_HEADER,
    csv_row as reading_csv_row,
    month_index,
    months,
    points,
    replacements,
)
from gen.writer import CsvWriter


def _month_aggregates(config, rng, replacements, months_list) -> dict:
    consumption = {}
    total = 0
    with CsvWriter(os.path.join(config.output_dir, "fact_meter_readings.csv"), READING_HEADER) as writer:
        for point in points(config, rng, replacements):
            writer.write(reading_csv_row(point))
            if not point.corrected:
                key = (point.realty_id, month_index(point.reading_date, config.readings_start))
                day_kw, night_kw = consumption.get(key, (0, 0))
                consumption[key] = (day_kw + point.day_delta, night_kw + point.night_delta)
            total += 1
    print(f"fact_meter_readings: {total}")
    return consumption


def generate(config) -> None:
    os.makedirs(config.output_dir, exist_ok=True)
    rng = random.Random(config.seed)
    replaced = replaced_realties(config)
    replacements_map = replacements(config, replaced)
    months_list = months(config)

    with CsvWriter(os.path.join(config.output_dir, "dim_realty.csv"), REALTY_HEADER) as writer:
        for row in realty_rows(config):
            writer.write(row)
    print(f"dim_realty: {writer.rows}")

    with CsvWriter(os.path.join(config.output_dir, "dim_gauges.csv"), GAUGE_HEADER) as writer:
        for row in gauge_rows(config, replaced):
            writer.write(row)
    print(f"dim_gauges: {writer.rows}")

    with CsvWriter(os.path.join(config.output_dir, "dim_tariffs.csv"), TARIFF_HEADER) as writer:
        for row in tariff_rows(config):
            writer.write(row)
    print(f"dim_tariffs: {writer.rows}")

    with CsvWriter(os.path.join(config.output_dir, "dim_users.csv"), USER_HEADER) as writer:
        for row in user_rows(config):
            writer.write(row)
    print(f"dim_users: {writer.rows}")

    with CsvWriter(os.path.join(config.output_dir, "fact_gauge_bindings.csv"), BINDING_HEADER) as writer:
        for row in binding_rows(config, replacements_map):
            writer.write(row)
    print(f"fact_gauge_bindings: {writer.rows}")

    consumption = _month_aggregates(config, rng, replacements_map, months_list)

    with CsvWriter(os.path.join(config.output_dir, "fact_invoice.csv"), INVOICE_HEADER) as writer:
        for row in invoice_rows(config, consumption, months_list, replacements_map):
            writer.write(invoice_csv_row(row))
    print(f"fact_invoice: {writer.rows}")


def main() -> None:
    try:
        config = DatasetConfig.from_env()
    except ValueError as error:
        sys.exit(f"config error: {error}")
    generate(config)


if __name__ == "__main__":
    main()
