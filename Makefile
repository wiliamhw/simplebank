postgres:
	docker run --name postgres142 -dp 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=password postgres:14.2-alpine

createdb:
	docker exec -it postgres142 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres142 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:password@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:password@localhost:5432/simple_bank?sslmode=disable" -verbose down

.PHONY: postgres createdb dropdb migrateup migratedown