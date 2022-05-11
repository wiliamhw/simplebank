#!/bin/sh

set -e

echo "run db migration"
/app/migrate -path /app/db/migration -database "$DB_CONNECTION://$DB_USERNAME:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_DATABASE?sslmode=disable" -verbose up

echo "start the app"
exec "$@" # Run all arguments
