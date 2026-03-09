#!/bin/bash

set -e

CONTAINER_NAME_MYSQL="mystisql-test-mysql"
CONTAINER_NAME_POSTGRES="mystisql-test-postgres"

echo "=== MystiSql E2E Test Environment Teardown ==="

echo ""
echo "Step 1: Stopping test containers..."
podman stop "${CONTAINER_NAME_MYSQL}" 2>/dev/null && echo "  MySQL container stopped" || echo "  MySQL container not running"
podman stop "${CONTAINER_NAME_POSTGRES}" 2>/dev/null && echo "  PostgreSQL container stopped" || echo "  PostgreSQL container not running"

echo ""
echo "Step 2: Removing test containers..."
podman rm "${CONTAINER_NAME_MYSQL}" 2>/dev/null && echo "  MySQL container removed" || echo "  MySQL container not found"
podman rm "${CONTAINER_NAME_POSTGRES}" 2>/dev/null && echo "  PostgreSQL container removed" || echo "  PostgreSQL container not found"

CLEAN_VOLUMES="${CLEAN_VOLUMES:-false}"

if [ "${CLEAN_VOLUMES}" = "true" ]; then
    echo ""
    echo "Step 3: Cleaning up dangling volumes..."
    podman volume prune -f && echo "  Volumes cleaned" || echo "  No volumes to clean"
else
    echo ""
    echo "Step 3: Skipping volume cleanup (set CLEAN_VOLUMES=true to clean)"
fi

echo ""
echo "=== Teardown Complete ==="
echo ""
echo "Remaining containers:"
podman ps -a --filter "name=mystisql-test-" --format "table {{.Names}}\t{{.Status}}" || echo "  No containers found"
