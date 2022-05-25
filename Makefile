DB_URL=postgresql://root:password@localhost:5432/simple_bank?sslmode=disable

postgres:
	docker run --name postgres142 -dp 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=password postgres:14.2-alpine

createdb:
	docker exec -it postgres142 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres142 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mockdb:
	mockgen -package mockdb -destination db/mock/store.go github.com/wiliamhw/simplebank/db/sqlc Store

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

.PHONY: postgres sqlc createdb dropdb migrateup migratedown test server mockdb db_docs db_schema