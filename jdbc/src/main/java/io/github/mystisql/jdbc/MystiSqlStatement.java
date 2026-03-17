package io.github.mystisql.jdbc;

import io.github.mystisql.jdbc.client.ExecResult;
import io.github.mystisql.jdbc.client.Transport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.sql.*;

/**
 * MystiSql JDBC Statement implementation.
 */
public class MystiSqlStatement implements Statement {
    
    private static final Logger logger = LoggerFactory.getLogger(MystiSqlStatement.class);
    
    private final MystiSqlConnection connection;
    private ResultSet currentResultSet;
    private int updateCount = -1;
    private long largeUpdateCount = -1;
    private int queryTimeout = 0;
    private boolean closed = false;
    private boolean closeOnCompletion = false;
    private Long lastInsertId = null;
    
    public MystiSqlStatement(MystiSqlConnection connection) {
        this.connection = connection;
    }
    
    @Override
    public ResultSet executeQuery(String sql) throws SQLException {
        checkClosed();
        logger.debug("executeQuery: {}", sql);
        
        closeCurrentResultSet();
        
        Transport transport = connection.getTransport();
        currentResultSet = transport.executeQuery(connection.getInstanceName(), sql);
        updateCount = -1;
        largeUpdateCount = -1;
        lastInsertId = null;
        
        return currentResultSet;
    }
    
    @Override
    public int executeUpdate(String sql) throws SQLException {
        checkClosed();
        logger.debug("executeUpdate: {}", sql);
        
        closeCurrentResultSet();
        
        Transport transport = connection.getTransport();
        ExecResult result = transport.executeUpdate(connection.getInstanceName(), sql);
        
        updateCount = result.getRowsAffected() != null ? result.getRowsAffected().intValue() : 0;
        largeUpdateCount = result.getRowsAffected() != null ? result.getRowsAffected() : 0L;
        lastInsertId = result.getLastInsertId();
        currentResultSet = null;
        
        return updateCount;
    }
    
    @Override
    public boolean execute(String sql) throws SQLException {
        checkClosed();
        logger.debug("execute: {}", sql);
        
        closeCurrentResultSet();
        
        String normalizedSql = sql.trim().toUpperCase();
        boolean isQuery = normalizedSql.startsWith("SELECT") || 
                          normalizedSql.startsWith("SHOW") ||
                          normalizedSql.startsWith("DESCRIBE") ||
                          normalizedSql.startsWith("EXPLAIN");
        
        if (isQuery) {
            executeQuery(sql);
            return true;
        } else {
            executeUpdate(sql);
            return false;
        }
    }
    
    @Override
    public void close() throws SQLException {
        if (closed) return;
        closed = true;
        closeCurrentResultSet();
    }
    
    private void closeCurrentResultSet() throws SQLException {
        if (currentResultSet != null) {
            currentResultSet.close();
            currentResultSet = null;
        }
    }
    
    @Override
    public boolean isClosed() throws SQLException {
        return closed;
    }
    
    @Override
    public void setQueryTimeout(int seconds) throws SQLException {
        checkClosed();
        this.queryTimeout = seconds;
    }
    
    @Override
    public int getQueryTimeout() throws SQLException {
        checkClosed();
        return queryTimeout;
    }
    
    @Override
    public Connection getConnection() throws SQLException {
        checkClosed();
        return connection;
    }
    
    @Override
    public ResultSet getResultSet() throws SQLException {
        checkClosed();
        return currentResultSet;
    }
    
    @Override
    public int getUpdateCount() throws SQLException {
        checkClosed();
        return updateCount;
    }
    
    @Override
    public long getLargeUpdateCount() throws SQLException {
        checkClosed();
        return largeUpdateCount;
    }
    
    @Override
    public ResultSet getGeneratedKeys() throws SQLException {
        checkClosed();
        if (lastInsertId == null) {
            MystiSqlResultSet.Column[] columns = {
                new MystiSqlResultSet.Column("GENERATED_KEY", Types.BIGINT, "BIGINT")
            };
            return new MystiSqlResultSet(columns, new Object[0][]);
        }
        
        MystiSqlResultSet.Column[] columns = {
            new MystiSqlResultSet.Column("GENERATED_KEY", Types.BIGINT, "BIGINT")
        };
        Object[][] rows = { { lastInsertId } };
        return new MystiSqlResultSet(columns, rows);
    }
    
    protected void checkClosed() throws SQLException {
        if (closed) {
            throw new SQLException("Statement is closed");
        }
    }
    
    @Override
    public int getMaxFieldSize() throws SQLException { return 0; }
    
    @Override
    public void setMaxFieldSize(int max) throws SQLException { }
    
    @Override
    public int getMaxRows() throws SQLException { return 0; }
    
    @Override
    public void setMaxRows(int max) throws SQLException { }
    
    @Override
    public void setEscapeProcessing(boolean enable) throws SQLException { }
    
    @Override
    public void cancel() throws SQLException { }
    
    @Override
    public SQLWarning getWarnings() throws SQLException { return null; }
    
    @Override
    public void clearWarnings() throws SQLException { }
    
    @Override
    public void setCursorName(String name) throws SQLException { }
    
    @Override
    public boolean execute(String sql, int autoGeneratedKeys) throws SQLException {
        return execute(sql);
    }
    
    @Override
    public boolean execute(String sql, int[] columnIndexes) throws SQLException {
        return execute(sql);
    }
    
    @Override
    public boolean execute(String sql, String[] columnNames) throws SQLException {
        return execute(sql);
    }
    
    @Override
    public int executeUpdate(String sql, int autoGeneratedKeys) throws SQLException {
        return executeUpdate(sql);
    }
    
    @Override
    public int executeUpdate(String sql, int[] columnIndexes) throws SQLException {
        return executeUpdate(sql);
    }
    
    @Override
    public int executeUpdate(String sql, String[] columnNames) throws SQLException {
        return executeUpdate(sql);
    }
    
    @Override
    public void addBatch(String sql) throws SQLException {
        throw new SQLFeatureNotSupportedException("Batch operations not supported");
    }
    
    @Override
    public void clearBatch() throws SQLException { }
    
    @Override
    public int[] executeBatch() throws SQLException {
        throw new SQLFeatureNotSupportedException("Batch operations not supported");
    }
    
    @Override
    public long[] executeLargeBatch() throws SQLException {
        throw new SQLFeatureNotSupportedException("Batch operations not supported");
    }
    
    @Override
    public long executeLargeUpdate(String sql) throws SQLException {
        executeUpdate(sql);
        return largeUpdateCount;
    }
    
    @Override
    public long executeLargeUpdate(String sql, int autoGeneratedKeys) throws SQLException {
        return executeLargeUpdate(sql);
    }
    
    @Override
    public long executeLargeUpdate(String sql, int[] columnIndexes) throws SQLException {
        return executeLargeUpdate(sql);
    }
    
    @Override
    public long executeLargeUpdate(String sql, String[] columnNames) throws SQLException {
        return executeLargeUpdate(sql);
    }
    
    @Override
    public void setLargeMaxRows(long max) throws SQLException { }
    
    @Override
    public long getLargeMaxRows() throws SQLException { return 0; }
    
    @Override
    public boolean isPoolable() throws SQLException { return false; }
    
    @Override
    public void setPoolable(boolean poolable) throws SQLException { }
    
    @Override
    public void closeOnCompletion() throws SQLException {
        this.closeOnCompletion = true;
    }
    
    @Override
    public boolean isCloseOnCompletion() throws SQLException {
        return closeOnCompletion;
    }
    
    @Override
    public <T> T unwrap(Class<T> iface) throws SQLException {
        if (iface.isAssignableFrom(getClass())) {
            return iface.cast(this);
        }
        throw new SQLFeatureNotSupportedException("Not a wrapper for " + iface.getName());
    }
    
    @Override
    public boolean isWrapperFor(Class<?> iface) throws SQLException {
        return iface.isAssignableFrom(getClass());
    }
    
    @Override
    public boolean getMoreResults() throws SQLException {
        checkClosed();
        closeCurrentResultSet();
        updateCount = -1;
        largeUpdateCount = -1;
        return false;
    }
    
    @Override
    public boolean getMoreResults(int current) throws SQLException {
        return getMoreResults();
    }
    
    @Override
    public int getResultSetType() throws SQLException {
        return ResultSet.TYPE_FORWARD_ONLY;
    }
    
    @Override
    public int getResultSetConcurrency() throws SQLException {
        return ResultSet.CONCUR_READ_ONLY;
    }
    
    @Override
    public int getResultSetHoldability() throws SQLException {
        return ResultSet.HOLD_CURSORS_OVER_COMMIT;
    }
    
    @Override
    public void setFetchDirection(int direction) throws SQLException {
        checkClosed();
        if (direction != ResultSet.FETCH_FORWARD) {
            throw new SQLException("Only FETCH_FORWARD is supported");
        }
    }
    
    @Override
    public int getFetchDirection() throws SQLException {
        return ResultSet.FETCH_FORWARD;
    }
    
    @Override
    public void setFetchSize(int rows) throws SQLException {
        checkClosed();
    }
    
    @Override
    public int getFetchSize() throws SQLException {
        return 0;
    }
}
