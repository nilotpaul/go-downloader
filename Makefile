include ./.env

MIGRATION_PATH = migrations

run: build
	@ENVIRONMENT=PROD ./bin/go-downloader
	
build:
	@ENVIRONMENT=PROD go build -tags '!dev' -o bin/go-downloader

test:
	@go test -v ../...

db-status:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_URL) goose -dir=$(MIGRATION_PATH) status

up:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_URL) goose -dir=$(MIGRATION_PATH) up

down:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_URL) goose -dir=$(MIGRATION_PATH) down

reset:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_URL) goose -dir=$(MIGRATION_PATH) reset