package io.github.mystisql.jdbc;

import java.sql.*;
import java.util.Properties;
import java.util.logging.Logger;

/**
 * MystiSql JDBC Driver implementation.
 * 
 * <p>This driver allows Java applications to connect to MystiSql Gateway
 * and access databases in Kubernetes clusters.</p>
 * 
 * <p>URL format: jdbc:mystisql://gateway-host:port/instance-name?params</p>
 * 
 * <p>Example:</p>
 * <pre>
 * jdbc:mystisql://gateway.example.com:8080/production-mysql
 * jdbc:mystisql://localhost:8080/test-db?timeout=60&ssl=true
 * </pre>
 */
public class MystiSqlDriver implements Driver {
    
    private static final String URL_PREFIX = "jdbc:mystisql://";
    private static final int MAJOR_VERSION = 1;
    private static final int MINOR_VERSION = 0;
    
    static {
        try {
            DriverManager.registerDriver(new MystiSqlDriver());
        } catch (SQLException e) {
            throw new RuntimeException("Failed to register MystiSqlDriver", e);
        }
    }
    
    @Override
    public Connection connect(String url, Properties info) throws SQLException {
        if (!acceptsURL(url)) {
            throw new SQLException("Invalid URL format: " + url);
        }
        
        // Parse URL
        String host = parseHost(url);
        int port = parsePort(url);
        String instanceName = parseInstanceName(url);
        
        if (instanceName == null || instanceName.isEmpty()) {
            throw new SQLException("Instance name is required in URL");
        }
        
        // Extract token from URL or properties
        String token = parseParameter(url, "token");
        if (token == null && info != null) {
            // Try to get token from password property
            token = info.getProperty("password");
        }
        
        // Extract username
        String username = info != null ? info.getProperty("user", "") : "";
        
        // Parse additional parameters
        String timeoutStr = parseParameter(url, "timeout");
        int timeout = timeoutStr != null ? Integer.parseInt(timeoutStr) : 30;
        
        String sslStr = parseParameter(url, "ssl");
        boolean ssl = sslStr != null && Boolean.parseBoolean(sslStr);
        
        String verifySslStr = parseParameter(url, "verifySsl");
        boolean verifySsl = verifySslStr == null || Boolean.parseBoolean(verifySslStr);
        
        String maxConnectionsStr = parseParameter(url, "maxConnections");
        int maxConnections = maxConnectionsStr != null ? Integer.parseInt(maxConnectionsStr) : 20;
        
        // Create connection
        return new MystiSqlConnection(host, port, instanceName, username, token, timeout, ssl, verifySsl, maxConnections);
    }
    
    @Override
    public boolean acceptsURL(String url) throws SQLException {
        return url != null && url.startsWith(URL_PREFIX);
    }
    
    @Override
    public DriverPropertyInfo[] getPropertyInfo(String url, Properties info) throws SQLException {
        DriverPropertyInfo[] props = new DriverPropertyInfo[5];
        
        props[0] = new DriverPropertyInfo("token", null);
        props[0].description = "Authentication token for MystiSql Gateway";
        props[0].required = false;
        
        props[1] = new DriverPropertyInfo("timeout", "30");
        props[1].description = "Query timeout in seconds";
        props[1].required = false;
        
        props[2] = new DriverPropertyInfo("ssl", "false");
        props[2].description = "Enable HTTPS";
        props[2].required = false;
        
        props[3] = new DriverPropertyInfo("verifySsl", "true");
        props[3].description = "Verify SSL certificate";
        props[3].required = false;
        
        props[4] = new DriverPropertyInfo("maxConnections", "20");
        props[4].description = "Maximum HTTP connections in pool";
        props[4].required = false;
        
        return props;
    }
    
    @Override
    public int getMajorVersion() {
        return MAJOR_VERSION;
    }
    
    @Override
    public int getMinorVersion() {
        return MINOR_VERSION;
    }
    
    @Override
    public boolean jdbcCompliant() {
        // Phase 2.5: Not fully JDBC compliant yet
        // Phase 3: Will return true after implementing all features
        return false;
    }
    
    @Override
    public Logger getParentLogger() throws SQLFeatureNotSupportedException {
        throw new SQLFeatureNotSupportedException("getParentLogger() not supported");
    }
    
    /**
     * Parse host from JDBC URL.
     * 
     * @param url JDBC URL
     * @return host name or IP
     */
    public String parseHost(String url) throws SQLException {
        if (!acceptsURL(url)) {
            throw new SQLException("Invalid MystiSql URL format");
        }
        
        String remainder = url.substring(URL_PREFIX.length());
        int portIndex = remainder.indexOf(':');
        int pathIndex = remainder.indexOf('/');
        
        if (portIndex > 0) {
            return remainder.substring(0, portIndex);
        } else if (pathIndex > 0) {
            return remainder.substring(0, pathIndex);
        } else {
            throw new SQLException("Invalid URL format: missing instance name");
        }
    }
    
    /**
     * Parse port from JDBC URL.
     * 
     * @param url JDBC URL
     * @return port number (default 8080)
     */
    public int parsePort(String url) throws SQLException {
        if (!acceptsURL(url)) {
            throw new SQLException("Invalid MystiSql URL format");
        }
        
        String remainder = url.substring(URL_PREFIX.length());
        int portIndex = remainder.indexOf(':');
        
        if (portIndex > 0) {
            int pathIndex = remainder.indexOf('/', portIndex);
            if (pathIndex > portIndex) {
                return Integer.parseInt(remainder.substring(portIndex + 1, pathIndex));
            } else {
                int queryIndex = remainder.indexOf('?', portIndex);
                if (queryIndex > portIndex) {
                    return Integer.parseInt(remainder.substring(portIndex + 1, queryIndex));
                } else {
                    return Integer.parseInt(remainder.substring(portIndex + 1));
                }
            }
        }
        
        // Default port
        return 8080;
    }
    
    /**
     * Parse instance name from JDBC URL.
     * 
     * @param url JDBC URL
     * @return instance name
     */
    public String parseInstanceName(String url) throws SQLException {
        if (!acceptsURL(url)) {
            throw new SQLException("Invalid MystiSql URL format");
        }
        
        String remainder = url.substring(URL_PREFIX.length());
        int pathIndex = remainder.indexOf('/');
        
        if (pathIndex < 0) {
            throw new SQLException("Invalid URL format: missing instance name");
        }
        
        int queryIndex = remainder.indexOf('?', pathIndex);
        if (queryIndex > pathIndex) {
            return remainder.substring(pathIndex + 1, queryIndex);
        } else {
            return remainder.substring(pathIndex + 1);
        }
    }
    
    /**
     * Parse parameter from JDBC URL query string.
     * 
     * @param url JDBC URL
     * @param paramName parameter name
     * @return parameter value or null if not found
     */
    public String parseParameter(String url, String paramName) throws SQLException {
        if (!acceptsURL(url)) {
            throw new SQLException("Invalid MystiSql URL format");
        }
        
        int queryIndex = url.indexOf('?');
        if (queryIndex < 0) {
            return null;
        }
        
        String queryString = url.substring(queryIndex + 1);
        String[] pairs = queryString.split("&");
        
        for (String pair : pairs) {
            String[] keyValue = pair.split("=", 2);
            if (keyValue.length == 2 && keyValue[0].equals(paramName)) {
                return keyValue[1];
            }
        }
        
        return null;
    }
}
