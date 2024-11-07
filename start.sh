#!/bin/sh

set -e

#echo "run db migration"
#/app/migrate -path /app/migration -database "postgresql://allen:xzw990609@@rm-cn-g4t3srndq000e4to.rwlb.rds.aliyuncs.com:5432/simple_bank?sslmode=disable" -verbose up

echo "start the app"
exec "$@"