#!/bin/bash
set -e

docker-entrypoint.sh postgres -c wal_level=logical &
echo "connecting to PostGIS..."
while ! pg_isready -h localhost -p 5432 -U postgres; do
    echo "Waiting for PostGIS to be ready..."
    sleep 2
done

/app/migrate.sh

echo "✅ PostGIS migration completed."

sleep infinity
