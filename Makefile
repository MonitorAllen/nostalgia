DB_URL=postgresql://root:secret@localhost:5432/nostalgia?sslmode=disable

postgres:
	docker run --name postgres --network nostalgia-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d groonga/pgroonga:3.2.3-alpine-16

createdb:
	docker exec -it postgres createdb --username=root --owner=root nostalgia

dropdb:
	docker exec -it postgres dropdb nostalgia

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

db_dbml:
	dbdocs db2dbml postgres "$(DB_URL)" -o doc/db.dbml

sqlc:
	sqlc generate

test:
	go test -v -cover -short -count=1 ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/MonitorAllen/nostalgia/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go github.com/MonitorAllen/nostalgia/worker TaskDistributor
	mockgen -package mockservice -destination internal/service/mock/redis_service.go github.com/MonitorAllen/nostalgia/internal/service Redis

swag:
	rm -f doc/swagger/*.swagger.json
	swag init -o ./doc/swagger --instanceName nostalgia
	statik -src=./doc/swagger -dest=./doc

proto:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
        --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
        --grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
        proto/*.proto

evans:
	evans --host localhost --port 9090 -r repl

redis:
	docker run --name redis -p 6379:6379 -d redis:7-alpine

server_docker_up:
	docker start postgres
	docker start redis
	go run main.go

decrypt_env:
	gpg --batch --yes --passphrase "${ENV_PASSPHRASE}" --output .env --decrypt .env.$(env).enc

encrypt_env:
	gpg --batch --yes --symmetric --cipher-algo AES256 --output .env.$(env).enc .env$(suf)

.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 new_migration db_docs db_schema sqlc test server mock proto evans redis db_dbml decrypt_env encrypt_env
