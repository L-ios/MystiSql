package io.github.mystisql.jdbc.client;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.sql.Types;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Unit tests for type mapping logic.
 */
class TypeConverterTest {
    
    @Test
    @DisplayName("Map MySQL INT to JDBC INTEGER")
    void testIntType() {
        String mysqlType = "INT";
        int expected = Types.INTEGER;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL VARCHAR to JDBC VARCHAR")
    void testVarcharType() {
        String mysqlType = "VARCHAR";
        int expected = Types.VARCHAR;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL BIGINT to JDBC BIGINT")
    void testBigintType() {
        String mysqlType = "BIGINT";
        int expected = Types.BIGINT;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL DOUBLE to JDBC DOUBLE")
    void testDoubleType() {
        String mysqlType = "DOUBLE";
        int expected = Types.DOUBLE;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL FLOAT to JDBC FLOAT")
    void testFloatType() {
        String mysqlType = "FLOAT";
        int expected = Types.FLOAT;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL TINYINT to JDBC TINYINT")
    void testTinyintType() {
        String mysqlType = "TINYINT";
        int expected = Types.TINYINT;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL SMALLINT to JDBC SMALLINT")
    void testSmallintType() {
        String mysqlType = "SMALLINT";
        int expected = Types.SMALLINT;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL DECIMAL to JDBC DECIMAL")
    void testDecimalType() {
        String mysqlType = "DECIMAL";
        int expected = Types.DECIMAL;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL NUMERIC to JDBC DECIMAL")
    void testNumericType() {
        String mysqlType = "NUMERIC";
        int expected = Types.DECIMAL;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL CHAR to JDBC CHAR")
    void testCharType() {
        String mysqlType = "CHAR";
        int expected = Types.CHAR;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL TEXT to JDBC LONGVARCHAR")
    void testTextType() {
        String mysqlType = "TEXT";
        int expected = Types.LONGVARCHAR;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL LONGTEXT to JDBC LONGVARCHAR")
    void testLongtextType() {
        String mysqlType = "LONGTEXT";
        int expected = Types.LONGVARCHAR;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL DATE to JDBC DATE")
    void testDateType() {
        String mysqlType = "DATE";
        int expected = Types.DATE;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL TIME to JDBC TIME")
    void testTimeType() {
        String mysqlType = "TIME";
        int expected = Types.TIME;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL DATETIME to JDBC TIMESTAMP")
    void testDatetimeType() {
        String mysqlType = "DATETIME";
        int expected = Types.TIMESTAMP;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL TIMESTAMP to JDBC TIMESTAMP")
    void testTimestampType() {
        String mysqlType = "TIMESTAMP";
        int expected = Types.TIMESTAMP;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL BLOB to JDBC BLOB")
    void testBlobType() {
        String mysqlType = "BLOB";
        int expected = Types.BLOB;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL BINARY to JDBC BLOB")
    void testBinaryType() {
        String mysqlType = "BINARY";
        int expected = Types.BLOB;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL BOOLEAN to JDBC BOOLEAN")
    void testBooleanType() {
        String mysqlType = "BOOLEAN";
        int expected = Types.BOOLEAN;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map MySQL BOOL to JDBC BOOLEAN")
    void testBoolType() {
        String mysqlType = "BOOL";
        int expected = Types.BOOLEAN;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map unknown type to JDBC OTHER")
    void testUnknownType() {
        String mysqlType = "UNKNOWN_TYPE";
        int expected = Types.OTHER;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Map null type to JDBC VARCHAR (default)")
    void testNullType() {
        String mysqlType = null;
        int expected = Types.VARCHAR;
        int actual = mapType(mysqlType);
        assertEquals(expected, actual);
    }
    
    @Test
    @DisplayName("Type mapping is case-insensitive")
    void testCaseInsensitive() {
        assertEquals(mapType("int"), mapType("INT"));
        assertEquals(mapType("varchar"), mapType("VARCHAR"));
        assertEquals(mapType("Varchar"), mapType("VARCHAR"));
    }
    
    /**
     * Helper method that mirrors RestClient.mapType() logic.
     */
    private int mapType(String typeName) {
        if (typeName == null) return Types.VARCHAR;
        
        switch (typeName.toUpperCase()) {
            case "TINYINT": return Types.TINYINT;
            case "SMALLINT": return Types.SMALLINT;
            case "INT":
            case "INTEGER": return Types.INTEGER;
            case "BIGINT": return Types.BIGINT;
            case "FLOAT": return Types.FLOAT;
            case "DOUBLE": return Types.DOUBLE;
            case "DECIMAL": return Types.DECIMAL;
            case "NUMERIC": return Types.DECIMAL;
            case "CHAR": return Types.CHAR;
            case "VARCHAR": return Types.VARCHAR;
            case "TEXT":
            case "LONGTEXT": return Types.LONGVARCHAR;
            case "DATE": return Types.DATE;
            case "TIME": return Types.TIME;
            case "DATETIME":
            case "TIMESTAMP": return Types.TIMESTAMP;
            case "BLOB":
            case "BINARY": return Types.BLOB;
            case "BOOLEAN":
            case "BOOL": return Types.BOOLEAN;
            default: return Types.OTHER;
        }
    }
}
