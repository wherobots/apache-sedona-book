#!/bin/bash
set -e

/usr/bin/dumb-init /entrypoint standalone &

# wait for Airflow to start
curl --retry 10 --retry-delay 3 --silent --retry-all-errors http://localhost:8080/health

echo "Initializing Airflow database..."
password=$(cat standalone_admin_password.txt)

echo "Airflow admin user: admin with password: $password"

sed -i -e "s|<AWS_ACCESS_KEY_ID>|$AWS_ACCESS_KEY_ID|g" -e "s|<AWS_SECRET_ACCESS_KEY>|$AWS_SECRET_ACCESS_KEY|g" /home/airflow/.local/lib/python3.12/site-packages/pyspark/conf/spark-defaults.conf

airflow connections add spark \
    --conn-type spark \
    --conn-host local[*] || true

# keep the container running
sleep infinity
