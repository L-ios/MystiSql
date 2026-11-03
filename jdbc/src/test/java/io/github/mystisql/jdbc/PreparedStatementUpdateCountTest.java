package io.github.mystisql.jdbc;

import io.github.mystisql.jdbc.client.ExecResult;
import io.github.mystisql.jdbc.client.QueryRequest;
import io.github.mystisql.jdbc.client.Transport;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

import java.sql.ResultSet;
import java.sql.SQLException;

/**
 * 验证 PreparedStatement executeUpdate 后 updateCount / generatedKeys 正确传播（P1-004 回归测试）。
 */
class PreparedStatementUpdateCountTest {

    private MystiSqlConnection connection;
    private MystiSqlPreparedStatement pstmt;

    /** 返回固定 ExecResult 的 mock transport */
    private static class MockTransport implements Transport {
        ExecResult result = new ExecResult();

        MockTransport(long rowsAffected, Long lastInsertId) {
            result.setRowsAffected(rowsAffected);
            result.setLastInsertId(lastInsertId);
        }

        @Override
        public MystiSqlResultSet executeQuery(String instance, String sql) { return null; }

        @Override
        public MystiSqlResultSet executeQuery(QueryRequest request) { return null; }

        @Override
        public ExecResult executeUpdate(String instance, String sql) { return result; }

        @Override
        public ExecResult executeUpdate(QueryRequest request) { return result; }

        @Override
        public boolean healthCheck(String instance) { return true; }

        @Override
        public String getTransportType() { return "mock"; }

        @Override
        public void close() {}
    }

    @BeforeEach
    void setUp() {
        connection = new MystiSqlConnection(
            "localhost", 8080, "test-instance", "testuser", "test-token",
            30, false, true, 10
        );
    }

    @Test
    @DisplayName("executeUpdate should propagate rowsAffected to getUpdateCount")
    void testUpdateCountPropagated() throws SQLException {
        connection.setTransportForTesting(new MockTransport(42L, 100L));
        pstmt = new MystiSqlPreparedStatement(connection, "UPDATE users SET name = 'x' WHERE id > ?");

        int affected = pstmt.executeUpdate();

        assertEquals(42, affected);
        assertEquals(42, pstmt.getUpdateCount(),
            "getUpdateCount 应返回实际受影响行数而非 -1（P1-004）");
    }

    @Test
    @DisplayName("executeUpdate should propagate lastInsertId to getGeneratedKeys")
    void testGeneratedKeysPropagated() throws SQLException {
        connection.setTransportForTesting(new MockTransport(1L, 12345L));
        pstmt = new MystiSqlPreparedStatement(connection, "INSERT INTO users (name) VALUES (?)");

        pstmt.executeUpdate();
        ResultSet keys = pstmt.getGeneratedKeys();

        assertNotNull(keys);
        assertTrue(keys.next(), "getGeneratedKeys 应包含 lastInsertId（P1-004）");
        assertEquals(12345L, keys.getLong(1));
        assertFalse(keys.next());
    }

    @Test
    @DisplayName("execute(false) path should also propagate updateCount")
    void testExecuteUpdatePath() throws SQLException {
        connection.setTransportForTesting(new MockTransport(7L, null));
        pstmt = new MystiSqlPreparedStatement(connection, "DELETE FROM users WHERE id = ?");

        boolean isResultSet = pstmt.execute();

        assertFalse(isResultSet);
        assertEquals(7, pstmt.getUpdateCount());
    }

    @Test
    @DisplayName("getUpdateCount should be -1 before any execution")
    void testInitialUpdateCount() throws SQLException {
        connection.setTransportForTesting(new MockTransport(1L, 1L));
        pstmt = new MystiSqlPreparedStatement(connection, "UPDATE users SET name = ?");

        assertEquals(-1, pstmt.getUpdateCount());
    }
}
