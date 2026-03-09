#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

CONTAINER_NAME_MYSQL="mystisql-test-mysql"
CONTAINER_NAME_POSTGRES="mystisql-test-postgres"

MYSQL_ROOT_PASSWORD="test123456"
MYSQL_DATABASE="mystisql_test"
POSTGRES_PASSWORD="test123456"
POSTGRES_DB="mystisql_test"

echo "=== MystiSql E2E Test Database Reset ==="
echo ""

reset_mysql() {
    echo "Resetting MySQL database..."
    
    if ! podman ps --filter "name=${CONTAINER_NAME_MYSQL}" --filter "status=running" --format "{{.Names}}" | grep -q "${CONTAINER_NAME_MYSQL}"; then
        echo "Error: MySQL container is not running"
        echo "Please start the test environment first: scripts/e2e/setup-test-env.sh"
        return 1
    fi
    
    echo "  Dropping and recreating database..."
    podman exec "${CONTAINER_NAME_MYSQL}" mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "DROP DATABASE IF EXISTS ${MYSQL_DATABASE};"
    podman exec "${CONTAINER_NAME_MYSQL}" mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" -e "CREATE DATABASE ${MYSQL_DATABASE};"
    
    local init_sql="${PROJECT_ROOT}/test/e2e/init-mysql.sql"
    if [ -f "${init_sql}" ]; then
        echo "  Re-initializing test data..."
        podman exec -i "${CONTAINER_NAME_MYSQL}" mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" "${MYSQL_DATABASE}" < "${init_sql}"
        echo "  ✓ MySQL database reset complete"
    else
        echo "  Warning: ${init_sql} not found, database created but not initialized"
    fi
}

reset_postgres() {
    echo "Resetting PostgreSQL database..."
    
    if ! podman ps --filter "name=${CONTAINER_NAME_POSTGRES}" --filter "status=running" --format "{{.Names}}" | grep -q "${CONTAINER_NAME_POSTGRES}"; then
        echo "Error: PostgreSQL container is not running"
        echo "Please start the test environment first: scripts/e2e/setup-test-env.sh"
        return 1
    fi
    
    echo "  Dropping and recreating database..."
    podman exec "${CONTAINER_NAME_POSTGRES}" psql -U postgres -c "DROP DATABASE IF EXISTS ${POSTGRES_DB};" || true
    podman exec "${CONTAINER_NAME_POSTGRES}" psql -U postgres -c "CREATE DATABASE ${POSTGRES_DB};"
    
    local init_sql="${PROJECT_ROOT}/test/e2e/init-postgres.sql"
    if [ -f "${init_sql}" ]; then
        echo "  Re-initializing test data..."
        podman exec -i "${CONTAINER_NAME_POSTGRES}" psql -U postgres -d "${POSTGRES_DB}" < "${init_sql}"
        echo "  ✓ PostgreSQL database reset complete"
    else
        echo "  Warning: ${init_sql} not found, database created but not initialized"
    fi
}

DB_TYPE="${1:-all}"

case "${DB_TYPE}" in
    mysql)
        reset_mysql
        ;;
    postgres|postgresql)
        reset_postgres
        ;;
    all)
        reset_mysql
        echo ""
        reset_postgres
        ;;
    *)
        echo "Usage: $0 [mysql|postgres|all]"
        echo "  Default: all"
        exit 1
        ;;
esac

echo ""
echo "=== Database Reset Complete ==="
