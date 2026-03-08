package io.github.mystisql.jdbc;

import io.github.mystisql.jdbc.client.RestClient;
import java.sql.*;
import java.util.Map;
import java.util.Properties;
import java.util.concurrent.Executor;

/**
 * MystiSql JDBC Connection implementation.
 */
public class MystiSqlConnection implements Connection {
    
    private final String host;
    private final int port;
    private final String instanceName;
    private final String username;
    private final String token;
    private final int timeout;
    private final boolean ssl;
    private final boolean verifySsl;
    private final int maxConnections;
    private final RestClient restClient;
    
    private boolean closed = false;
    private boolean autoCommit = true;
    
    public MystiSqlConnection(String host, int port, String instanceName, String username, 
                              String token, int timeout, boolean ssl, boolean verifySsl, 
                              int maxConnections) {
        this.host = host;
        this.port = port;
        this.instanceName = instanceName;
        this.username = username;
        this.token = token;
        this.timeout = timeout;
        this.ssl = ssl;
        this.verifySsl = verifySsl;
        this.maxConnections = maxConnections;
        
        // Initialize REST client
        String protocol = ssl ? "https" : "http";
        String baseUrl = String.format("%s://%s:%d", protocol, host, port);
        this.restClient = new RestClient(baseUrl, token, timeout);
    }
    
    @Override
    public Statement createStatement() throws SQLException {
        checkClosed();
        return new MystiSqlStatement(this);
    }
    
    @Override
    public PreparedStatement prepareStatement(String sql) throws SQLException {
        checkClosed();
        return new MystiSqlPreparedStatement(this, sql);
    }
    
    @Override
    public void close() throws SQLException {
        if (closed) {
            return;
        }
        closed = true;
        if (restClient != null) {
            restClient.close();
        }
    }
    
    @Override
    public boolean isClosed() throws SQLException {
        return closed;
    }
    
    @Override
    public DatabaseMetaData getMetaData() throws SQLException {
        checkClosed();
        return new MystiSqlDatabaseMetaData(this);
    }
    
    @Override
    public void setAutoCommit(boolean autoCommit) throws SQLException {
        checkClosed();
        this.autoCommit = autoCommit;
    }
    
    @Override
    public boolean getAutoCommit() throws SQLException {
        checkClosed();
        return autoCommit;
    }
    
    @Override
    public void commit() throws SQLException {
        checkClosed();
        if (autoCommit) {
            throw new SQLException("Cannot commit when auto-commit is enabled");
        }
        throw new SQLFeatureNotSupportedException("Transaction management not supported in Phase 2.5");
    }
    
    @Override
    public void rollback() throws SQLException {
        checkClosed();
        if (autoCommit) {
            throw new SQLException("Cannot rollback when auto-commit is enabled");
        }
        throw new SQLFeatureNotSupportedException("Transaction management not supported in Phase 2.5");
    }
    
    @Override
    public boolean isValid(int timeout) throws SQLException {
        if (timeout < 0) {
            throw new SQLException("Timeout must be >= 0");
        }
        if (closed) {
            return false;
        }
        try {
            return restClient.healthCheck(instanceName);
        } catch (Exception e) {
            return false;
        }
    }
    
    private void checkClosed() throws SQLException {
        if (closed) {
            throw new SQLException("Connection is closed");
        }
    }
    
    // Getters
    public String getHost() { return host; }
    public int getPort() { return port; }
    public String getInstanceName() { return instanceName; }
    public String getUsername() { return username; }
    public String getToken() { return token; }
    public int getTimeout() { return timeout; }
    public boolean isSsl() { return ssl; }
    public RestClient getRestClient() { return restClient; }
    
    // Placeholder implementations for remaining methods
    @Override public Statement createStatement(int resultSetType, int resultSetConcurrency) throws SQLException { return createStatement(); }
    @Override public PreparedStatement prepareStatement(String sql, int resultSetType, int resultSetConcurrency) throws SQLException { return prepareStatement(sql); }
    @Override public CallableStatement prepareCall(String sql) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public String nativeSQL(String sql) throws SQLException { return sql; }
    @Override public Statement createStatement(int resultSetType, int resultSetConcurrency, int resultSetHoldability) throws SQLException { return createStatement(); }
    @Override public PreparedStatement prepareStatement(String sql, int resultSetType, int resultSetConcurrency, int resultSetHoldability) throws SQLException { return prepareStatement(sql); }
    @Override public CallableStatement prepareCall(String sql, int resultSetType, int resultSetConcurrency) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public PreparedStatement prepareStatement(String sql, int autoGeneratedKeys) throws SQLException { return prepareStatement(sql); }
    @Override public PreparedStatement prepareStatement(String sql, int[] columnIndexes) throws SQLException { return prepareStatement(sql); }
    @Override public PreparedStatement prepareStatement(String sql, String[] columnNames) throws SQLException { return prepareStatement(sql); }
    @Override public Clob createClob() throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public Blob createBlob() throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public NClob createNClob() throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public SQLXML createSQLXML() throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setClientInfo(String name, String value) throws SQLClientInfoException { }
    @Override public void setClientInfo(Properties properties) throws SQLClientInfoException { }
    @Override public String getClientInfo(String name) throws SQLException { return null; }
    @Override public Properties getClientInfo() throws SQLException { return new Properties(); }
    @Override public Array createArrayOf(String typeName, Object[] elements) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public Struct createStruct(String typeName, Object[] attributes) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setSchema(String schema) throws SQLException { }
    @Override public String getSchema() throws SQLException { return null; }
    @Override public void abort(Executor executor) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setNetworkTimeout(Executor executor, int milliseconds) throws SQLException { }
    @Override public int getNetworkTimeout() throws SQLException { return 0; }
    @Override public <T> T unwrap(Class<T> iface) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public boolean isWrapperFor(Class<?> iface) throws SQLException { return false; }
    @Override public Map<String, Class<?>> getTypeMap() throws SQLException { return null; }
    @Override public void setTypeMap(Map<String, Class<?>> map) throws SQLException { }
    @Override public void setHoldability(int holdability) throws SQLException { }
    @Override public int getHoldability() throws SQLException { return ResultSet.HOLD_CURSORS_OVER_COMMIT; }
    @Override public Savepoint setSavepoint() throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public Savepoint setSavepoint(String name) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void rollback(Savepoint savepoint) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void releaseSavepoint(Savepoint savepoint) throws SQLException { throw new SQLFeatureNotSupportedException(); }
}
