package io.github.mystisql.jdbc;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import static org.junit.jupiter.api.Assertions.*;

import java.sql.SQLException;
import java.sql.Types;
import java.util.Arrays;
import java.util.Collections;

/**
 * Unit tests for MystiSqlResultSet
 */
class MystiSqlResultSetTest {

    private MystiSqlResultSet resultSet;
    private MystiSqlResultSet emptyResultSet;

    @BeforeEach
    void setUp() {
        // Create a result set with sample data
        MystiSqlResultSet.Column[] columns = {
            new MystiSqlResultSet.Column("id", Types.INTEGER, "INT"),
            new MystiSqlResultSet.Column("name", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("age", Types.INTEGER, "INT"),
            new MystiSqlResultSet.Column("active", Types.BOOLEAN, "BOOLEAN")
        };

        Object[][] rows = {
            {1, "Alice", 25, true},
            {2, "Bob", 30, false},
            {3, "Charlie", 35, true}
        };

        resultSet = new MystiSqlResultSet(columns, rows);

        // Create empty result set
        emptyResultSet = new MystiSqlResultSet(columns, new Object[0][]);
    }

    @Test
    @DisplayName("Next should move cursor correctly")
    void testNext() throws SQLException {
        // Initially before first row
        assertFalse(resultSet.isBeforeFirst());
        
        // Move to first row
        assertTrue(resultSet.next());
        assertTrue(resultSet.isFirst());
        
        // Move to second row
        assertTrue(resultSet.next());
        
        // Move to third row
        assertTrue(resultSet.next());
        assertTrue(resultSet.isLast());
        
        // No more rows
        assertFalse(resultSet.next());
        assertTrue(resultSet.isAfterLast());
    }

    @Test
    @DisplayName("Empty result set should return false on first next()")
    void testEmptyResultSet() throws SQLException {
        assertFalse(emptyResultSet.next());
    }

    @Test
    @DisplayName("GetString by column name should work")
    void testGetStringByName() throws SQLException {
        resultSet.next();
        assertEquals("Alice", resultSet.getString("name"));
        
        resultSet.next();
        assertEquals("Bob", resultSet.getString("name"));
    }

    @Test
    @DisplayName("GetString by column index should work (1-based)")
    void testGetStringByIndex() throws SQLException {
        resultSet.next();
        assertEquals("Alice", resultSet.getString(2));
    }

    @Test
    @DisplayName("GetInt should work")
    void testGetInt() throws SQLException {
        resultSet.next();
        assertEquals(1, resultSet.getInt("id"));
        assertEquals(25, resultSet.getInt("age"));
        assertEquals(1, resultSet.getInt(1));
    }

    @Test
    @DisplayName("GetLong should work")
    void testGetLong() throws SQLException {
        resultSet.next();
        assertEquals(1L, resultSet.getLong("id"));
        assertEquals(1L, resultSet.getLong(1));
    }

    @Test
    @DisplayName("GetBoolean should work")
    void testGetBoolean() throws SQLException {
        resultSet.next();
        assertTrue(resultSet.getBoolean("active"));
        
        resultSet.next();
        assertFalse(resultSet.getBoolean("active"));
    }

    @Test
    @DisplayName("GetObject should return the raw value")
    void testGetObject() throws SQLException {
        resultSet.next();
        assertEquals(1, resultSet.getObject("id"));
        assertEquals("Alice", resultSet.getObject("name"));
        assertEquals(25, resultSet.getObject("age"));
    }

    @Test
    @DisplayName("WasNull should detect NULL values")
    void testWasNull() throws SQLException {
        // Create result set with null value
        MystiSqlResultSet.Column[] columns = {
            new MystiSqlResultSet.Column("id", Types.INTEGER, "INT"),
            new MystiSqlResultSet.Column("name", Types.VARCHAR, "VARCHAR")
        };
        Object[][] rows = {
            {1, null}
        };
        MystiSqlResultSet rsWithNull = new MystiSqlResultSet(columns, rows);
        
        rsWithNull.next();
        String name = rsWithNull.getString("name");
        assertNull(name);
        assertTrue(rsWithNull.wasNull());
        
        int id = rsWithNull.getInt("id");
        assertEquals(1, id);
        assertFalse(rsWithNull.wasNull());
    }

    @Test
    @DisplayName("GetMetaData should return metadata")
    void testGetMetaData() throws SQLException {
        var metaData = resultSet.getMetaData();
        assertNotNull(metaData);
        assertEquals(4, metaData.getColumnCount());
        assertEquals("id", metaData.getColumnName(1));
        assertEquals("name", metaData.getColumnName(2));
    }

    @Test
    @DisplayName("FindColumn should return correct column index")
    void testFindColumn() throws SQLException {
        assertEquals(1, resultSet.findColumn("id"));
        assertEquals(2, resultSet.findColumn("name"));
        assertEquals(3, resultSet.findColumn("age"));
        assertEquals(4, resultSet.findColumn("active"));
    }

    @Test
    @DisplayName("FindColumn should throw for invalid column")
    void testFindColumnInvalid() {
        assertThrows(SQLException.class, () -> resultSet.findColumn("invalid"));
    }

    @Test
    @DisplayName("GetXXX with invalid column index should throw")
    void testInvalidColumnIndex() throws SQLException {
        resultSet.next();
        assertThrows(SQLException.class, () -> resultSet.getString(0)); // 0 is invalid
        assertThrows(SQLException.class, () -> resultSet.getString(10)); // too large
    }

    @Test
    @DisplayName("GetXXX before next() should throw")
    void testGetBeforeNext() {
        assertThrows(SQLException.class, () -> resultSet.getString("id"));
    }

    @Test
    @DisplayName("Close should mark result set as closed")
    void testClose() throws SQLException {
        assertFalse(resultSet.isClosed());
        resultSet.close();
        assertTrue(resultSet.isClosed());
    }

    @Test
    @DisplayName("Operations on closed result set should throw")
    void testOperationsAfterClose() throws SQLException {
        resultSet.close();
        assertThrows(SQLException.class, () -> resultSet.next());
        assertThrows(SQLException.class, () -> resultSet.getString("id"));
    }

    @Test
    @DisplayName("GetRow should return current row number")
    void testGetRow() throws SQLException {
        assertEquals(0, resultSet.getRow()); // Before first
        
        resultSet.next();
        assertEquals(1, resultSet.getRow());
        
        resultSet.next();
        assertEquals(2, resultSet.getRow());
        
        resultSet.next();
        assertEquals(3, resultSet.getRow());
    }

    @Test
    @DisplayName("GetType should return TYPE_FORWARD_ONLY")
    void testGetType() throws SQLException {
        assertEquals(java.sql.ResultSet.TYPE_FORWARD_ONLY, resultSet.getType());
    }

    @Test
    @DisplayName("GetStatement should return null (no parent statement)")
    void testGetStatement() throws SQLException {
        assertNull(resultSet.getStatement());
    }

    @Test
    @DisplayName("Null values should return 0 for getInt")
    void testNullIntConversion() throws SQLException {
        MystiSqlResultSet.Column[] columns = {
            new MystiSqlResultSet.Column("value", Types.INTEGER, "INT")
        };
        Object[][] rows = {{null}};
        MystiSqlResultSet rs = new MystiSqlResultSet(columns, rows);
        
        rs.next();
        assertEquals(0, rs.getInt("value"));
        assertTrue(rs.wasNull());
    }
}
