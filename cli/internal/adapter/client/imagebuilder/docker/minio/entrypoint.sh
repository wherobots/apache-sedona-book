#!/bin/bash
set -e

/usr/bin/docker-entrypoint.sh server /data --console-address ":9001" &

until mc ready local > /dev/null; do
  echo "Waiting for minio..."
  sleep 2
done

echo "✅ minio is reachable! open the MinIO console at http://localhost:9001"
echo "USER: $MINIO_ROOT_USER"
echo "PASSWORD: $MINIO_ROOT_PASSWORD"

# creating buckets
echo "Creating bucket apache-sedona-book and uploading sources..."
mc alias set sedona http://localhost:9000 $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD
mc mb sedona/apache-sedona-book-local || true
mc cp /app/sources/ sedona/apache-sedona-book-local/ --recursive || true

sleep infinity
