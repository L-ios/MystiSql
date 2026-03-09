#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

CONTAINER_NAME_MYSQL="mystisql-test-mysql"
CONTAINER_NAME_POSTGRES="mystisql-test-postgres"

MYSQL_PORT="${MYSQL_PORT:-13306}"
POSTGRES_PORT="${POSTGRES_PORT:-15432}"

echo "=== MystiSql E2E Test Environment Check ==="
echo ""

ERRORS=0

echo "1. Checking Podman..."
if command -v podman &> /dev/null; then
    PODMAN_VERSION=$(podman --version)
    echo "   ✓ Podman installed: ${PODMAN_VERSION}"
else
    echo "   ✗ Podman not installed"
    ERRORS=$((ERRORS + 1))
fi

echo ""
echo "2. Checking container images..."
if podman images mysql:8 --format "{{.Repository}}:{{.Tag}}" | grep -q "mysql:8"; then
    echo "   ✓ MySQL 8 image available"
else
    echo "   ✗ MySQL 8 image not found (run: podman pull mysql:8)"
    ERRORS=$((ERRORS + 1))
fi

if podman images postgres:14 --format "{{.Repository}}:{{.Tag}}" | grep -q "postgres:14"; then
    echo "   ✓ PostgreSQL 14 image available"
else
    echo "   ✗ PostgreSQL 14 image not found (run: podman pull postgres:14)"
    ERRORS=$((ERRORS + 1))
fi

echo ""
echo "3. Checking container status..."
MYSQL_RUNNING=$(podman ps --filter "name=${CONTAINER_NAME_MYSQL}" --filter "status=running" --format "{{.Names}}" 2>/dev/null)
POSTGRES_RUNNING=$(podman ps --filter "name=${CONTAINER_NAME_POSTGRES}" --filter "status=running" --format "{{.Names}}" 2>/dev/null)

if [ -n "${MYSQL_RUNNING}" ]; then
    echo "   ✓ MySQL container is running"
else
    echo "   ✗ MySQL container is not running"
    ERRORS=$((ERRORS + 1))
fi

if [ -n "${POSTGRES_RUNNING}" ]; then
    echo "   ✓ PostgreSQL container is running"
else
    echo "   ✗ PostgreSQL container is not running"
    ERRORS=$((ERRORS + 1))
fi

echo ""
echo "4. Checking database connectivity..."
if command -v mysql &> /dev/null && [ -n "${MYSQL_RUNNING}" ]; then
    if mysql -h 127.0.0.1 -P "${MYSQL_PORT}" -uroot -ptest123456 -e "SELECT 1" &>/dev/null; then
        echo "   ✓ MySQL connection successful"
    else
        echo "   ✗ MySQL connection failed"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo "   - MySQL client not available, skipping connection test"
fi

if command -v psql &> /dev/null && [ -n "${POSTGRES_RUNNING}" ]; then
    if PGPASSWORD=test123456 psql -h 127.0.0.1 -p "${POSTGRES_PORT}" -U postgres -d mystisql_test -c "SELECT 1" &>/dev/null; then
        echo "   ✓ PostgreSQL connection successful"
    else
        echo "   ✗ PostgreSQL connection failed"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo "   - PostgreSQL client not available, skipping connection test"
fi

echo ""
echo "5. Checking ports..."
if netstat -tuln 2>/dev/null | grep -q ":${MYSQL_PORT} " || ss -tuln 2>/dev/null | grep -q ":${MYSQL_PORT} "; then
    echo "   ✓ MySQL port ${MYSQL_PORT} is listening"
else
    echo "   ✗ MySQL port ${MYSQL_PORT} is not listening"
    ERRORS=$((ERRORS + 1))
fi

if netstat -tuln 2>/dev/null | grep -q ":${POSTGRES_PORT} " || ss -tuln 2>/dev/null | grep -q ":${POSTGRES_PORT} "; then
    echo "   ✓ PostgreSQL port ${POSTGRES_PORT} is listening"
else
    echo "   ✗ PostgreSQL port ${POSTGRES_PORT} is not listening"
    ERRORS=$((ERRORS + 1))
fi

echo ""
echo "=== Summary ==="
if [ ${ERRORS} -eq 0 ]; then
    echo "✓ All checks passed! Environment is ready for e2e testing."
    echo ""
    echo "Connection details:"
    echo "  MySQL:      mysql://root:test123456@127.0.0.1:${MYSQL_PORT}/mystisql_test"
    echo "  PostgreSQL: postgresql://postgres:test123456@127.0.0.1:${POSTGRES_PORT}/mystisql_test"
    exit 0
else
    echo "✗ ${ERRORS} check(s) failed. Please fix the issues above."
    echo ""
    echo "To setup the environment, run: scripts/e2e/setup-test-env.sh"
    exit 1
fi
