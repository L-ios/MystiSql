package io.github.mystisql.jdbc;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import static org.junit.jupiter.api.Assertions.*;

import java.sql.*;
import java.util.Properties;

/**
 * Unit tests for MystiSqlConnection
 */
class MystiSqlConnectionTest {

    private MystiSqlConnection connection;

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
    }

    @Test
    @DisplayName("Connection should be open initially")
    void testInitialState() throws SQLException {
        assertFalse(connection.isClosed());
    }

    @Test
    @DisplayName("Close should mark connection as closed")
    void testClose() throws SQLException {
        connection.close();
        assertTrue(connection.isClosed());
    }

    @Test
    @DisplayName("Double close should be idempotent")
    void testDoubleClose() throws SQLException {
        connection.close();
        connection.close();
        assertTrue(connection.isClosed());
    }

    @Test
    @DisplayName("CreateStatement should return a MystiSqlStatement")
    void testCreateStatement() throws SQLException {
        Statement stmt = connection.createStatement();
        assertNotNull(stmt);
        assertTrue(stmt instanceof MystiSqlStatement);
    }

    @Test
    @DisplayName("PrepareStatement should return a MystiSqlPreparedStatement")
    void testPrepareStatement() throws SQLException {
        PreparedStatement pstmt = connection.prepareStatement("SELECT * FROM users WHERE id = ?");
        assertNotNull(pstmt);
        assertTrue(pstmt instanceof MystiSqlPreparedStatement);
    }

    @Test
    @DisplayName("GetMetaData should return a MystiSqlDatabaseMetaData")
    void testGetMetaData() throws SQLException {
        DatabaseMetaData metaData = connection.getMetaData();
        assertNotNull(metaData);
        assertTrue(metaData instanceof MystiSqlDatabaseMetaData);
    }

    @Test
    @DisplayName("AutoCommit should be true by default")
    void testAutoCommitDefault() throws SQLException {
        assertTrue(connection.getAutoCommit());
    }

    @Test
    @DisplayName("SetAutoCommit should change autoCommit state")
    void testSetAutoCommit() throws SQLException {
        connection.setAutoCommit(false);
        assertFalse(connection.getAutoCommit());
        
        connection.setAutoCommit(true);
        assertTrue(connection.getAutoCommit());
    }

    @Test
    @DisplayName("Commit should throw when autoCommit is enabled")
    void testCommitWithAutoCommit() {
        assertThrows(SQLException.class, () -> connection.commit());
    }

    @Test
    @DisplayName("Commit should throw SQLFeatureNotSupportedException when autoCommit is disabled")
    void testCommitWithoutAutoCommit() throws SQLException {
        connection.setAutoCommit(false);
        assertThrows(SQLFeatureNotSupportedException.class, () -> connection.commit());
    }

    @Test
    @DisplayName("Rollback should throw when autoCommit is enabled")
    void testRollbackWithAutoCommit() {
        assertThrows(SQLException.class, () -> connection.rollback());
    }

    @Test
    @DisplayName("Rollback should throw SQLFeatureNotSupportedException when autoCommit is disabled")
    void testRollbackWithoutAutoCommit() throws SQLException {
        connection.setAutoCommit(false);
        assertThrows(SQLFeatureNotSupportedException.class, () -> connection.rollback());
    }

    @Test
    @DisplayName("IsValid should return false when connection is closed")
    void testIsValidClosed() throws SQLException {
        connection.close();
        assertFalse(connection.isValid(0));
    }

    @Test
    @DisplayName("IsValid should throw for negative timeout")
    void testIsValidNegativeTimeout() {
        assertThrows(SQLException.class, () -> connection.isValid(-1));
    }

    @Test
    @DisplayName("Operations on closed connection should throw SQLException")
    void testOperationsAfterClose() throws SQLException {
        connection.close();
        
        assertThrows(SQLException.class, () -> connection.createStatement());
        assertThrows(SQLException.class, () -> connection.prepareStatement("SELECT 1"));
        assertThrows(SQLException.class, () -> connection.getMetaData());
        assertThrows(SQLException.class, () -> connection.getAutoCommit());
        assertThrows(SQLException.class, () -> connection.setAutoCommit(false));
    }

    @Test
    @DisplayName("Getters should return correct values")
    void testGetters() {
        assertEquals("localhost", connection.getHost());
        assertEquals(8080, connection.getPort());
        assertEquals("test-instance", connection.getInstanceName());
        assertEquals("testuser", connection.getUsername());
        assertEquals("test-token", connection.getToken());
        assertEquals(30, connection.getTimeout());
        assertFalse(connection.isSsl());
        assertNotNull(connection.getRestClient());
    }

    @Test
    @DisplayName("CreateStatement with parameters should work")
    void testCreateStatementWithParams() throws SQLException {
        Statement stmt = connection.createStatement(
            ResultSet.TYPE_FORWARD_ONLY,
            ResultSet.CONCUR_READ_ONLY
        );
        assertNotNull(stmt);
    }

    @Test
    @DisplayName("PrepareStatement with parameters should work")
    void testPrepareStatementWithParams() throws SQLException {
        PreparedStatement pstmt = connection.prepareStatement(
            "SELECT * FROM users",
            ResultSet.TYPE_FORWARD_ONLY,
            ResultSet.CONCUR_READ_ONLY
        );
        assertNotNull(pstmt);
    }

    @Test
    @DisplayName("PrepareCall should throw SQLFeatureNotSupportedException")
    void testPrepareCall() {
        assertThrows(SQLFeatureNotSupportedException.class, 
            () -> connection.prepareCall("{call test()}"));
    }

    @Test
    @DisplayName("NativeSQL should return the same SQL")
    void testNativeSQL() throws SQLException {
        String sql = "SELECT * FROM users";
        assertEquals(sql, connection.nativeSQL(sql));
    }

    @Test
    @DisplayName("GetHoldability should return HOLD_CURSORS_OVER_COMMIT")
    void testGetHoldability() throws SQLException {
        assertEquals(ResultSet.HOLD_CURSORS_OVER_COMMIT, connection.getHoldability());
    }

    @Test
    @DisplayName("GetSchema should return null")
    void testGetSchema() throws SQLException {
        assertNull(connection.getSchema());
    }

    @Test
    @DisplayName("SetSchema should not throw")
    void testSetSchema() throws SQLException {
        connection.setSchema("test");
    }

    @Test
    @DisplayName("GetClientInfo should return empty Properties")
    void testGetClientInfo() throws SQLException {
        Properties props = connection.getClientInfo();
        assertNotNull(props);
        assertTrue(props.isEmpty());
    }

    @Test
    @DisplayName("GetClientInfo by name should return null")
    void testGetClientInfoByName() throws SQLException {
        assertNull(connection.getClientInfo("test"));
    }

    @Test
    @DisplayName("SetClientInfo should not throw")
    void testSetClientInfo() throws SQLException {
        connection.setClientInfo("test", "value");
    }

    @Test
    @DisplayName("CreateBlob should throw SQLFeatureNotSupportedException")
    void testCreateBlob() {
        assertThrows(SQLFeatureNotSupportedException.class, () -> connection.createBlob());
    }

    @Test
    @DisplayName("CreateClob should throw SQLFeatureNotSupportedException")
    void testCreateClob() {
        assertThrows(SQLFeatureNotSupportedException.class, () -> connection.createClob());
    }

    @Test
    @DisplayName("CreateArray should throw SQLFeatureNotSupportedException")
    void testCreateArray() {
        assertThrows(SQLFeatureNotSupportedException.class, 
            () -> connection.createArrayOf("INT", new Object[]{}));
    }

    @Test
    @DisplayName("Connection with SSL should have correct protocol")
    void testSSLConnection() {
        MystiSqlConnection sslConn = new MystiSqlConnection(
            "secure.example.com",
            443,
            "prod-instance",
            "admin",
            "secure-token",
            60,
            true,
            true,
            20
        );
        
        assertTrue(sslConn.isSsl());
        assertEquals("secure.example.com", sslConn.getHost());
        assertEquals(443, sslConn.getPort());
    }
}
