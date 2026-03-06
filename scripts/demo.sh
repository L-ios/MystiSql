#!/bin/bash

# MystiSql Phase 1 功能演示脚本

set -e

echo "===================================="
echo "MystiSql Phase 1 功能演示"
echo "===================================="
echo ""

# 检查可执行文件
if [ ! -f "bin/mystisql" ]; then
    echo "❌ 可执行文件不存在，请先运行: go build -o bin/mystisql ./cmd/mystisql"
    exit 1
fi

echo "✅ 可执行文件存在"
echo ""

# 1. 显示帮助信息
echo "1. 显示帮助信息"
echo "------------------------------------"
./bin/mystisql --help
echo ""

# 2. 显示版本信息
echo "2. 显示版本信息"
echo "------------------------------------"
./bin/mystisql version
echo ""

# 3. 列出实例（使用测试配置）
echo "3. 列出数据库实例（表格格式）"
echo "------------------------------------"
./bin/mystisql -c test/config.yaml instances list
echo ""

echo "4. 列出数据库实例（JSON 格式）"
echo "------------------------------------"
./bin/mystisql -c test/config.yaml instances list --format json
echo ""

# 5. 获取单个实例详情
echo "5. 获取单个实例详情"
echo "------------------------------------"
./bin/mystisql -c test/config.yaml instances get test-mysql
echo ""

# 6. 测试查询命令（需要真实数据库）
echo "6. 测试查询命令"
echo "------------------------------------"
echo "注意: 此命令需要真实 MySQL 数据库，如果数据库不可用会跳过"
if timeout 2 bash -c "echo > /dev/tcp/localhost/3306" 2>/dev/null; then
    echo "✅ MySQL 数据库可达，执行查询..."
    ./bin/mystisql -c test/config.yaml query test-mysql "SELECT 1 as value" 2>&1 || echo "查询失败（可能是认证问题）"
else
    echo "⚠️  MySQL 数据库不可达，跳过查询测试"
fi
echo ""

# 7. 测试 API 服务器
echo "7. 测试 API 服务器"
echo "------------------------------------"
echo "启动 API 服务器（后台运行）..."
./bin/mystisql -c test/config.yaml serve > /tmp/mystisql_test.log 2>&1 &
SERVER_PID=$!
echo "服务器 PID: $SERVER_PID"

# 等待服务器启动
echo "等待服务器启动..."
sleep 2

# 测试健康检查端点
echo ""
echo "测试健康检查端点:"
curl -s http://localhost:8080/health | jq .

# 测试实例列表端点
echo ""
echo "测试实例列表端点:"
curl -s http://localhost:8080/api/v1/instances | jq .

# 测试查询端点（需要真实数据库）
echo ""
echo "测试查询端点:"
if timeout 2 bash -c "echo > /dev/tcp/localhost/3306" 2>/dev/null; then
    echo "✅ MySQL 数据库可达，执行查询..."
    curl -s -X POST http://localhost:8080/api/v1/query \
         -H "Content-Type: application/json" \
         -d '{"instance":"test-mysql","sql":"SELECT 1 as value"}' | jq . || echo "查询失败"
else
    echo "⚠️  MySQL 数据库不可达，跳过 API 查询测试"
fi

# 停止服务器
echo ""
echo "停止 API 服务器..."
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

echo ""
echo "===================================="
echo "✅ 演示完成"
echo "===================================="
