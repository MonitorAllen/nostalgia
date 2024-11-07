#DB_URL=postgresql://allen:xzw990609@@rm-cn-g4t3srndq000e4to.rwlb.rds.aliyuncs.com:5432/simple_bank?sslmode=disable
DB_URL=postgresql://root:root@localhost:5432/nostalgia?sslmode=disable

postgres:
	docker run --name postgres12 --network nostalgia-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=root -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root nostalgia

dropdb:
	docker exec -it postgres12 dropdb nostalgia

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

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/MonitorAllen/nostalgia/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go github.com/MonitorAllen/nostalgia/worker TaskDistributor

swag:
	rm -f doc/swagger/*.swagger.json
	swag init -o ./doc/swagger --instanceName nostalgia
	statik -src=./doc/swagger -dest=./doc

evans:
	evans --host localhost --port 9090 -r repl

redis:
	docker run --name redis -p 6379:6379 -d redis:7-alpine

.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 new_migration db_docs db_schema sqlc test server mock proto evans redis
