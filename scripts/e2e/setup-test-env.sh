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

MYSQL_PORT="${MYSQL_PORT:-13306}"
POSTGRES_PORT="${POSTGRES_PORT:-15432}"

echo "=== MystiSql E2E Test Environment Setup ==="
echo "MySQL Port: ${MYSQL_PORT}"
echo "PostgreSQL Port: ${POSTGRES_PORT}"
echo ""

check_podman() {
    if ! command -v podman &> /dev/null; then
        echo "Error: podman is not installed"
        exit 1
    fi
}

check_port_available() {
    local port=$1
    if netstat -tuln 2>/dev/null | grep -q ":${port} " || \
       ss -tuln 2>/dev/null | grep -q ":${port} "; then
        echo "Warning: Port ${port} is already in use"
        return 1
    fi
    return 0
}

wait_for_mysql() {
    echo "Waiting for MySQL to be ready..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if podman exec "${CONTAINER_NAME_MYSQL}" mysqladmin ping -h localhost -uroot -p"${MYSQL_ROOT_PASSWORD}" &>/dev/null; then
            echo "MySQL is ready!"
            return 0
        fi
        echo "  Attempt ${attempt}/${max_attempts}..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "Error: MySQL failed to start within timeout"
    return 1
}

wait_for_postgres() {
    echo "Waiting for PostgreSQL to be ready..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if podman exec "${CONTAINER_NAME_POSTGRES}" pg_isready -U postgres &>/dev/null; then
            echo "PostgreSQL is ready!"
            return 0
        fi
        echo "  Attempt ${attempt}/${max_attempts}..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "Error: PostgreSQL failed to start within timeout"
    return 1
}

init_mysql_data() {
    echo "Initializing MySQL test data..."
    local init_sql="${PROJECT_ROOT}/test/e2e/init-mysql.sql"
    
    if [ -f "${init_sql}" ]; then
        if podman exec -i "${CONTAINER_NAME_MYSQL}" mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" "${MYSQL_DATABASE}" < "${init_sql}" 2>&1 | grep -i "error" | grep -v "Using a password"; then
            echo "Warning: Some errors occurred during MySQL initialization"
        else
            echo "MySQL test data initialized successfully"
        fi
    else
        echo "Warning: ${init_sql} not found, skipping initialization"
    fi
}

init_postgres_data() {
    echo "Initializing PostgreSQL test data..."
    local init_sql="${PROJECT_ROOT}/test/e2e/init-postgres.sql"
    
    if [ -f "${init_sql}" ]; then
        podman exec -i "${CONTAINER_NAME_POSTGRES}" psql -U postgres -d "${POSTGRES_DB}" < "${init_sql}"
        echo "PostgreSQL test data initialized successfully"
    else
        echo "Warning: ${init_sql} not found, skipping initialization"
    fi
}

check_podman

echo "Step 1: Checking ports..."
check_port_available "${MYSQL_PORT}" || true
check_port_available "${POSTGRES_PORT}" || true

echo ""
echo "Step 2: Removing existing containers (if any)..."
podman rm -f "${CONTAINER_NAME_MYSQL}" 2>/dev/null || true
podman rm -f "${CONTAINER_NAME_POSTGRES}" 2>/dev/null || true

echo ""
echo "Step 3: Starting MySQL 8 container..."
podman run -d \
    --name "${CONTAINER_NAME_MYSQL}" \
    -e MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD}" \
    -e MYSQL_DATABASE="${MYSQL_DATABASE}" \
    -p "${MYSQL_PORT}:3306" \
    --health-cmd="mysqladmin ping -h localhost -uroot -p${MYSQL_ROOT_PASSWORD}" \
    --health-interval=5s \
    --health-timeout=3s \
    --health-retries=5 \
    mysql:8

echo ""
echo "Step 4: Starting PostgreSQL 14 container..."
podman run -d \
    --name "${CONTAINER_NAME_POSTGRES}" \
    -e POSTGRES_PASSWORD="${POSTGRES_PASSWORD}" \
    -e POSTGRES_DB="${POSTGRES_DB}" \
    -p "${POSTGRES_PORT}:5432" \
    --health-cmd="pg_isready -U postgres" \
    --health-interval=5s \
    --health-timeout=3s \
    --health-retries=5 \
    postgres:14

echo ""
echo "Step 5: Waiting for databases to be ready..."
wait_for_mysql
wait_for_postgres

echo ""
echo "Step 6: Initializing test data..."
init_mysql_data
init_postgres_data

echo ""
echo "=== Test Environment Ready ==="
echo ""
echo "MySQL Connection:"
echo "  Host: localhost"
echo "  Port: ${MYSQL_PORT}"
echo "  User: root"
echo "  Password: ${MYSQL_ROOT_PASSWORD}"
echo "  Database: ${MYSQL_DATABASE}"
echo ""
echo "PostgreSQL Connection:"
echo "  Host: localhost"
echo "  Port: ${POSTGRES_PORT}"
echo "  User: postgres"
echo "  Password: ${POSTGRES_PASSWORD}"
echo "  Database: ${POSTGRES_DB}"
echo ""
echo "Container Status:"
podman ps --filter "name=mystisql-test-" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
