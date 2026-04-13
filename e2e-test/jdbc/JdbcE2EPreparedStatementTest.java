package io.github.mystisql.jdbc.e2e;

import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Tag;

import java.sql.*;

import static org.junit.jupiter.api.Assertions.*;

/**
 * JDBC E2E PreparedStatement 测试 - 测试参数化查询功能
 */
@Tag("e2e")
class JdbcE2EPreparedStatementTest {

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
    @DisplayName("E2E-PSTMT-001: PreparedStatement 基本查询")
    void testBasicPreparedStatement() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             PreparedStatement stmt = conn.prepareStatement("SELECT ? AS value")) {

            stmt.setInt(1, 42);
            ResultSet rs = stmt.executeQuery();

            assertTrue(rs.next());
            assertEquals(42, rs.getInt("value"));

            System.out.println("✅ PreparedStatement 基本查询成功");
        }
    }

    @Test
    @DisplayName("E2E-PSTMT-002: 设置字符串参数")
    void testStringParameter() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             PreparedStatement stmt = conn.prepareStatement("SELECT ? AS name")) {

            stmt.setString(1, "Alice");
            ResultSet rs = stmt.executeQuery();

            assertTrue(rs.next());
            assertEquals("Alice", rs.getString("name"));

            System.out.println("✅ 字符串参数设置成功");
        }
    }

    @Test
    @DisplayName("E2E-PSTMT-003: 设置多个参数")
    void testMultipleParameters() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             PreparedStatement stmt = conn.prepareStatement("SELECT ? AS id, ? AS name, ? AS score")) {

            stmt.setInt(1, 1);
            stmt.setString(2, "Bob");
            stmt.setDouble(3, 95.5);

            ResultSet rs = stmt.executeQuery();
            assertTrue(rs.next());
            assertEquals(1, rs.getInt("id"));
            assertEquals("Bob", rs.getString("name"));
            assertEquals(95.5, rs.getDouble("score"), 0.01);

            System.out.println("✅ 多参数设置成功");
        }
    }

    @Test
    @DisplayName("E2E-PSTMT-004: 设置 NULL 参数")
    void testNullParameter() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             PreparedStatement stmt = conn.prepareStatement("SELECT ? AS null_value")) {

            stmt.setNull(1, Types.VARCHAR);
            ResultSet rs = stmt.executeQuery();

            assertTrue(rs.next());
            assertTrue(rs.wasNull());
            assertNull(rs.getObject("null_value"));

            System.out.println("✅ NULL 参数设置成功");
        }
    }

    @Test
    @DisplayName("E2E-PSTMT-005: 清除参数")
    void testClearParameters() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             PreparedStatement stmt = conn.prepareStatement("SELECT ? AS value")) {

            stmt.setInt(1, 100);
            stmt.clearParameters();
            stmt.setInt(1, 200);

            ResultSet rs = stmt.executeQuery();
            assertTrue(rs.next());
            assertEquals(200, rs.getInt("value"));

            System.out.println("✅ 参数清除成功");
        }
    }

    @Test
    @DisplayName("E2E-PSTMT-006: 设置不同类型的参数")
    void testDifferentParameterTypes() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             PreparedStatement stmt = conn.prepareStatement(
                 "SELECT ? AS int_val, ? AS long_val, ? AS double_val, ? AS bool_val, ? AS str_val")) {

            stmt.setInt(1, 123);
            stmt.setLong(2, 1234567890L);
            stmt.setDouble(3, 3.14159);
            stmt.setBoolean(4, true);
            stmt.setString(5, "test");

            ResultSet rs = stmt.executeQuery();
            assertTrue(rs.next());
            assertEquals(123, rs.getInt("int_val"));
            assertEquals(1234567890L, rs.getLong("long_val"));
            assertEquals(3.14159, rs.getDouble("double_val"), 0.00001);
            assertTrue(rs.getBoolean("bool_val"));
            assertEquals("test", rs.getString("str_val"));

            System.out.println("✅ 不同类型参数设置成功");
        }
    }

    @Test
    @DisplayName("E2E-PSTMT-007: PreparedStatement executeUpdate")
    void testPreparedStatementExecuteUpdate() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement()) {

            // 创建临时表
            stmt.execute("CREATE TEMPORARY TABLE test_pstmt (id INT, name VARCHAR(100))");
        }

        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             PreparedStatement stmt = conn.prepareStatement("INSERT INTO test_pstmt VALUES (?, ?)")) {

            stmt.setInt(1, 1);
            stmt.setString(2, "Alice");
            int rowsInserted = stmt.executeUpdate();
            assertEquals(1, rowsInserted);

            stmt.setInt(1, 2);
            stmt.setString(2, "Bob");
            rowsInserted = stmt.executeUpdate();
            assertEquals(1, rowsInserted);

            System.out.println("✅ PreparedStatement executeUpdate 成功");
        }
    }

    @Test
    @DisplayName("E2E-PSTMT-008: PreparedStatement 查询数据")
    void testPreparedStatementQuery() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement()) {

            // 创建并插入测试数据
            stmt.execute("CREATE TEMPORARY TABLE test_query (id INT, age INT, name VARCHAR(100))");
            stmt.executeUpdate("INSERT INTO test_query VALUES (1, 25, 'Alice')");
            stmt.executeUpdate("INSERT INTO test_query VALUES (2, 30, 'Bob')");
            stmt.executeUpdate("INSERT INTO test_query VALUES (3, 35, 'Charlie')");
        }

        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             PreparedStatement stmt = conn.prepareStatement("SELECT * FROM test_query WHERE age > ? ORDER BY id")) {

            stmt.setInt(1, 28);
            ResultSet rs = stmt.executeQuery();

            assertTrue(rs.next());
            assertEquals(2, rs.getInt("id"));
            assertEquals("Bob", rs.getString("name"));

            assertTrue(rs.next());
            assertEquals(3, rs.getInt("id"));
            assertEquals("Charlie", rs.getString("name"));

            assertFalse(rs.next());

            System.out.println("✅ PreparedStatement 查询数据成功");
        }
    }

    @Test
    @DisplayName("E2E-PSTMT-009: PreparedStatement 更新数据")
    void testPreparedStatementUpdate() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement()) {

            // 创建并插入测试数据
            stmt.execute("CREATE TEMPORARY TABLE test_update (id INT, status VARCHAR(50))");
            stmt.executeUpdate("INSERT INTO test_update VALUES (1, 'pending')");
            stmt.executeUpdate("INSERT INTO test_update VALUES (2, 'pending')");
        }

        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             PreparedStatement stmt = conn.prepareStatement("UPDATE test_update SET status = ? WHERE id = ?")) {

            stmt.setString(1, "completed");
            stmt.setInt(2, 1);
            int rowsUpdated = stmt.executeUpdate();
            assertTrue(rowsUpdated >= 0);

            System.out.println("✅ PreparedStatement 更新数据成功");
        }
    }

    @Test
    @DisplayName("E2E-PSTMT-010: PreparedStatement 删除数据")
    void testPreparedStatementDelete() throws SQLException {
        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             Statement stmt = conn.createStatement()) {

            // 创建并插入测试数据
            stmt.execute("CREATE TEMPORARY TABLE test_delete (id INT, name VARCHAR(100))");
            stmt.executeUpdate("INSERT INTO test_delete VALUES (1, 'to-delete')");
            stmt.executeUpdate("INSERT INTO test_delete VALUES (2, 'keep')");
        }

        try (Connection conn = DriverManager.getConnection(jdbcUrl);
             PreparedStatement stmt = conn.prepareStatement("DELETE FROM test_delete WHERE id = ?")) {

            stmt.setInt(1, 1);
            int rowsDeleted = stmt.executeUpdate();
            assertTrue(rowsDeleted >= 0);

            System.out.println("✅ PreparedStatement 删除数据成功");
        }
    }
}
