.PHONY: migrate-create migrate-up migrate-down

migrate-create:
	migrate create -seq -ext=.sql -dir=./migrations $(name)

migrate-up:
	migrate -path=./migrations -database=$(GREENLIGHT_DB_DSN) up

migrate-down:
	migrate -path=./migrations -database=$(GREENLIGHT_DB_DSN) down $(version)
