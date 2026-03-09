package io.github.mystisql.jdbc;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import static org.junit.jupiter.api.Assertions.*;

import java.sql.*;

/**
 * Unit tests for MystiSqlStatement
 */
class MystiSqlStatementTest {

    private MystiSqlConnection connection;
    private MystiSqlStatement statement;

    @BeforeEach
    void setUp() {
        connection = new MystiSqlConnection(
            "localhost",
            8080,
            "test-instance",
            "testuser",
            "test-token",
            30,
            false,
            true,
            10
        );
        statement = new MystiSqlStatement(connection);
    }

    @Test
    @DisplayName("Statement should be open initially")
    void testInitialState() throws SQLException {
        assertFalse(statement.isClosed());
    }

    @Test
    @DisplayName("Close should mark statement as closed")
    void testClose() throws SQLException {
        statement.close();
        assertTrue(statement.isClosed());
    }

    @Test
    @DisplayName("GetConnection should return the parent connection")
    void testGetConnection() throws SQLException {
        Connection conn = statement.getConnection();
        assertNotNull(conn);
        assertSame(connection, conn);
    }

    @Test
    @DisplayName("SetQueryTimeout should store timeout value")
    void testSetQueryTimeout() throws SQLException {
        statement.setQueryTimeout(30);
        assertEquals(30, statement.getQueryTimeout());
    }

    @Test
    @DisplayName("Default query timeout should be 0")
    void testDefaultQueryTimeout() throws SQLException {
        assertEquals(0, statement.getQueryTimeout());
    }

    @Test
    @DisplayName("GetUpdateCount should return -1 initially")
    void testGetUpdateCount() throws SQLException {
        assertEquals(-1, statement.getUpdateCount());
    }

    @Test
    @DisplayName("GetResultSet should return null initially")
    void testGetResultSet() throws SQLException {
        assertNull(statement.getResultSet());
    }

    @Test
    @DisplayName("GetMaxFieldSize should return 0")
    void testGetMaxFieldSize() throws SQLException {
        assertEquals(0, statement.getMaxFieldSize());
    }

    @Test
    @DisplayName("SetMaxFieldSize should not throw")
    void testSetMaxFieldSize() throws SQLException {
        statement.setMaxFieldSize(1000);
    }

    @Test
    @DisplayName("GetMaxRows should return 0")
    void testGetMaxRows() throws SQLException {
        assertEquals(0, statement.getMaxRows());
    }

    @Test
    @DisplayName("SetMaxRows should not throw")
    void testSetMaxRows() throws SQLException {
        statement.setMaxRows(100);
    }

    @Test
    @DisplayName("GetWarnings should return null")
    void testGetWarnings() throws SQLException {
        assertNull(statement.getWarnings());
    }

    @Test
    @DisplayName("ClearWarnings should not throw")
    void testClearWarnings() throws SQLException {
        statement.clearWarnings();
    }

    @Test
    @DisplayName("IsPoolable should return false")
    void testIsPoolable() throws SQLException {
        assertFalse(statement.isPoolable());
    }

    @Test
    @DisplayName("SetPoolable should not throw")
    void testSetPoolable() throws SQLException {
        statement.setPoolable(true);
    }

    @Test
    @DisplayName("ExecuteQuery should throw SQLException (not implemented)")
    void testExecuteQuery() {
        assertThrows(SQLException.class, () -> statement.executeQuery("SELECT * FROM users"));
    }

    @Test
    @DisplayName("ExecuteUpdate should throw SQLException (not implemented)")
    void testExecuteUpdate() {
        assertThrows(SQLException.class, () -> statement.executeUpdate("DELETE FROM users"));
    }

    @Test
    @DisplayName("Execute should throw SQLException (not implemented)")
    void testExecute() {
        assertThrows(SQLException.class, () -> statement.execute("SELECT * FROM users"));
    }

    @Test
    @DisplayName("Operations on closed statement should throw SQLException")
    void testOperationsAfterClose() throws SQLException {
        statement.close();
        
        assertThrows(SQLException.class, () -> statement.getQueryTimeout());
        assertThrows(SQLException.class, () -> statement.setQueryTimeout(10));
        assertThrows(SQLException.class, () -> statement.getConnection());
        assertThrows(SQLException.class, () -> statement.getResultSet());
        assertThrows(SQLException.class, () -> statement.getUpdateCount());
        assertThrows(SQLException.class, () -> statement.executeQuery("SELECT 1"));
        assertThrows(SQLException.class, () -> statement.executeUpdate("DELETE FROM t"));
    }

    @Test
    @DisplayName("Cancel should not throw")
    void testCancel() throws SQLException {
        statement.cancel();
    }

    @Test
    @DisplayName("SetEscapeProcessing should not throw")
    void testSetEscapeProcessing() throws SQLException {
        statement.setEscapeProcessing(true);
    }

    @Test
    @DisplayName("SetCursorName should not throw")
    void testSetCursorName() throws SQLException {
        statement.setCursorName("cursor1");
    }

    @Test
    @DisplayName("AddBatch should throw SQLFeatureNotSupportedException")
    void testAddBatch() {
        assertThrows(SQLFeatureNotSupportedException.class, 
            () -> statement.addBatch("SELECT * FROM users"));
    }

    @Test
    @DisplayName("ExecuteBatch should throw SQLFeatureNotSupportedException")
    void testExecuteBatch() {
        assertThrows(SQLFeatureNotSupportedException.class, 
            () -> statement.executeBatch());
    }

    @Test
    @DisplayName("ClearBatch should not throw")
    void testClearBatch() throws SQLException {
        statement.clearBatch();
    }

    @Test
    @DisplayName("GetGeneratedKeys should return empty ResultSet")
    void testGetGeneratedKeys() throws SQLException {
        ResultSet rs = statement.getGeneratedKeys();
        assertNotNull(rs);
        assertFalse(rs.next()); // Empty result set
    }

    @Test
    @DisplayName("ExecuteUpdate with autoGeneratedKeys should delegate to executeUpdate")
    void testExecuteUpdateWithAutoGeneratedKeys() {
        assertThrows(SQLException.class, 
            () -> statement.executeUpdate("DELETE FROM t", Statement.NO_GENERATED_KEYS));
    }

    @Test
    @DisplayName("Execute with autoGeneratedKeys should delegate to execute")
    void testExecuteWithAutoGeneratedKeys() {
        assertThrows(SQLException.class, 
            () -> statement.execute("SELECT * FROM t", Statement.NO_GENERATED_KEYS));
    }

    @Test
    @DisplayName("CloseOnCompletion should not throw")
    void testCloseOnCompletion() throws SQLException {
        statement.closeOnCompletion();
    }

    @Test
    @DisplayName("IsCloseOnCompletion should return false")
    void testIsCloseOnCompletion() throws SQLException {
        assertFalse(statement.isCloseOnCompletion());
    }

    @Test
    @DisplayName("GetLargeUpdateCount should return updateCount as long")
    void testGetLargeUpdateCount() throws SQLException {
        assertEquals(-1L, statement.getLargeUpdateCount());
    }

    @Test
    @DisplayName("GetLargeMaxRows should return 0")
    void testGetLargeMaxRows() throws SQLException {
        assertEquals(0L, statement.getLargeMaxRows());
    }

    @Test
    @DisplayName("SetLargeMaxRows should not throw")
    void testSetLargeMaxRows() throws SQLException {
        statement.setLargeMaxRows(1000L);
    }

    @Test
    @DisplayName("IsWrapperFor should return false for unsupported interface")
    void testIsWrapperFor() throws SQLException {
        assertFalse(statement.isWrapperFor(ResultSet.class));
    }
    
    @Test
    @DisplayName("Unwrap should throw SQLFeatureNotSupportedException")
    void testUnwrap() {
        assertThrows(SQLFeatureNotSupportedException.class, () -> statement.unwrap(ResultSet.class));
    }

    @Test
    @DisplayName("Close should be idempotent")
    void testDoubleClose() throws SQLException {
        statement.close();
        statement.close();
        assertTrue(statement.isClosed());
    }
}
