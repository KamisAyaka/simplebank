MOCKGEN := $(shell go env GOPATH)/bin/mockgen
MODULE := $(shell go list -m)

postgres:
	docker run --name postgre15 -p 5433:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:15-alpine

createdb:
	docker exec -it postgre15 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgre15 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mkdir -p db/mock
	$(MOCKGEN) -source=db/sqlc/store.go -destination=db/mock/store_mock.go -package=mockdb -aux_files $(MODULE)/db/sqlc=db/sqlc/querier.go

.PHONY: createdb postgres dropdb migrateup migratedown sqlc test server mock migratedown1 migrateup1
