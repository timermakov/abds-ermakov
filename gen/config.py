import os
from dataclasses import dataclass
from datetime import date, datetime, time
from pathlib import Path

DATE_FORMAT = "%Y-%m-%d"
DATETIME_FORMAT = "%Y-%m-%d %H:%M:%S"
TIME_FORMAT = "%H:%M:%S"
OUTPUT_DIR = "data"
ENV_FILE = Path(__file__).resolve().parent.parent / ".env.local"


def _unquote(value: str) -> str:
    if len(value) >= 2 and value[0] == value[-1] and value[0] in ("'", '"'):
        return value[1:-1]
    return value


def load_env_file(path: Path = ENV_FILE) -> None:
    if not path.exists():
        return
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        os.environ.setdefault(key.strip(), _unquote(value.strip()))


class ConfigError(ValueError):
    pass


def _required(name: str) -> str:
    value = os.environ.get(name)
    if value is None or value == "":
        raise ConfigError(f"environment variable {name} is not set")
    return value


def _integer(name: str) -> int:
    try:
        return int(_required(name))
    except ValueError:
        raise ConfigError(f"environment variable {name} must be an integer") from None


def _decimal(name: str) -> float:
    try:
        return float(_required(name))
    except ValueError:
        raise ConfigError(f"environment variable {name} must be a number") from None


def _day(name: str) -> date:
    try:
        return datetime.strptime(_required(name), DATE_FORMAT).date()
    except ValueError:
        raise ConfigError(f"environment variable {name} must be a date in {DATE_FORMAT} format") from None


def _moment(name: str) -> datetime:
    try:
        return datetime.strptime(_required(name), DATETIME_FORMAT)
    except ValueError:
        raise ConfigError(f"environment variable {name} must be a datetime in {DATETIME_FORMAT} format") from None


def _time_of_day(name: str) -> time:
    try:
        return datetime.strptime(_required(name), TIME_FORMAT).time()
    except ValueError:
        raise ConfigError(f"environment variable {name} must be a time in {TIME_FORMAT} format") from None


@dataclass(frozen=True)
class DatasetConfig:
    seed: int
    realty_count: int
    units_per_building: int
    non_residential_units: int
    readings_start: date
    readings_end: date
    correction_ratio: float
    replacement_ratio: float
    replacement_date_fraction: float
    tariff_v1_start: datetime
    tariff_v2_start: datetime
    tariff_open_end: datetime
    tariff_v1_day_price: int
    tariff_v1_night_price: int
    tariff_v2_day_price: int
    tariff_v2_night_price: int
    dim_created_at: datetime
    binding_created_at: datetime
    day_kw_delta_min: int
    day_kw_delta_max: int
    night_kw_delta_min: int
    night_kw_delta_max: int
    counter_start_min: int
    counter_start_max: int
    correction_increment_min: int
    correction_increment_max: int
    reading_created_time: time
    correction_created_time: time
    invoice_created_day: int
    invoice_created_time: time
    invoice_deadline_day: int
    output_dir: str

    @staticmethod
    def from_env() -> "DatasetConfig":
        load_env_file()
        config = DatasetConfig(
            seed=_integer("DATASET_SEED"),
            realty_count=_integer("DATASET_REALTY_COUNT"),
            units_per_building=_integer("DATASET_UNITS_PER_BUILDING"),
            non_residential_units=_integer("DATASET_NON_RESIDENTIAL_UNITS"),
            readings_start=_day("DATASET_READINGS_START_DATE"),
            readings_end=_day("DATASET_READINGS_END_DATE"),
            correction_ratio=_decimal("DATASET_READING_CORRECTION_RATIO"),
            replacement_ratio=_decimal("DATASET_GAUGE_REPLACEMENT_RATIO"),
            replacement_date_fraction=_decimal("DATASET_REPLACEMENT_DATE_FRACTION"),
            tariff_v1_start=_moment("DATASET_TARIFF_V1_START"),
            tariff_v2_start=_moment("DATASET_TARIFF_V2_START"),
            tariff_open_end=_moment("DATASET_TARIFF_OPEN_END"),
            tariff_v1_day_price=_integer("DATASET_TARIFF_V1_DAY_PRICE"),
            tariff_v1_night_price=_integer("DATASET_TARIFF_V1_NIGHT_PRICE"),
            tariff_v2_day_price=_integer("DATASET_TARIFF_V2_DAY_PRICE"),
            tariff_v2_night_price=_integer("DATASET_TARIFF_V2_NIGHT_PRICE"),
            dim_created_at=_moment("DATASET_DIM_CREATED_AT"),
            binding_created_at=_moment("DATASET_BINDING_CREATED_AT"),
            day_kw_delta_min=_integer("DATASET_DAY_KW_DELTA_MIN"),
            day_kw_delta_max=_integer("DATASET_DAY_KW_DELTA_MAX"),
            night_kw_delta_min=_integer("DATASET_NIGHT_KW_DELTA_MIN"),
            night_kw_delta_max=_integer("DATASET_NIGHT_KW_DELTA_MAX"),
            counter_start_min=_integer("DATASET_COUNTER_START_MIN"),
            counter_start_max=_integer("DATASET_COUNTER_START_MAX"),
            correction_increment_min=_integer("DATASET_CORRECTION_INCREMENT_MIN"),
            correction_increment_max=_integer("DATASET_CORRECTION_INCREMENT_MAX"),
            reading_created_time=_time_of_day("DATASET_READING_CREATED_TIME"),
            correction_created_time=_time_of_day("DATASET_CORRECTION_CREATED_TIME"),
            invoice_created_day=_integer("DATASET_INVOICE_CREATED_DAY"),
            invoice_created_time=_time_of_day("DATASET_INVOICE_CREATED_TIME"),
            invoice_deadline_day=_integer("DATASET_INVOICE_DEADLINE_DAY"),
            output_dir=OUTPUT_DIR,
        )
        config.validate()
        return config

    def validate(self) -> None:
        if self.realty_count <= 0:
            raise ConfigError("DATASET_REALTY_COUNT must be positive")
        if self.non_residential_units < 0 or self.non_residential_units >= self.units_per_building:
            raise ConfigError("DATASET_NON_RESIDENTIAL_UNITS must be less than DATASET_UNITS_PER_BUILDING")
        if self.readings_start > self.readings_end:
            raise ConfigError("DATASET_READINGS_START_DATE must not be after DATASET_READINGS_END_DATE")
        if self.correction_ratio < 0 or self.correction_ratio > 1:
            raise ConfigError("DATASET_READING_CORRECTION_RATIO must be between 0 and 1")
        if self.replacement_ratio <= 0 or self.replacement_ratio > 1:
            raise ConfigError("DATASET_GAUGE_REPLACEMENT_RATIO must be between 0 and 1")
        if self.replacement_date_fraction < 0 or self.replacement_date_fraction > 1:
            raise ConfigError("DATASET_REPLACEMENT_DATE_FRACTION must be between 0 and 1")
        if self.tariff_v1_start >= self.tariff_v2_start or self.tariff_v2_start >= self.tariff_open_end:
            raise ConfigError("tariff versions must follow the order v1 start, v2 start, open end")
        if min(
            self.tariff_v1_day_price,
            self.tariff_v1_night_price,
            self.tariff_v2_day_price,
            self.tariff_v2_night_price,
        ) <= 0:
            raise ConfigError("tariff prices must be positive")
        if self.tariff_v1_start.date() > self.readings_start:
            raise ConfigError("the first tariff version must start before the readings period")
        if self.tariff_open_end.date() <= self.readings_end:
            raise ConfigError("the open tariff version must outlast the readings period")
        if self.binding_created_at.date() > self.readings_start:
            raise ConfigError("gauge bindings must be created before the readings period")
        if self.dim_created_at > self.binding_created_at:
            raise ConfigError("dimension rows must be created before gauge bindings")
        if min(self.day_kw_delta_min, self.night_kw_delta_min) < 0:
            raise ConfigError("consumption deltas must not be negative")
        if self.day_kw_delta_max <= self.day_kw_delta_min or self.night_kw_delta_max <= self.night_kw_delta_min:
            raise ConfigError("consumption delta maximum must be greater than the minimum")
        if self.counter_start_min < 0 or self.counter_start_max <= self.counter_start_min:
            raise ConfigError("counter start bounds must satisfy 0 <= min < max")
        if self.correction_increment_min <= 0 or self.correction_increment_max <= self.correction_increment_min:
            raise ConfigError("correction increment bounds must satisfy 0 < min < max")
        if self.invoice_created_day < 1 or self.invoice_created_day > 28:
            raise ConfigError("DATASET_INVOICE_CREATED_DAY must be a day from 1 to 28")
        if self.invoice_deadline_day < 1 or self.invoice_deadline_day > 28:
            raise ConfigError("DATASET_INVOICE_DEADLINE_DAY must be a day from 1 to 28")
        if self.output_dir == "":
            raise ConfigError("output directory must not be empty")
