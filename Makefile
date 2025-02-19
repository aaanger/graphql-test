build:
	build/docker-compose.yml build
run:
	docker compose up
migrate:
	goose -dir pkg/db/migrations postgres "host=${PSQL_HOST} port=${PSQL_PORT} user=${PSQL_USERNAME} password=${PSQL_PASSWORD} dbname=${PSQL_DBNAME} sslmode=disable" up
rollback:
	goose -dir pkg/db/migrations postgres "host=${PSQL_HOST} port=${PSQL_PORT} user=${PSQL_USERNAME} password=${PSQL_PASSWORD} dbname=${PSQL_DBNAME} sslmode=disable" down
test:
	go test -v ./...