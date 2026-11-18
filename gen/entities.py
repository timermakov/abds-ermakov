from collections.abc import Iterator
from datetime import datetime, timedelta

REALTY_HEADER = ("id", "address", "square", "type", "number", "rooms", "floor", "porch", "created_at")
GAUGE_HEADER = ("id", "name", "country", "created_at")
TARIFF_HEADER = (
    "id",
    "name",
    "price_day_kwph",
    "price_night_kwph",
    "valid_from",
    "valid_to",
    "created_at",
    "updated_at",
)
USER_HEADER = ("id", "fio", "phone", "phone_prev", "created_at", "updated_at")
BINDING_HEADER = ("gauge_id", "realty_id", "created_at", "sign", "version")

FLAT_TYPE = "flat"
NON_RESIDENTIAL_TYPE = "nrcp"
TARIFF_NAME = "День-Ночь"
CITY = "г. Санкт-Петербург"
STREET = "улица Загородная"
GAUGE_NAMES = ("Меркурий 230", "Меркурий 201", "Нева МТ 314", "Энергомера CE308")
GAUGE_COUNTRIES = ("Россия", "Россия", "Беларусь")
SURNAMES = ("Иванов", "Петров", "Сидоров", "Кузнецов", "Смирнов", "Попов", "Васильев", "Соколов", "Михайлов", "Новиков")
NAMES = ("Иван", "Пётр", "Анна", "Мария", "Сергей", "Ольга", "Дмитрий", "Елена")
PATRONYMICS = ("Иванович", "Петрович", "Сергеевич", "Андреевич", "Николаевич", "Алексеевич", "Дмитриевич", "Валерьевна")

DATETIME_LAYOUT = "%Y-%m-%d %H:%M:%S"


def _moment(value: datetime) -> str:
    return value.strftime(DATETIME_LAYOUT)


def replaced_realties(config) -> tuple[int, ...]:
    divisor = max(1, round(1 / config.replacement_ratio))
    ids = tuple(
        realty_id
        for realty_id in range(2, config.realty_count + 1)
        if realty_id % divisor == 1 % divisor
    )
    return ids if ids else (2,)


def realty_rows(config) -> Iterator[tuple]:
    flat_units = config.units_per_building - config.non_residential_units
    for realty_id in range(1, config.realty_count + 1):
        index = realty_id - 1
        building = index // config.units_per_building + 1
        unit = index % config.units_per_building
        if unit < flat_units:
            number = str(unit + 1)
            rooms = unit % 4 + 1
            floor = unit // 4 + 1
            porch = unit % 3 + 1
            square = 30 + unit * 7 % 91
            address = f"{CITY}, {STREET}, дом {building}, кв. {number}"
            realty_type = FLAT_TYPE
        else:
            number = ""
            rooms = 0
            floor = (unit - flat_units) // 2 + 1
            porch = unit % 3 + 1
            square = 50 + unit * 13 % 151
            address = f"{CITY}, {STREET}, дом {building}, пом. {unit - flat_units + 1}"
            realty_type = NON_RESIDENTIAL_TYPE
        created_at = config.dim_created_at + timedelta(minutes=realty_id)
        yield (
            realty_id,
            address,
            square,
            realty_type,
            number,
            rooms,
            floor,
            porch,
            _moment(created_at),
        )


def gauge_rows(config, replaced: tuple[int, ...]) -> Iterator[tuple]:
    for gauge_id in range(1, config.realty_count + 1):
        yield (
            gauge_id,
            GAUGE_NAMES[gauge_id % len(GAUGE_NAMES)],
            GAUGE_COUNTRIES[gauge_id % len(GAUGE_COUNTRIES)],
            _moment(config.dim_created_at + timedelta(minutes=gauge_id)),
        )
    for ordinal in range(1, len(replaced) + 1):
        gauge_id = config.realty_count + ordinal
        yield (
            gauge_id,
            GAUGE_NAMES[gauge_id % len(GAUGE_NAMES)],
            GAUGE_COUNTRIES[gauge_id % len(GAUGE_COUNTRIES)],
            _moment(config.binding_created_at),
        )


def tariff_rows(config) -> Iterator[tuple]:
    created_at = _moment(config.dim_created_at)
    yield (
        1,
        TARIFF_NAME,
        config.tariff_v1_day_price,
        config.tariff_v1_night_price,
        _moment(config.tariff_v1_start),
        _moment(config.tariff_v2_start),
        created_at,
        _moment(config.tariff_v1_start),
    )
    yield (
        1,
        TARIFF_NAME,
        config.tariff_v2_day_price,
        config.tariff_v2_night_price,
        _moment(config.tariff_v2_start),
        _moment(config.tariff_open_end),
        created_at,
        _moment(config.tariff_v2_start),
    )


def user_rows(config) -> Iterator[tuple]:
    for user_id in range(1, config.realty_count + 1):
        fio = f"{SURNAMES[user_id % len(SURNAMES)]} {NAMES[user_id // 3 % len(NAMES)]} {PATRONYMICS[user_id // 7 % len(PATRONYMICS)]}"
        created_at = config.dim_created_at + timedelta(minutes=user_id)
        yield (
            user_id,
            fio,
            f"+7900{user_id % 10000000:07d}",
            "",
            _moment(created_at),
            _moment(created_at),
        )


def binding_rows(config, replacements: dict) -> Iterator[tuple]:
    for realty_id in range(1, config.realty_count + 1):
        yield (realty_id, realty_id, _moment(config.binding_created_at), 1, 1)
    for realty_id, replacement in replacements.items():
        moment = _moment(replacement.date)
        yield (realty_id, realty_id, moment, -1, 1)
        yield (replacement.gauge_id, realty_id, moment, 1, 2)
