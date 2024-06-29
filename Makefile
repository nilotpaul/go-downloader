include ./.env

run: build
	@./bin/go-downloader
	
build:
	@templ fmt ./www
	@templ generate
	@pnpm tailwindcss -i ./www/css/app.css -o ./public/styles.css --minify
	@go build -tags '!dev' -o bin/go-downloader

css:
	@pnpm tailwindcss -i ./www/css/app.css -o ./public/styles.css --watch

templ:
	@templ generate -watch --proxy=http://localhost:$(PORT) 

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