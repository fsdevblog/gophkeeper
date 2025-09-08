db-up:
	docker compose up -d postgres
db-down:
	docker compose down --remove-orphans
sqlc:
	rm -rf internal/storage/repos/sqlc/sqlcgen/* && docker compose run --rm sqlc