.DEFAULT_GOAL = run

COMPOSE = docker compose --env-file .env.local
PYTHON = python3

.PHONY: run rund services stop clean gen load setup teardown

run:
	$(COMPOSE) up -d --build

rund:
	$(COMPOSE) up --build

services:
	$(COMPOSE) ps

stop:
	$(COMPOSE) down

clean:
	$(COMPOSE) down -v

gen:
	$(PYTHON) -m gen.main

load:
	$(PYTHON) scripts/load.py

setup: gen
	$(PYTHON) scripts/schema.py
	$(PYTHON) scripts/load.py

teardown:
	$(PYTHON) scripts/teardown.py
