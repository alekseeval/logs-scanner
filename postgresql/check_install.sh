#!/bin/bash
# Script for creating DB, installing scheme and DB functions
# Just run it by bash
# Change parameters below if necessary
# You can determine environment variables instead:
#   - HOST              -- host of Postgres server
#   - PORT              -- port of Postgres server
#   - POSTGRES          -- postgres admin user
#   - POSTGRES_PASSWORD -- postgres user password
#   - DB_ADMIN          -- Mind Palace admin user name
#   - DB_ADMIN_PASSWORD -- Mind Palace admin user password
#   - DB_NAME           -- Mind Palace DB name

# Parameters
host='192.168.9.235'
port='5432'
postgres='postgres'
postgres_password='postgres'
db_admin='scanadmin'
db_admin_password='scanadmin'
db_name="tool_db"

if [[ -n "$HOST" ]]; then
  host="$HOST"
fi
if [[ -n "$PORT" ]]; then
  port="$PORT"
fi
if [[ -n "$POSTGRES" ]]; then
  postgres="$POSTGRES"
fi
if [[ -n "$POSTGRES_PASSWORD" ]]; then
  postgres_password="$POSTGRES_PASSWORD"
fi
if [[ -n "$DB_ADMIN" ]]; then
  db_admin="$DB_ADMIN"
fi
if [[ -n "$DB_ADMIN_PASSWORD" ]]; then
  db_admin_password="$DB_ADMIN_PASSWORD"
fi
if [[ -n "$DB_NAME" ]]; then
  db_name="$DB_NAME"
fi

# Going to script directory
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd "$SCRIPT_DIR"

# Set log file of script
mkdir -p "log"
SCRIPT_LOG="$(pwd)/log/log_$(date -Iseconds).log"
exec &>"$SCRIPT_LOG"

# Creating DB and admin user
echo "---> Start DB creation..."
export PGPASSWORD="$postgres_password"

echo "
      CREATE DATABASE $db_name;
      create user $db_admin with encrypted password '$db_admin_password';
      grant all privileges on database $db_name to $db_admin;
      " | psql -h "$host" -p "$port" --user "$postgres" --dbname "postgres" &> /dev/null
echo -e "---> DB $db_name was successfully created\n"

# Installation DB scheme from create_scheme.sql
echo "---> Start DB scheme creation..."
PGPASSWORD="$db_admin_password"
psql -h "$host" -p "$port" --user "$db_admin" --dbname "$db_name" < create_scheme.sql
echo -e "---> DB scheme was successfully created\n"

# Create init DB data by init_data.sql
echo "---> Fill DB with init data..."
PGPASSWORD="$db_admin_password"
psql -h "$host" -p "$port" --user "$db_admin" --dbname "$db_name" < init_data.sql
echo -e "---> Init data was successfully inserted\n"

# Installation DB api functions from /db_api directory
echo "---> Installing DB API..."
cd "db_api/"
for FILE in *; do
  psql -h "$host" -p "$port" --user "$db_admin" --dbname "$db_name" < "$FILE"
done
echo "---> DB API was successfully installed"

exec &>/dev/tty
cat "$SCRIPT_LOG"