.DEFAULT_GOAL = run

COMPOSE = docker compose --env-file .env.local

.PHONY: run rund services stop clean gen

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
	set -a && . ./.env.local && set +a && python3 -m gen.main
