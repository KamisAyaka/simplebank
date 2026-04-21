MOCKGEN := $(shell go env GOPATH)/bin/mockgen
MODULE := $(shell go list -m)
GRPC_GATEWAY_DIR := $(shell go list -m -f '{{.Dir}}' github.com/grpc-ecosystem/grpc-gateway/v2)
DB_URL ?= postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable
EVANS_HOST ?= localhost
EVANS_PORT ?= 9090
EVANS_PACKAGE ?= pb
EVANS_SERVICE ?= SimpleBank

postgres:
	docker run --name postgre15 -p 5433:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:15-alpine

createdb:
	docker exec -it postgre15 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgre15 dropdb simple_bank

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

mock:
	mkdir -p db/mock
	$(MOCKGEN) -source=db/sqlc/store.go -destination=db/mock/store_mock.go -package=mockdb -aux_files $(MODULE)/db/sqlc=db/sqlc/querier.go

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto \
		--proto_path=$(GRPC_GATEWAY_DIR) \
		--go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
		proto/*.proto
		statik -src=./doc/swagger -dest=./doc

evans:
	evans --host $(EVANS_HOST) --port $(EVANS_PORT) -r repl --package $(EVANS_PACKAGE) --service $(EVANS_SERVICE)

.PHONY: createdb postgres dropdb migrateup migratedown sqlc test server mock migratedown1 migrateup1 proto evans
