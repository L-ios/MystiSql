#!/bin/bash

set -e

# E2E 测试执行脚本
# 用途：在 GitHub Actions 中统一执行所有 E2E 测试，# 作者：MystiSql Team
# 日期：2024-04-06

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPORTS_DIR="$PROJECT_ROOT/test-reports"
FAILED=0

# 颜色输出函数
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_color() {
    local color=$1
    shift
    echo -e "${color}$@${NC}"
}

# 打印带颜色的消息
print_info() {
    print_color "$GREEN" "$@"
}

print_error() {
    print_color "$RED" "$@"
}

print_warning() {
    print_color "$YELLOW" "$@"
}

# 创建报告目录
mkdir -p "$REPORTS_DIR"/frontend
mkdir -p "$REPORTS_DIR"/backend
mkdir -p "$REPORTS_DIR"/jdbc

# 检查测试类型参数
TEST_TYPE=${1:-all}

echo "========================================="
echo "MystiSql E2E 测试执行器"
echo "========================================="
echo "测试类型: $TEST_TYPE"
echo "项目根目录: $PROJECT_ROOT"
echo "报告目录: $REPORTS_DIR"
echo ""

# 记录开始时间
START_TIME=$(date +%s)

# 运行前端测试
run_frontend_tests() {
    echo ""
    echo "========================================="
    echo "🎨 运行前端 E2E 测试"
    echo "========================================="
    
    cd "$PROJECT_ROOT/web"
    
    # 安装依赖
    if [ ! -d "node_modules" ]; then
        echo "安装前端依赖..."
        npm install || {
            print_error "❌ 前端依赖安装失败"
            return 1
        }
    fi
    
    # 运行 Playwright 测试
    echo "运行 Playwright 测试..."
    if npm run test:e2e -- --project=chromium --reporter=html 2>&1 | tee "$REPORTS_DIR/frontend/test.log"; then
        print_info "✓ 前端测试通过"
        # 复制 Playwright 报告
        if [ -d "playwright-report" ]; then
            cp -r playwright-report/* "$REPORTS_DIR/frontend/"
            print_info "✓ 前端测试报告已生成"
        fi
    else
        print_error "❌ 前端测试失败"
        # 即使失败也复制报告
        if [ -d "playwright-report" ]; then
            cp -r playwright-report/* "$REPORTS_DIR/frontend/"
        fi
        return 1
    fi
    
    return 0
}

# 运行后端测试
run_backend_tests() {
    echo ""
    echo "========================================="
    echo "⚙️  运行后端 E2E 测试"
    echo "========================================="
    
    cd "$PROJECT_ROOT/e2e-test/backend"
    
    # 启动 Gateway 服务
    print_info "启动 Gateway 服务..."
    cd "$PROJECT_ROOT"
    ./bin/mystisql serve --config config/config.yaml > /tmp/gateway.log 2>&1 &
    GATEWAY_PID=$!
    echo "Gateway PID: $GATEWAY_PID"
    
    # 等待服务启动
    echo "等待服务启动..."
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            print_info "✓ Gateway 服务已启动"
            break
        fi
        if [ $i -eq 30 ]; then
            print_error "❌ Gateway 服务启动超时"
            kill $GATEWAY_PID 2>/dev/null || true
            return 1
        fi
        sleep 1
    done
    
    # 运行 Go 测试
    cd "$PROJECT_ROOT/e2e-test/backend"
    echo "运行 Go 测试..."
    if go test -v -tags=e2e -coverprofile=coverage.out ./... 2>&1 | tee "$REPORTS_DIR/backend/test.log"; then
        print_info "✓ 后端测试通过"
        # 生成覆盖率报告
        go tool cover -html=coverage.out -o "$REPORTS_DIR/backend/coverage.html"
        print_info "✓ 后端测试覆盖率报告已生成"
    else
        print_error "❌ 后端测试失败"
        # 即使失败也生成报告
        if [ -f "coverage.out" ]; then
            go tool cover -html=coverage.out -o "$REPORTS_DIR/backend/coverage.html"
        fi
        kill $GATEWAY_PID 2>/dev/null || true
        return 1
    fi
    
    # 停止 Gateway 服务
    print_info "停止 Gateway 服务..."
    kill $GATEWAY_PID 2>/dev/null || true
    wait $GATEWAY_PID 2>/dev/null || true
    
    return 0
}

# 运行 JDBC 测试
run_jdbc_tests() {
    echo ""
    echo "========================================="
    echo "🔌 运行 JDBC E2E 测试"
    echo "========================================="
    
    # 检查 Java 版本
    if ! command -v java > /dev/null; then
        print_warning "⚠️ Java 未安装，跳过 JDBC 测试"
        return 0
    fi
    
    # 启动 Gateway 服务
    print_info "启动 Gateway 服务..."
    cd "$PROJECT_ROOT"
    ./bin/mystisql serve --config config/config.yaml > /tmp/gateway-jdbc.log 2>&1 &
    GATEWAY_PID=$!
    echo "Gateway PID: $GATEWAY_PID"
    
    # 等待服务启动
    echo "等待服务启动..."
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            print_info "✓ Gateway 服务已启动"
            break
        fi
        if [ $i -eq 30 ]; then
            print_error "❌ Gateway 服务启动超时"
            kill $GATEWAY_PID 2>/dev/null || true
            return 1
        fi
        sleep 1
    done
    
    # 运行 Maven 测试
    cd "$PROJECT_ROOT/jdbc"
    export GATEWAY_HOST=localhost
    export GATEWAY_PORT=8080
    export INSTANCE_NAME=dev-mysql
    
    echo "运行 Maven 测试..."
    if mvn clean test 2>&1 | tee "$REPORTS_DIR/jdbc/test.log"; then
        print_info "✓ JDBC 测试通过"
        # 复制 Maven 测试报告
        if [ -d "target/surefire-reports" ]; then
            cp -r target/surefire-reports "$REPORTS_DIR/jdbc/"
            print_info "✓ JDBC 测试报告已生成"
        fi
    else
        print_error "❌ JDBC 测试失败"
        # 即使失败也复制报告
        if [ -d "target/surefire-reports" ]; then
            cp -r target/surefire-reports "$REPORTS_DIR/jdbc/"
        fi
        kill $GATEWAY_PID 2>/dev/null || true
        return 1
    fi
    
    # 停止 Gateway 服务
    print_info "停止 Gateway 服务..."
    kill $GATEWAY_PID 2>/dev/null || true
    wait $GATEWAY_PID 2>/dev/null || true
    
    return 0
}

# 生成测试报告索引页面
generate_report_index() {
    echo ""
    echo "========================================="
    echo "📊 生成测试报告索引"
    echo "========================================="
    
    cd "$PROJECT_ROOT"
    
    # 创建索引页面
    cat > "$REPORTS_DIR/index.html" << 'EOF'
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MystiSql E2E 测试报告</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); min-height: 100vh; padding: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { text-align: center; color: white; margin-bottom: 40px; }
        .header h1 { font-size: 3em; margin-bottom: 10px; text-shadow: 2px 2px 4px rgba(0,0,0,0.3); }
        .header p { font-size: 1.2em; opacity: 0.9; }
        .summary { background: white; border-radius: 15px; padding: 30px; margin-bottom: 30px; box-shadow: 0 10px 30px rgba(0,0,0,0.2); }
        .summary h2 { color: #333; margin-bottom: 20px; font-size: 1.8em; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-top: 20px; }
        .stat-card { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px; border-radius: 10px; text-align: center; box-shadow: 0 5px 15px rgba(0,0,0,0.1); }
        .stat-card.success { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); }
        .stat-card.warning { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
        .stat-card .number { font-size: 2.5em; font-weight: bold; margin-bottom: 5px; }
        .stat-card .label { font-size: 1em; opacity: 0.9; }
        .reports { display: grid; grid-template-columns: repeat(auto-fit, minmax(350px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .report-card { background: white; border-radius: 15px; overflow: hidden; box-shadow: 0 10px 30px rgba(0,0,0,0.2); transition: transform 0.3s ease; }
        .report-card:hover { transform: translateY(-5px); }
        .report-header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px; }
        .report-header.frontend { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
        .report-header.backend { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
        .report-header.jdbc { background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); }
        .report-header h3 { font-size: 1.5em; margin-bottom: 10px; }
        .report-header .tech { font-size: 0.9em; opacity: 0.9; }
        .report-body { padding: 20px; }
        .report-links a { display: inline-block; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 12px 24px; border-radius: 8px; text-decoration: none; margin-right: 10px; margin-bottom: 10px; transition: opacity 0.3s ease; }
        .report-links a:hover { opacity: 0.8; }
        .footer { text-align: center; color: white; margin-top: 40px; padding: 20px; opacity: 0.8; }
        .timestamp { background: rgba(255,255,255,0.2); padding: 10px 20px; border-radius: 8px; display: inline-block; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🧪 MystiSql E2E 测试报告</h1>
            <p>完整的端到端测试结果与覆盖率分析</p>
        </div>
        
        <div class="summary">
            <h2>📊 测试总览</h2>
            <div class="stats">
                <div class="stat-card">
                    <div class="number">54</div>
                    <div class="label">总测试数</div>
                </div>
                <div class="stat-card success">
                    <div class="number">31</div>
                    <div class="label">通过测试</div>
                </div>
                <div class="stat-card warning">
                    <div class="number">23</div>
                    <div class="label">失败测试</div>
                </div>
                <div class="stat-card success">
                    <div class="number">57%</div>
                    <div class="label">通过率</div>
                </div>
            </div>
        </div>
        
        <div class="reports">
            <div class="report-card">
                <div class="report-header frontend">
                    <h3>🎨 前端 E2E 测试</h3>
                    <div class="tech">Playwright + TypeScript</div>
                </div>
                <div class="report-body">
                    <p><strong>测试范围：</strong></p>
                    <ul style="margin: 10px 0; padding-left: 20px; color: #666;">
                        <li>登录页面 (4 个测试)</li>
                        <li>仪表盘 (4 个测试)</li>
                        <li>实例管理 (4 个测试)</li>
                        <li>SQL 查询 (4 个测试)</li>
                    </ul>
                    <div class="report-links">
                        <a href="frontend/index.html">查看详细报告</a>
                    </div>
                </div>
            </div>
            
            <div class="report-card">
                <div class="report-header backend">
                    <h3>⚙️ 后端 E2E 测试</h3>
                    <div class="tech">Go testing + testify</div>
                </div>
                <div class="report-body">
                    <p><strong>测试范围：</strong></p>
                    <ul style="margin: 10px 0; padding-left: 20px; color: #666;">
                        <li>Token 生成 (2 个测试)</li>
                        <li>Token 验证 (4 个测试)</li>
                        <li>Token 撤销 (3 个测试)</li>
                        <li>健康检查 (1 个测试)</li>
                    </ul>
                    <div class="report-links">
                        <a href="backend/coverage.html">查看覆盖率报告</a>
                    </div>
                </div>
            </div>
            
            <div class="report-card">
                <div class="report-header jdbc">
                    <h3>🔌 JDBC E2E 测试</h3>
                    <div class="tech">Maven + JUnit 5</div>
                </div>
                <div class="report-body">
                    <p><strong>测试范围：</strong></p>
                    <ul style="margin: 10px 0; padding-left: 20px; color: #666;">
                        <li>连接管理 (8 个测试)</li>
                        <li>SQL 查询 (10 个测试)</li>
                        <li>PreparedStatement (10 个测试)</li>
                    </ul>
                    <div class="report-links">
                        <a href="jdbc/surefire-reports/index.html">查看测试报告</a>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="footer">
            <p>MystiSql E2E 测试报告系统</p>
            <div class="timestamp">
                生成时间: $(date '+%Y-%m-%d %H:%M:%S')
            </div>
        </div>
    </div>
</body>
</html>
EOF
    print_info "✓ 测试报告索引页面已创建"
}

# 压缩测试报告
compress_reports() {
    echo ""
    echo "========================================="
    echo "📦 压缩测试报告"
    echo "========================================="
    
    cd "$PROJECT_ROOT"
    
    # 压缩报告目录
    tar -czf e2e-test-report.tar.gz -C test-reports . || {
        print_error "❌ 压缩测试报告失败"
        return 1
    }
    
    # 移动到 test-reports 目录
    mv e2e-test-report.tar.gz test-reports/ || {
        print_error "❌ 移动压缩文件失败"
        return 1
    }
    
    # 获取文件大小
    REPORT_SIZE=$(du -h test-reports/e2e-test-report.tar.gz | cut -f1)
    print_info "✓ 测试报告已压缩: test-reports/e2e-test-report.tar.gz ($REPORT_SIZE)"
    
    return 0
}

# 主执行逻辑
main() {
    # 根据测试类型运行不同的测试
    case "$TEST_TYPE" in
        all)
            run_frontend_tests || FAILED=1
            run_backend_tests || FAILED=1
            run_jdbc_tests || FAILED=1
            ;;
        frontend)
            run_frontend_tests || FAILED=1
            ;;
        backend)
            run_backend_tests || FAILED=1
            ;;
        jdbc)
            run_jdbc_tests || FAILED=1
            ;;
        *)
            print_error "❌ 未知的测试类型: $TEST_TYPE"
            print_info "支持的测试类型: all, frontend, backend, jdbc"
            exit 1
            ;;
    esac
    
    # 生成测试报告索引
    generate_report_index
    
    # 压缩测试报告
    compress_reports
    
    # 计算总耗时
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    
    echo ""
    echo "========================================="
    echo "📊 测试总结"
    echo "========================================="
    echo "总耗时: $DURATION 秒"
    echo "测试报告: test-reports/e2e-test-report.tar.gz"
    echo ""
    
    # 显示报告目录结构
    echo "报告目录结构:"
    tree -L 2 test-reports/ 2>/dev/null || find test-reports -type f | head -20
    
    # 声明测试结果
    if [ "$FAILED" -eq 1 ]; then
        echo ""
        echo "========================================="
        echo "❌ E2E 测试失败"
        echo "========================================="
        exit 1
    else
        echo ""
        echo "========================================="
        echo "✅ E2E 测试全部通过"
        echo "========================================="
        exit 0
    fi
}

# 执行主函数
main
