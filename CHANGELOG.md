# Changelog

Все значимые изменения в этом проекте будут документированы в этом файле.

Формат основан на [Keep a Changelog](https://keepachangelog.com/ru/1.0.0/),
и этот проект придерживается [Semantic Versioning](https://semver.org/lang/ru/).

## [Unreleased]

### Добавлено

- Лабораторная работа №1: SCD 1/2/3 в ClickHouse и HTTP-сервис приёма показаний
- `docker-compose.yml` с ClickHouse 26.7 и приложением на Go, healthcheck и ожиданием готовности БД
- `Dockerfile` с multi-stage сборкой; зависимости Go вендорятся, сборка образа не требует сети
- Схема данных `schema/001_tables.sql`: 7 таблиц с движками `MergeTree`, `ReplacingMergeTree`, `VersionedCollapsingMergeTree`
- Генератор датасетов на Python stdlib (`gen/`) с тестами, `make gen` пишет CSV в `data/`
- Команды `make gen`, `make load`, `make setup`, `make teardown` и скрипты загрузки (`scripts/`)
- HTTP-сервис: `ping`, собственники (SCD 1/3), объекты, показания с корректировками, счета с пагинацией, тарифы на дату (SCD 2), установка счётчиков с аннулированием
- Unit-тесты Go (`go test ./...`) и генератора (`python3 -m unittest`)
- Postman-коллекция `api/energosbyt.postman_collection.json`
- Документация: описание запуска, обоснование ключей сортировки и партиций, SQL-сценарии SCD в `README.md`

## [v1.0.0] - 2026-10-30

### Добавлено

- CI-воркфлоу `Energosbyt Labs` для автоматической проверки лабораторных работ
(`.github/workflows/energosbyt.lab.yml`) и скрипт
[определения номера лабораторной работы из `.labrc`](.github/scripts/lab_number.sh)
- Файл `.gitignore`
- Файл `.labrc` с номером проверяемой лабораторной работы
- Файл `.env.local` с переменными окружения для Docker Compose
- Файл `CODEOWNERS` с назначением владельца кода
- `Makefile` с командами управления сервисами: `run`, `rund`, `services`, `stop`, `clean`
- `CONTRIBUTING.md` с требованиями к структуре проекта, процессу разработки и проверки
- Инструкция по настройке репозитория в `README.md`: защита мастер-ветки, добавление коллабораторов
