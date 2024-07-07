include ./.env

run: build
	@./bin/go-downloader
	
build:
	@go build -o bin/go-downloader

test:
	@go test -v ../...

db-status:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_URL) goose -dir=$(migrationPath) status

up:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_URL) goose -dir=$(migrationPath) up

down:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_URL) goose -dir=$(migrationPath) down

reset:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_URL) goose -dir=$(migrationPath) reset