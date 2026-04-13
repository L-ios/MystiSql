package io.github.mystisql.jdbc.e2e;

import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Tag;

import java.sql.*;

import static org.junit.jupiter.api.Assertions.*;

/**
 * JDBC E2E 查询测试 - 测试基本的 SQL 查询功能
 */
@Tag("e2e")
class JdbcE2EQueryTest {

    private static String jdbcUrl;

    @BeforeAll
    static void setUp() {
        String gatewayHost = System.getenv().getOrDefault("GATEWAY_HOST", "localhost");
        int gatewayPort = Integer.parseInt(System.getenv().getOrDefault("GATEWAY_PORT", "8080"));
        String instanceName = System.getenv().getOrDefault("INSTANCE_NAME", "local-mysql");
        String authToken = System.getenv().getOrDefault("AUTH_TOKEN", "");

        if (authToken.isEmpty()) {
            jdbcUrl = String.format("jdbc:mystisql://%s:%d/%s", gatewayHost, gatewayPort, instanceName);
        } else {
            jdbcUrl = String.format("jdbc:mystisql://%s:%d/%s?token=%s", gatewayHost, gatewayPort, instanceName, authToken);
        }
    }

    @Test
    @DisplayName("E2E-QUERY-001: 执行简单 SELECT 查询")
    void testSimpleSelectQuery() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery("SELECT 1 AS id, 'test' AS name")) {

            assertTrue(rs.next(), "结果集应至少有一行");
            assertEquals(1, rs.getInt("id"));
            assertEquals("test", rs.getString("name"));
            assertFalse(rs.next(), "结果集应只有一行");

            System.out.println("✅ 简单 SELECT 查询成功");
        }
    }

    @Test
    @DisplayName("E2E-QUERY-002: 执行 SELECT 1 查询")
    void testSelectOne() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery("SELECT 1")) {

            assertTrue(rs.next());
            assertEquals(1, rs.getInt(1));
            assertFalse(rs.next());

            System.out.println("✅ SELECT 1 查询成功");
        }
    }

    @Test
    @DisplayName("E2E-QUERY-003: 执行多列 SELECT 查询")
    void testMultiColumnSelect() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery("SELECT 1 AS id, 'Alice' AS name, 25.5 AS score, true AS active")) {

            assertTrue(rs.next());
            assertEquals(1, rs.getInt("id"));
            assertEquals("Alice", rs.getString("name"));
            assertEquals(25.5, rs.getDouble("score"), 0.01);
            assertTrue(rs.getBoolean("active"));

            System.out.println("✅ 多列 SELECT 查询成功");
        }
    }

    @Test
    @DisplayName("E2E-QUERY-004: 测试 ResultSet 元数据")
    void testResultSetMetadata() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery("SELECT 1 AS id, 'test' AS name")) {

            ResultSetMetaData metaData = rs.getMetaData();
            assertNotNull(metaData);
            assertEquals(2, metaData.getColumnCount());
            assertEquals("id", metaData.getColumnName(1));
            assertEquals("name", metaData.getColumnName(2));

            System.out.println("✅ ResultSet 元数据获取成功");
        }
    }

    @Test
    @DisplayName("E2E-QUERY-005: 测试空结果集")
    void testEmptyResultSet() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery("SELECT 1 WHERE 1 = 0")) {

            assertFalse(rs.next(), "空结果集应没有数据");

            System.out.println("✅ 空结果集处理正常");
        }
    }

    @Test
    @DisplayName("E2E-QUERY-006: 测试 NULL 值处理")
    void testNullHandling() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery("SELECT NULL AS null_value")) {

            assertTrue(rs.next());
            Object value = rs.getObject("null_value");
            assertTrue(rs.wasNull(), "应为 NULL 值");
            assertNull(value);

            System.out.println("✅ NULL 值处理正常");
        }
    }

    @Test
    @DisplayName("E2E-QUERY-007: 测试不同数据类型")
    void testDifferentDataTypes() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery(
                 "SELECT " +
                 "  123 AS int_val, " +
                 "  1234567890 AS long_val, " +
                 "  3.14 AS double_val, " +
                 "  'hello' AS string_val, " +
                 "  true AS bool_val")) {

            assertTrue(rs.next());
            assertEquals(123, rs.getInt("int_val"));
            assertEquals(1234567890L, rs.getLong("long_val"));
            assertEquals(3.14, rs.getDouble("double_val"), 0.01);
            assertEquals("hello", rs.getString("string_val"));
            assertTrue(rs.getBoolean("bool_val"));

            System.out.println("✅ 不同数据类型处理正常");
        }
    }

    @Test
    @DisplayName("E2E-QUERY-008: 测试 execute 方法")
    void testExecuteMethod() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement()) {

            boolean isResultSet = stmt.execute("SELECT 1");
            assertTrue(isResultSet, "SELECT 应返回 true");

            ResultSet rs = stmt.getResultSet();
            assertNotNull(rs);
            assertTrue(rs.next());

            System.out.println("✅ execute 方法正常");
        }
    }

    @Test
    @DisplayName("E2E-QUERY-009: 测试 executeUpdate 方法")
    void testExecuteUpdateMethod() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement()) {

            // 创建临时表
            stmt.execute("CREATE TEMPORARY TABLE test_e2e (id INT, name VARCHAR(100))");
            
            // 插入数据
            int rowsInserted = stmt.executeUpdate("INSERT INTO test_e2e VALUES (1, 'test')");
            assertEquals(1, rowsInserted, "应插入 1 行");

            // 更新数据
            int rowsUpdated = stmt.executeUpdate("UPDATE test_e2e SET name = 'updated' WHERE id = 1");
            assertTrue(rowsUpdated >= 0, "更新行数应 >= 0");

            // 删除数据
            int rowsDeleted = stmt.executeUpdate("DELETE FROM test_e2e WHERE id = 1");
            assertTrue(rowsDeleted >= 0, "删除行数应 >= 0");

            System.out.println("✅ executeUpdate 方法正常");
        }
    }

    @Test
    @DisplayName("E2E-QUERY-010: 测试查询超时")
    void testQueryTimeout() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement()) {

            stmt.setQueryTimeout(5);
            assertEquals(5, stmt.getQueryTimeout());

            ResultSet rs = stmt.executeQuery("SELECT 1");
            assertTrue(rs.next());

            System.out.println("✅ 查询超时设置正常");
        }
    }
}
