package io.github.mystisql.jdbc;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;
import static org.junit.jupiter.api.Assertions.*;

import java.sql.Driver;
import java.sql.DriverManager;
import java.sql.SQLException;
import java.sql.DriverPropertyInfo;

/**
 * Unit tests for MystiSqlDriver
 */
class MystiSqlDriverTest {

    @Test
    @DisplayName("Driver should be registered via SPI")
    void testDriverRegistration() throws SQLException {
        // When: Loading driver via DriverManager
        Driver driver = DriverManager.getDriver("jdbc:mystisql://localhost:8080/test");
        
        // Then: Driver should be MystiSqlDriver
        assertNotNull(driver);
        assertTrue(driver instanceof MystiSqlDriver);
    }

    @Test
    @DisplayName("Driver should accept MystiSql URLs")
    void testAcceptsURL() throws SQLException {
        // Given: A MystiSqlDriver instance
        MystiSqlDriver driver = new MystiSqlDriver();
        
        // Then: Should accept MystiSql URLs
        assertTrue(driver.acceptsURL("jdbc:mystisql://localhost:8080/test"));
        assertTrue(driver.acceptsURL("jdbc:mystisql://gateway.example.com:8080/production-mysql"));
        assertTrue(driver.acceptsURL("jdbc:mystisql://192.168.1.100/instance?timeout=60"));
        
        // And: Should NOT accept non-MystiSql URLs
        assertFalse(driver.acceptsURL("jdbc:mysql://localhost:3306/test"));
        assertFalse(driver.acceptsURL("jdbc:postgresql://localhost:5432/test"));
        assertFalse(driver.acceptsURL("http://localhost:8080"));
        assertFalse(driver.acceptsURL("invalid-url"));
    }

    @Test
    @DisplayName("Driver should parse basic URL correctly")
    void testParseBasicURL() throws SQLException {
        // Given: A MystiSqlDriver instance
        MystiSqlDriver driver = new MystiSqlDriver();
        
        // When: Parsing a basic URL
        String url = "jdbc:mystisql://gateway.example.com:8080/production-mysql";
        
        // Then: Should parse host, port, instance correctly
        assertEquals("gateway.example.com", driver.parseHost(url));
        assertEquals(8080, driver.parsePort(url));
        assertEquals("production-mysql", driver.parseInstanceName(url));
    }

    @Test
    @DisplayName("Driver should parse URL with parameters")
    void testParseURLWithParameters() throws SQLException {
        // Given: A MystiSqlDriver instance
        MystiSqlDriver driver = new MystiSqlDriver();
        
        // When: Parsing URL with parameters
        String url = "jdbc:mystisql://localhost:8080/test-db?timeout=60&ssl=true&token=abc123";
        
        // Then: Should parse parameters correctly
        assertEquals("localhost", driver.parseHost(url));
        assertEquals(8080, driver.parsePort(url));
        assertEquals("test-db", driver.parseInstanceName(url));
        assertEquals("60", driver.parseParameter(url, "timeout"));
        assertEquals("true", driver.parseParameter(url, "ssl"));
        assertEquals("abc123", driver.parseParameter(url, "token"));
    }

    @Test
    @DisplayName("Driver should reject invalid URLs")
    void testInvalidURL() {
        // Given: A MystiSqlDriver instance
        MystiSqlDriver driver = new MystiSqlDriver();
        
        // Then: Should reject invalid URLs
        assertThrows(SQLException.class, () -> driver.connect("jdbc:mysql://localhost:3306/test", null));
        assertThrows(SQLException.class, () -> driver.connect("invalid-url", null));
        assertThrows(SQLException.class, () -> driver.connect("jdbc:mystisql://", null));
        assertThrows(SQLException.class, () -> driver.connect("jdbc:mystisql://localhost", null));
    }

    @Test
    @DisplayName("Driver should return correct version info")
    void testVersionInfo() throws SQLException {
        // Given: A MystiSqlDriver instance
        MystiSqlDriver driver = new MystiSqlDriver();
        
        // Then: Should return version info
        assertTrue(driver.getMajorVersion() >= 1);
        assertTrue(driver.getMinorVersion() >= 0);
    }

    @Test
    @DisplayName("Driver should return property info")
    void testGetPropertyInfo() throws SQLException {
        // Given: A MystiSqlDriver instance
        MystiSqlDriver driver = new MystiSqlDriver();
        
        // When: Getting property info
        DriverPropertyInfo[] props = driver.getPropertyInfo("jdbc:mystisql://localhost:8080/test", null);
        
        // Then: Should include common properties
        assertNotNull(props);
        assertTrue(props.length > 0);
        
        // Should have token property
        boolean hasTokenProperty = false;
        for (DriverPropertyInfo prop : props) {
            if ("token".equals(prop.name)) {
                hasTokenProperty = true;
                break;
            }
        }
        assertTrue(hasTokenProperty, "Should have 'token' property");
    }

    @Test
    @DisplayName("Driver should report JDBC compliance")
    void testJDBCCompliance() throws SQLException {
        // Given: A MystiSqlDriver instance
        MystiSqlDriver driver = new MystiSqlDriver();
        
        // Then: Phase 2.5 should return false (not fully compliant)
        assertFalse(driver.jdbcCompliant());
    }
}
