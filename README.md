# Лабораторный проект. ЭнергоСбыт

[![Labs](https://github.com/timermakov/abds-ermakov/actions/workflows/energosbyt.lab.yml/badge.svg)](https://github.com/timermakov/abds-ermakov/actions/workflows/energosbyt.lab.yml)

Backend-сервис вымышленной ЖКХ-компании «ЭнергоСбыт» для курса
«Архитектура систем обработки больших данных».
Лабораторная работа №1: SCD 1/2/3 в ClickHouse и HTTP-сервис приёма показаний.

## Запуск

Нужен установленный Docker с Docker Compose и Python 3 в PATH
(`python3`, на Windows достаточно установленного Python). Команды `make`
работают одинаково из Windows (PowerShell, cmd) / Linux (Bash), а генератор и скрипты сами читают `.env.local`.

| Команда | Что делает |
| --- | --- |
| `make run` | Поднимает ClickHouse и приложение через `docker compose` |
| `make stop` | Останавливает контейнеры |
| `make gen` | Генерирует датасеты в `data/` (Python stdlib, без зависимостей) |
| `make load` | Загружает датасеты из `data/` в ClickHouse |
| `make setup` | Создаёт схему, генерирует и загружает данные. Требует запущенного ClickHouse |
| `make teardown` | Удаляет таблицы и данные, база возвращается в исходное состояние |
| `make services` | Показывает статус сервисов |
| `make rund` | Запускает сервисы без detached-режима (для отладки) |
| `make clean` | Останавливает контейнеры и удаляет volumes |

Типичный порядок работы:

```bash
make run
make setup
curl -i http://localhost:8080/ping
make teardown
make stop
```

После `make run` ClickHouse доступен на порту `8123`:

```bash
curl http://localhost:8123/ping
```

## Конфигурация

Вся конфигурация живёт в одном файле — `.env.local`. Его читают Docker Compose,
приложение, генератор и автогрейдер, поэтому проект поднимается и проверяется
«как есть», без правок кода. Захардкоженных значений в коде и compose нет.

| Переменная | Назначение |
| --- | --- |
| `APP_PORT` | Порт HTTP-сервиса |
| `APP_CLICKHOUSE_RETRIES` | Попыток подключения приложения к ClickHouse при старте |
| `APP_CLICKHOUSE_RETRY_DELAY_SEC` | Пауза между попытками |
| `APP_CLICKHOUSE_PING_TIMEOUT_SEC` | Таймаут проверки подключения |
| `CLICKHOUSE_HOST` | Адрес ClickHouse с хоста (в контейнере приложения переопределяется на имя сервиса `clickhouse`) |
| `CLICKHOUSE_PORT` | Нативный порт ClickHouse |
| `CLICKHOUSE_HTTP_PORT` | HTTP-порт ClickHouse |
| `CLICKHOUSE_USER`, `CLICKHOUSE_PASSWORD` | Учётные данные ClickHouse |
| `CLICKHOUSE_DATABASE` | База данных проекта |
| `CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT` | Открывает сетевой доступ пользователю `default` |
| `DATASET_*` | Параметры генератора: объёмы, период, доли корректировок и замен, даты и цены версий тарифа, seed |

## Архитектура

- **HTTP-сервис** — Go 1.26, `net/http` (метод+wildcard маршруты) и `clickhouse-go/v2` по нативному протоколу. Зависимости вендорятся (`vendor/`), Docker-образ собирается без доступа к сети.
- **Генератор датасетов** — Python 3 stdlib (`gen/`), пишет CSV потоково в `data/`.
- **Загрузка** — `clickhouse-client` внутри контейнера, `INSERT ... FORMAT CSVWithNames`.
- **ClickHouse 26.7** — один инстанс, данные в локальной папке `volume/` (не коммитится).

```text
cmd/server/          точка входа HTTP-сервиса
internal/config/     чтение и валидация окружения
internal/httpx/      ответы, ошибки {"message"}, middleware
internal/model/      структуры предметной области
internal/handler/    обработчики по сущностям (users, realty, readings, invoices, tariffs, gauges, ping)
internal/repo/       запросы к ClickHouse по сущностям
gen/                 генератор датасетов и его тесты (unittest)
schema/001_tables.sql  DDL всех таблиц
scripts/             schema.py, load.py, teardown.py — запуск через python, без шелл-скриптов
api/                 Postman-коллекция
```

## HTTP API

Тело запросов и ответов — JSON, поля в `snake_case`, даты `YYYY-MM-DD`, даты со временем `YYYY-MM-DD HH:MM:SS`, ошибки — `{"message": "..."}`. Коллекция Postman с примерами — в [api/energosbyt.postman_collection.json](api/energosbyt.postman_collection.json).

| Метод и путь | Назначение | Успех | Ошибки |
| --- | --- | --- | --- |
| `GET /ping` | Проверка доступности | `200`, пустое тело | — |
| `GET /users/{id}` | Собственник | `200` | `404` |
| `PATCH /users/{id}` | `fio` (SCD 1), `phone` (SCD 3) | `204`, пустое тело | `400`, `404` |
| `GET /realty/{id}` | Объект недвижимости | `200` | `404` |
| `POST /realty/{id}/readings` | Передача показания, повтор за дату — корректировка | `201`, пустое тело | `400`, `404` |
| `GET /realty/{id}/readings?date_from&date_to` | Показания за период, одно актуальное значение за дату | `200` | `400`, `404` |
| `GET /realty/{id}/invoices?limit&offset` | Счета объекта, сортировка по `created_at` по убыванию | `200` | `400`, `404` |
| `GET /tariffs?date=` | Версия тарифа на дату: `valid_from <= date < valid_to` | `200` | `400`, `404` |
| `POST /realty/{id}/gauges` | Установка счётчика с аннулированием активной привязки | `201`, пустое тело | `400`, `404` |

## Схема данных

База `energosbyt`, DDL — в [schema/001_tables.sql](schema/001_tables.sql).

| Таблица | Движок | Ключ сортировки | Партиции | Обоснование |
| --- | --- | --- | --- | --- |
| `dim_realty` | `MergeTree` | `id` | нет | Малый справочник, поиск только по `id` |
| `dim_gauges` | `MergeTree` | `id` | нет | Малый справочник, поиск только по `id` |
| `dim_tariffs` | `ReplacingMergeTree(updated_at)` | `(id, valid_from)` | нет | `valid_from` в ключе: две версии одного тарифа не схлопываются при `FINAL` и остаются историей SCD 2 |
| `dim_users` | `ReplacingMergeTree(updated_at)` | `id` | нет | SCD 1/3 заменяет строку собственника целиком, `FINAL` отдаёт последнюю версию |
| `fact_gauge_bindings` | `VersionedCollapsingMergeTree(sign, version)` | `(realty_id, gauge_id)` | нет | Аннулирование (`sign = -1`, тот же `version`) схлопывается с установкой; `created_at` не входит в ключ, иначе пары не схлопнутся |
| `fact_meter_readings` | `ReplacingMergeTree(created_at)` | `(realty_id, reading_date)` | `toYYYYMM(reading_date)` | Корректировка побеждает по `created_at` внутри ключа «объект + дата»; партиции режут периоды в `GET /readings` |
| `fact_invoice` | `MergeTree` | `(realty_id, created_at)` | `toYYYYMM(created_at)` | Ключ совпадает с порядком выдачи (`ORDER BY created_at DESC` внутри объекта) и пагинацией |

## SCD-сценарии

Каждый сценарий из задания проверялся вручную: сначала запрос к сервису, потом —
что реально лежит в таблицах.

### Собственник сменил фамилию (SCD 1)

Фамилия перезаписывается — старое значение в таблице не сохраняется:

```bash
curl -X PATCH http://localhost:8080/users/1 -d '{"fio": "Петрова Анна Сергеевна"}'
curl http://localhost:8080/users/1
```

```sql
SELECT fio FROM energosbyt.dim_users FINAL WHERE id = 1;
-- fio: Петрова Анна Сергеевна
SELECT count() FROM energosbyt.dim_users FINAL WHERE id = 1 AND fio = 'Петров Иван Иванович';
-- 0: старого значения больше нет нигде
```

### Собственник сменил телефон дважды (SCD 3)

Старый телефон никуда не пропадает — переезжает в `phone_prev`:

```bash
curl -X PATCH http://localhost:8080/users/1 -d '{"phone": "+79111111111"}'
curl -X PATCH http://localhost:8080/users/1 -d '{"phone": "+79222222222"}'
curl http://localhost:8080/users/1
```

```json
{"id": 1, "fio": "Петрова Анна Сергеевна", "phone": "+79222222222", "phone_prev": "+79111111111"}
```

`phone` — последний, `phone_prev` — предыдущий. Если несколько правок попали
в одну секунду, `FINAL` мог бы выбрать любую строку, поэтому приложение пишет
`updated_at` монотонно — каждая новая версия получает `max(now, prev + 1s)`.

### Изменился тариф с 1 ноября (SCD 2)

```bash
curl "http://localhost:8080/tariffs?date=2026-10-27"
curl "http://localhost:8080/tariffs?date=2026-11-01"
```

```sql
SELECT valid_from, valid_to, price_day_kwph, price_night_kwph
FROM energosbyt.dim_tariffs FINAL WHERE id = 1 ORDER BY valid_from;
-- 2023-11-01 00:00:00 | 2026-11-01 00:00:00 | 620 | 310
-- 2026-11-01 00:00:00 | 2099-01-01 00:00:00 | 700 | 350
```

За октябрь сервис возвращает старую цену (620/310), за ноябрь — уже новую (700/350).

### Пересчёт счёта за старый месяц (SCD 2)

Старые счета не пересчитываются: при выставлении цена берётся по версии тарифа,
которая действовала на дату счёта.

```sql
SELECT created_at, day_kw, night_kw, price
FROM energosbyt.fact_invoice
WHERE realty_id = 1 ORDER BY created_at DESC LIMIT 1;
-- 2026-10-27 10:00:00 | 222 | 107 | 170810
```

`170810 = 222 * 620 + 107 * 310` — то есть счёт посчитан по первой версии
тарифа, хотя сейчас действует уже 700/350.

### Корректировка показания (ReplacingMergeTree)

Повторная передача показания за ту же дату — это корректировка, новая строка
затирает старую по `created_at`:

```bash
curl -X POST http://localhost:8080/realty/1/readings -d '{"user_id": 1, "gauge_id": 1, "tariff_id": 1, "reading_date": "2026-10-01", "day_kw": 1250, "night_kw": 300}'
curl "http://localhost:8080/realty/1/readings?date_from=2026-10-01&date_to=2026-10-01"
```

```sql
SELECT count() FROM energosbyt.fact_meter_readings FINAL
WHERE realty_id = 1 AND reading_date = '2026-10-01';
-- 1: победила строка с более поздним created_at; старая пока лежит в таблице и схлопнется только после слияния частей
```

### Замена счётчика (VersionedCollapsingMergeTree)

При установке нового счётчика старая привязка аннулируется парной строкой
с тем же `version`:

```bash
curl -X POST http://localhost:8080/realty/1/gauges -d '{"gauge_id": 2}'
```

```sql
SELECT gauge_id, sign, version FROM energosbyt.fact_gauge_bindings WHERE realty_id = 1;
-- 1 |  1 | 1   установка
-- 1 | -1 | 1   аннулирование с той же версией
-- 2 |  1 | 2   новая установка
SELECT gauge_id FROM energosbyt.fact_gauge_bindings FINAL WHERE realty_id = 1;
-- 2: после слияния активна только новая привязка
```

### Ошибочная строка с лишним нулём (DELETE)

Опечатались при вводе — вставляем показание с лишним нулём и удаляем его
легковесным `DELETE`:

```sql
INSERT INTO energosbyt.fact_meter_readings
(realty_id, user_id, tariff_id, gauge_id, reading_date, day_kw, night_kw, created_at)
VALUES (1, 1, 1, 1, '2026-10-01', 85090000, 40590000, '2026-10-01 12:00:00');
SELECT count() FROM energosbyt.fact_meter_readings WHERE realty_id = 1 AND created_at = '2026-10-01 12:00:00';
-- 1
DELETE FROM energosbyt.fact_meter_readings WHERE realty_id = 1 AND created_at = '2026-10-01 12:00:00';
SELECT count() FROM energosbyt.fact_meter_readings WHERE realty_id = 1 AND created_at = '2026-10-01 12:00:00';
-- 0: строки больше нет
```

## Идемпотентность загрузки

`make setup` можно запускать сколько угодно раз подряд: он каждый раз пересоздаёт
схему с нуля (`DROP DATABASE` → `CREATE DATABASE` → таблицы), так что дубли
не накапливаются.

Если повторить только `make load`, картина разная. В `dim_tariffs`, `dim_users`
и `fact_meter_readings` дублей не будет: те же строки дают те же ключи и версии,
и `ReplacingMergeTree` схлопывает повторы при `FINAL`. А вот `dim_realty`,
`dim_gauges`, `fact_invoice` (обычный `MergeTree`) и `fact_gauge_bindings`
(`VersionedCollapsingMergeTree`) задублируются физически — поэтому между
повторными загрузками лучше сначала сделать `make teardown`.

## Датасеты

Генератор: `make gen`, детерминированный по `DATASET_SEED`, пишет CSV в `data/`
(`data/` и `volume/` в `.gitignore`). Параметры генерации — в `.env.local`.

| Таблица | Строк | Состав |
| --- | --- | --- |
| `dim_realty` | 1 000 | Квартиры и нежилые помещения, 10 домов |
| `dim_gauges` | 1 099 | Счётчик на каждый объект, 99 объектов с заменой |
| `dim_tariffs` | 2 | Две версии тарифа: `2023-11-01 → 2026-11-01` и `2026-11-01 → 2099-01-01` |
| `dim_users` | 1 000 | Собственник на каждый объект |
| `fact_gauge_bindings` | 1 198 | Установка на каждый объект, 99 аннулирований и повторных установок |
| `fact_meter_readings` | ~1 129 000 | Показание на каждый объект за каждый день `2023-11-01 … 2026-10-31`, ~3% корректировок за ту же дату |
| `fact_invoice` | 36 000 | Счёт за каждый месяц периода, `deadline` — 10-е число следующего месяца |

## Проверка решения

1. `git push` ветки и Pull Request в свой репозиторий (`feature/lab-1` → `main`).
2. GitHub Actions запускает проверку лабы автоматически (`.labrc` указывает номер работы).
3. Ревьюер назначается автоматически; после исправления замечаний пуш в ту же ветку перезапускает пайплайн.

## Настройка репозитория

Однократная настройка после создания репозитория из шаблона описана в
[CONTRIBUTING.md](CONTRIBUTING.md): защита ветки `main`, статус-чек «autograder»
и добавление коллабораторов из [CODEOWNERS](CODEOWNERS).
