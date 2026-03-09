package io.github.mystisql.jdbc;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import static org.junit.jupiter.api.Assertions.*;

import java.sql.*;
import java.math.BigDecimal;
import java.util.Calendar;

/**
 * Unit tests for MystiSqlPreparedStatement
 */
class MystiSqlPreparedStatementTest {

    private MystiSqlConnection connection;
    private MystiSqlPreparedStatement pstmt;

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
        pstmt = new MystiSqlPreparedStatement(
            connection, 
            "SELECT * FROM users WHERE id = ? AND name = ?"
        );
    }

    @Test
    @DisplayName("PreparedStatement should store SQL")
    void testGetSql() {
        assertEquals("SELECT * FROM users WHERE id = ? AND name = ?", pstmt.getSql());
    }

    @Test
    @DisplayName("Parameters should be empty initially")
    void testInitialParameters() {
        assertTrue(pstmt.getParameters().isEmpty());
    }

    @Test
    @DisplayName("SetInt should add parameter at correct index")
    void testSetInt() throws SQLException {
        pstmt.setInt(1, 42);
        
        assertEquals(1, pstmt.getParameters().size());
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("INTEGER", param.getType());
        assertEquals(42, param.getValue());
    }

    @Test
    @DisplayName("SetString should add parameter at correct index")
    void testSetString() throws SQLException {
        pstmt.setString(2, "Alice");
        
        assertEquals(2, pstmt.getParameters().size());
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(1);
        assertEquals("VARCHAR", param.getType());
        assertEquals("Alice", param.getValue());
    }

    @Test
    @DisplayName("SetLong should add parameter")
    void testSetLong() throws SQLException {
        pstmt.setLong(1, 12345L);
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("BIGINT", param.getType());
        assertEquals(12345L, param.getValue());
    }

    @Test
    @DisplayName("SetBoolean should add parameter")
    void testSetBoolean() throws SQLException {
        pstmt.setBoolean(1, true);
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("BOOLEAN", param.getType());
        assertEquals(true, param.getValue());
    }

    @Test
    @DisplayName("SetDouble should add parameter")
    void testSetDouble() throws SQLException {
        pstmt.setDouble(1, 3.14);
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("DOUBLE", param.getType());
        assertEquals(3.14, param.getValue());
    }

    @Test
    @DisplayName("SetFloat should add parameter")
    void testSetFloat() throws SQLException {
        pstmt.setFloat(1, 2.5f);
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("FLOAT", param.getType());
        assertEquals(2.5f, param.getValue());
    }

    @Test
    @DisplayName("SetShort should add parameter")
    void testSetShort() throws SQLException {
        pstmt.setShort(1, (short) 100);
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("SMALLINT", param.getType());
        assertEquals((short) 100, param.getValue());
    }

    @Test
    @DisplayName("SetByte should add parameter")
    void testSetByte() throws SQLException {
        pstmt.setByte(1, (byte) 10);
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("TINYINT", param.getType());
        assertEquals((byte) 10, param.getValue());
    }

    @Test
    @DisplayName("SetBigDecimal should add parameter")
    void testSetBigDecimal() throws SQLException {
        BigDecimal decimal = new BigDecimal("123.45");
        pstmt.setBigDecimal(1, decimal);
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("DECIMAL", param.getType());
        assertEquals(decimal, param.getValue());
    }

    @Test
    @DisplayName("SetNull should add parameter with null value")
    void testSetNull() throws SQLException {
        pstmt.setNull(1, Types.VARCHAR);
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("NULL", param.getType());
        assertNull(param.getValue());
    }

    @Test
    @DisplayName("SetDate should add parameter")
    void testSetDate() throws SQLException {
        java.sql.Date date = java.sql.Date.valueOf("2024-01-15");
        pstmt.setDate(1, date);
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("DATE", param.getType());
        assertEquals(date, param.getValue());
    }

    @Test
    @DisplayName("SetTime should add parameter")
    void testSetTime() throws SQLException {
        Time time = Time.valueOf("10:30:00");
        pstmt.setTime(1, time);
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("TIME", param.getType());
        assertEquals(time, param.getValue());
    }

    @Test
    @DisplayName("SetTimestamp should add parameter")
    void testSetTimestamp() throws SQLException {
        Timestamp timestamp = Timestamp.valueOf("2024-01-15 10:30:00");
        pstmt.setTimestamp(1, timestamp);
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("TIMESTAMP", param.getType());
        assertEquals(timestamp, param.getValue());
    }

    @Test
    @DisplayName("SetBytes should add parameter")
    void testSetBytes() throws SQLException {
        byte[] bytes = new byte[]{1, 2, 3, 4, 5};
        pstmt.setBytes(1, bytes);
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("BINARY", param.getType());
        assertArrayEquals(bytes, (byte[]) param.getValue());
    }

    @Test
    @DisplayName("Parameter index should be 1-based")
    void testParameterIndexOneBased() throws SQLException {
        pstmt.setInt(1, 10);
        pstmt.setString(2, "test");
        
        assertEquals(10, pstmt.getParameters().get(0).getValue());
        assertEquals("test", pstmt.getParameters().get(1).getValue());
    }

    @Test
    @DisplayName("Parameter index less than 1 should throw SQLException")
    void testInvalidParameterIndex() {
        assertThrows(SQLException.class, () -> pstmt.setInt(0, 10));
        assertThrows(SQLException.class, () -> pstmt.setInt(-1, 10));
    }

    @Test
    @DisplayName("ClearParameters should remove all parameters")
    void testClearParameters() throws SQLException {
        pstmt.setInt(1, 10);
        pstmt.setString(2, "test");
        
        assertEquals(2, pstmt.getParameters().size());
        
        pstmt.clearParameters();
        
        assertTrue(pstmt.getParameters().isEmpty());
    }

    @Test
    @DisplayName("Setting parameter at higher index should expand list")
    void testParameterExpansion() throws SQLException {
        pstmt.setInt(3, 100);
        
        assertEquals(3, pstmt.getParameters().size());
        assertNull(pstmt.getParameters().get(0));
        assertNull(pstmt.getParameters().get(1));
        assertEquals(100, pstmt.getParameters().get(2).getValue());
    }

    @Test
    @DisplayName("Overwriting parameter should work")
    void testOverwriteParameter() throws SQLException {
        pstmt.setInt(1, 10);
        pstmt.setInt(1, 20);
        
        assertEquals(1, pstmt.getParameters().size());
        assertEquals(20, pstmt.getParameters().get(0).getValue());
    }

    @Test
    @DisplayName("ExecuteQuery should throw SQLException (not implemented)")
    void testExecuteQuery() {
        assertThrows(SQLException.class, () -> pstmt.executeQuery());
    }

    @Test
    @DisplayName("ExecuteUpdate should throw SQLException (not implemented)")
    void testExecuteUpdate() {
        assertThrows(SQLException.class, () -> pstmt.executeUpdate());
    }

    @Test
    @DisplayName("Execute should throw SQLException (not implemented)")
    void testExecute() {
        assertThrows(SQLException.class, () -> pstmt.execute());
    }

    @Test
    @DisplayName("GetGeneratedKeys should return empty ResultSet")
    void testGetGeneratedKeys() throws SQLException {
        ResultSet rs = pstmt.getGeneratedKeys();
        assertNotNull(rs);
        assertFalse(rs.next()); // Empty result set
    }

    @Test
    @DisplayName("GetMetaData should return null")
    void testGetMetaData() throws SQLException {
        assertNull(pstmt.getMetaData());
    }

    @Test
    @DisplayName("GetParameterMetaData should return null")
    void testGetParameterMetaData() throws SQLException {
        assertNull(pstmt.getParameterMetaData());
    }

    @Test
    @DisplayName("AddBatch should throw SQLFeatureNotSupportedException")
    void testAddBatch() {
        assertThrows(SQLFeatureNotSupportedException.class, () -> pstmt.addBatch());
    }

    @Test
    @DisplayName("SetObject should work for basic types")
    void testSetObject() throws SQLException {
        pstmt.setObject(1, "test");
        pstmt.setObject(2, 123);
        pstmt.setObject(3, 45.67);
    }

    @Test
    @DisplayName("SetBlob should throw SQLFeatureNotSupportedException")
    void testSetBlob() {
        assertThrows(SQLFeatureNotSupportedException.class, 
            () -> pstmt.setBlob(1, (Blob) null));
    }

    @Test
    @DisplayName("SetClob should throw SQLFeatureNotSupportedException")
    void testSetClob() {
        assertThrows(SQLFeatureNotSupportedException.class, 
            () -> pstmt.setClob(1, (Clob) null));
    }

    @Test
    @DisplayName("SetArray should throw SQLFeatureNotSupportedException")
    void testSetArray() {
        assertThrows(SQLFeatureNotSupportedException.class, 
            () -> pstmt.setArray(1, null));
    }

    @Test
    @DisplayName("SetRef should throw SQLFeatureNotSupportedException")
    void testSetRef() {
        assertThrows(SQLFeatureNotSupportedException.class, 
            () -> pstmt.setRef(1, null));
    }

    @Test
    @DisplayName("SetURL should throw SQLFeatureNotSupportedException")
    void testSetURL() {
        assertThrows(SQLFeatureNotSupportedException.class, 
            () -> pstmt.setURL(1, null));
    }

    @Test
    @DisplayName("SetRowId should throw SQLFeatureNotSupportedException")
    void testSetRowId() {
        assertThrows(SQLFeatureNotSupportedException.class, 
            () -> pstmt.setRowId(1, null));
    }

    @Test
    @DisplayName("SetSQLXML should throw SQLFeatureNotSupportedException")
    void testSetSQLXML() {
        assertThrows(SQLFeatureNotSupportedException.class, 
            () -> pstmt.setSQLXML(1, null));
    }

    @Test
    @DisplayName("Operations on closed statement should throw SQLException")
    void testOperationsAfterClose() throws SQLException {
        pstmt.close();
        
        assertThrows(SQLException.class, () -> pstmt.setInt(1, 10));
        assertThrows(SQLException.class, () -> pstmt.clearParameters());
        assertThrows(SQLException.class, () -> pstmt.executeQuery());
    }

    @Test
    @DisplayName("SetNString should delegate to setString")
    void testSetNString() throws SQLException {
        pstmt.setNString(1, "test");
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("VARCHAR", param.getType());
        assertEquals("test", param.getValue());
    }

    @Test
    @DisplayName("SetDate with Calendar should delegate to setDate")
    void testSetDateWithCalendar() throws SQLException {
        java.sql.Date date = java.sql.Date.valueOf("2024-01-15");
        pstmt.setDate(1, date, Calendar.getInstance());
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("DATE", param.getType());
        assertEquals(date, param.getValue());
    }

    @Test
    @DisplayName("SetTime with Calendar should delegate to setTime")
    void testSetTimeWithCalendar() throws SQLException {
        Time time = Time.valueOf("10:30:00");
        pstmt.setTime(1, time, Calendar.getInstance());
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("TIME", param.getType());
        assertEquals(time, param.getValue());
    }

    @Test
    @DisplayName("SetTimestamp with Calendar should delegate to setTimestamp")
    void testSetTimestampWithCalendar() throws SQLException {
        Timestamp timestamp = Timestamp.valueOf("2024-01-15 10:30:00");
        pstmt.setTimestamp(1, timestamp, Calendar.getInstance());
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("TIMESTAMP", param.getType());
        assertEquals(timestamp, param.getValue());
    }

    @Test
    @DisplayName("SetNull with typeName should delegate to setNull")
    void testSetNullWithTypeName() throws SQLException {
        pstmt.setNull(1, Types.VARCHAR, "VARCHAR");
        
        MystiSqlPreparedStatement.Parameter param = pstmt.getParameters().get(0);
        assertEquals("NULL", param.getType());
        assertNull(param.getValue());
    }

    @Test
    @DisplayName("PreparedStatement should be a Statement")
    void testInheritance() {
        assertTrue(pstmt instanceof MystiSqlStatement);
    }

    @Test
    @DisplayName("Multiple parameter types in same statement")
    void testMixedParameterTypes() throws SQLException {
        pstmt.setInt(1, 42);
        pstmt.setString(2, "Alice");
        pstmt.setBoolean(3, true);
        pstmt.setDouble(4, 3.14);
        
        assertEquals(4, pstmt.getParameters().size());
        
        assertEquals("INTEGER", pstmt.getParameters().get(0).getType());
        assertEquals("VARCHAR", pstmt.getParameters().get(1).getType());
        assertEquals("BOOLEAN", pstmt.getParameters().get(2).getType());
        assertEquals("DOUBLE", pstmt.getParameters().get(3).getType());
    }
}
