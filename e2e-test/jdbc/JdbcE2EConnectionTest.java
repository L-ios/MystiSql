package io.github.mystisql.jdbc.e2e;

import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Tag;

import java.sql.*;

import static org.junit.jupiter.api.Assertions.*;

/**
 * JDBC E2E Tests - 连接到真实的 MystiSql Gateway
 * 
 * 运行条件：
 * 1. MystiSql Gateway 服务已启动（默认端口 8080）
 * 2. MySQL 实例已配置并运行
 * 3. 设置环境变量：
 *    - GATEWAY_HOST (默认: localhost)
 *    - GATEWAY_PORT (默认: 8080)
 *    - INSTANCE_NAME (默认: local-mysql)
 *    - AUTH_TOKEN (可选，如果 Gateway 启用了认证)
 * 
 * 运行命令：
 * ./gradlew test --tests "io.github.mystisql.jdbc.e2e.*" -Pe2e=true
 */
@Tag("e2e")
class JdbcE2EConnectionTest {

    private static String gatewayHost;
    private static int gatewayPort;
    private static String instanceName;
    private static String authToken;
    private static String jdbcUrl;

    @BeforeAll
    static void setUp() {
        gatewayHost = System.getenv().getOrDefault("GATEWAY_HOST", "localhost");
        gatewayPort = Integer.parseInt(System.getenv().getOrDefault("GATEWAY_PORT", "8080"));
        instanceName = System.getenv().getOrDefault("INSTANCE_NAME", "local-mysql");
        authToken = System.getenv().getOrDefault("AUTH_TOKEN", "");

        if (authToken.isEmpty()) {
            jdbcUrl = String.format("jdbc:mystisql://%s:%d/%s", gatewayHost, gatewayPort, instanceName);
        } else {
            jdbcUrl = String.format("jdbc:mystisql://%s:%d/%s?token=%s", gatewayHost, gatewayPort, instanceName, authToken);
        }

        System.out.println("JDBC E2E Test Configuration:");
        System.out.println("  Gateway: " + gatewayHost + ":" + gatewayPort);
        System.out.println("  Instance: " + instanceName);
        System.out.println("  JDBC URL: " + jdbcUrl);
    }

    @Test
    @DisplayName("E2E-001: 连接到 Gateway 并验证连接")
    void testConnectToGateway() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl)) {
            assertNotNull(conn, "连接不应为 null");
            assertFalse(conn.isClosed(), "连接不应立即关闭");
            assertTrue(conn.isValid(5), "连接应该是有效的");

            System.out.println("✅ 成功连接到 MystiSql Gateway");
        }
    }

    @Test
    @DisplayName("E2E-002: 获取连接元数据")
    void testGetConnectionMetadata() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl)) {
            DatabaseMetaData metaData = conn.getMetaData();

            assertNotNull(metaData, "元数据不应为 null");
            assertEquals("MystiSql Gateway", metaData.getDatabaseProductName(), "数据库名称应为 MystiSql Gateway");
            assertNotNull(metaData.getDatabaseProductVersion(), "数据库版本不应为 null");
            assertEquals("MystiSql JDBC Driver", metaData.getDriverName(), "驱动名称应为 MystiSql JDBC Driver");
            assertNotNull(metaData.getDriverVersion(), "驱动版本不应为 null");

            System.out.println("✅ 数据库产品: " + metaData.getDatabaseProductName());
            System.out.println("✅ 数据库版本: " + metaData.getDatabaseProductVersion());
            System.out.println("✅ 驱动名称: " + metaData.getDriverName());
            System.out.println("✅ 驱动版本: " + metaData.getDriverVersion());
        }
    }

    @Test
    @DisplayName("E2E-003: 测试自动提交模式")
    void testAutoCommitMode() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl)) {
            assertTrue(conn.getAutoCommit(), "默认应为自动提交模式");

            conn.setAutoCommit(false);
            assertFalse(conn.getAutoCommit(), "应能关闭自动提交模式");

            conn.setAutoCommit(true);
            assertTrue(conn.getAutoCommit(), "应能重新启用自动提交模式");

            System.out.println("✅ 自动提交模式切换正常");
        }
    }

    @Test
    @DisplayName("E2E-004: 创建 Statement 对象")
    void testCreateStatement() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl)) {
            Statement stmt = conn.createStatement();

            assertNotNull(stmt, "Statement 不应为 null");
            assertFalse(stmt.isClosed(), "Statement 不应立即关闭");

            stmt.close();
            assertTrue(stmt.isClosed(), "Statement 关闭后应标记为已关闭");

            System.out.println("✅ Statement 创建和关闭正常");
        }
    }

    @Test
    @DisplayName("E2E-005: 创建 PreparedStatement 对象")
    void testCreatePreparedStatement() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl)) {
            String sql = "SELECT 1 AS test";
            PreparedStatement stmt = conn.prepareStatement(sql);

            assertNotNull(stmt, "PreparedStatement 不应为 null");
            assertFalse(stmt.isClosed(), "PreparedStatement 不应立即关闭");

            stmt.close();
            assertTrue(stmt.isClosed(), "PreparedStatement 关闭后应标记为已关闭");

            System.out.println("✅ PreparedStatement 创建和关闭正常");
        }
    }

    @Test
    @DisplayName("E2E-006: 测试连接关闭")
    void testConnectionClose() throws SQLException {
        Connection conn = DriverManager.getConnection(jdbcUrl);
        assertNotNull(conn);
        assertFalse(conn.isClosed());

        conn.close();
        assertTrue(conn.isClosed(), "连接关闭后应标记为已关闭");

        System.out.println("✅ 连接关闭正常");
    }

    @Test
    @DisplayName("E2E-007: 测试无效连接参数")
    void testInvalidConnectionParameters() {
        String invalidUrl = "jdbc:mystisql://invalid-host:9999/invalid-instance?timeout=1";

        assertThrows(SQLException.class, () -> {
            DriverManager.getConnection(invalidUrl);
        }, "连接到无效地址应抛出 SQLException");

        System.out.println("✅ 无效连接参数正确抛出异常");
    }

    @Test
    @DisplayName("E2E-008: 测试连接超时设置")
    void testConnectionTimeout() throws SQLException {
        String urlWithTimeout = jdbcUrl + (jdbcUrl.contains("?") ? "&" : "?") + "timeout=5";

        try (Connection conn = DriverManager.getConnection(urlWithTimeout)) {
            assertNotNull(conn);
            assertTrue(conn.isValid(3));

            System.out.println("✅ 连接超时设置正常");
        }
    }
}
