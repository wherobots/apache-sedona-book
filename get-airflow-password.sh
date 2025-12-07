#!/bin/bash
set -e

password=$(docker exec -t airflow cat standalone_admin_password.txt)
echo "Airflow admin user: admin with password: $password"

