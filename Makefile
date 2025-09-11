DATABASE_DSN=postgres://gophkeeper-user:123123123@localhost:5436/postgres?sslmode=disable

CMD_DIR=cmd/gophkeeper
BINARY=gophkeeper

db-up:
	docker compose up -d postgres
db-down:
	docker compose down --remove-orphans
sqlc:
	rm -rf internal/storage/repos/sqlc/sqlcgen/* && docker compose run --rm sqlc

app-up:	db-up
	go run ./cmd/gophkeeper/main.go -s secret -d $(DATABASE_DSN)


app-build:
	cd $(CMD_DIR) && go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$$(date +%Y-%m-%d) -X main.buildCommit=$$(git rev-parse --short HEAD)" -o $(BINARY) *.go

# Создать миграцию
migrate-create:
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		echo "expecting migration name. Format: make migrate-create <name>"; \
		exit 1; \
	fi
	migrate create -ext sql -dir internal/db/migrations -seq $(filter-out $@,$(MAKECMDGOALS))

%:
	@:


# Миграция вверх
migrate-up:
	migrate -database $(DATABASE_DSN) -path ./internal/db/migrations up

# Миграция вниз
migrate-down:
	migrate -database $(DATABASE_DSN) -path ./internal/db/migrations down 1
