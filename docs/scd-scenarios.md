# SCD-сценарии

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
