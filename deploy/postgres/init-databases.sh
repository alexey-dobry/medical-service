#!/bin/sh
# Creates one database and one owner role per service, so auth_service and
# user_service cannot read each other's tables. Runs once, on an empty volume.
set -eu

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<SQL
CREATE ROLE auth_service LOGIN PASSWORD '${AUTH_DB_PASSWORD}';
CREATE DATABASE auth OWNER auth_service;
CREATE ROLE user_service LOGIN PASSWORD '${USER_DB_PASSWORD}';
CREATE DATABASE users OWNER user_service;
SQL
