package io.github.mystisql.jdbc;

import io.github.mystisql.jdbc.client.ExecResult;
import io.github.mystisql.jdbc.client.QueryParameter;
import io.github.mystisql.jdbc.client.QueryRequest;
import io.github.mystisql.jdbc.client.RestClient;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.InputStream;
import java.io.Reader;
import java.math.BigDecimal;
import java.net.URL;
import java.sql.*;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.List;

/**
 * MystiSql JDBC PreparedStatement implementation.
 */
public class MystiSqlPreparedStatement extends MystiSqlStatement implements PreparedStatement {
    
    private static final Logger logger = LoggerFactory.getLogger(MystiSqlPreparedStatement.class);
    
    private final String sql;
    private final List<Parameter> parameters;
    
    public MystiSqlPreparedStatement(MystiSqlConnection connection, String sql) {
        super(connection);
        this.sql = sql;
        this.parameters = new ArrayList<>();
    }
    
    @Override
    public ResultSet executeQuery() throws SQLException {
        checkClosed();
        logger.debug("executeQuery (prepared): {}", sql);
        
        QueryRequest request = buildQueryRequest();
        RestClient client = getConnection().getRestClient();
        
        return client.executeQuery(request);
    }
    
    @Override
    public int executeUpdate() throws SQLException {
        checkClosed();
        logger.debug("executeUpdate (prepared): {}", sql);
        
        QueryRequest request = buildQueryRequest();
        RestClient client = getConnection().getRestClient();
        
        ExecResult result = client.executeUpdate(request);
        
        int affectedRows = result.getRowsAffected() != null ? result.getRowsAffected().intValue() : 0;
        setUpdateCount(affectedRows, result.getRowsAffected() != null ? result.getRowsAffected() : 0L);
        setLastInsertId(result.getLastInsertId());
        
        return affectedRows;
    }
    
    @Override
    public boolean execute() throws SQLException {
        checkClosed();
        logger.debug("execute (prepared): {}", sql);
        
        String normalizedSql = sql.trim().toUpperCase();
        boolean isQuery = normalizedSql.startsWith("SELECT") || 
                          normalizedSql.startsWith("SHOW") ||
                          normalizedSql.startsWith("DESCRIBE") ||
                          normalizedSql.startsWith("EXPLAIN");
        
        if (isQuery) {
            executeQuery();
            return true;
        } else {
            executeUpdate();
            return false;
        }
    }
    
    private QueryRequest buildQueryRequest() throws SQLException {
        QueryRequest request = new QueryRequest();
        request.setInstance(getConnection().getInstanceName());
        request.setQuery(sql);
        
        if (!parameters.isEmpty()) {
            List<QueryParameter> queryParams = new ArrayList<>();
            for (int i = 0; i < parameters.size(); i++) {
                Parameter p = parameters.get(i);
                if (p != null) {
                    QueryParameter qp = new QueryParameter();
                    qp.setType(p.getType());
                    qp.setValue(p.getValue());
                    queryParams.add(qp);
                } else {
                    queryParams.add(new QueryParameter("NULL", null));
                }
            }
            request.setParameters(queryParams);
        }
        
        return request;
    }
    
    private void setUpdateCount(int count, long largeCount) {
    }
    
    private void setLastInsertId(Long id) {
    }
    
    @Override
    public void setNull(int parameterIndex, int sqlType) throws SQLException {
        setParameter(parameterIndex, new Parameter("NULL", null));
    }
    
    @Override
    public void setBoolean(int parameterIndex, boolean x) throws SQLException {
        setParameter(parameterIndex, new Parameter("BOOLEAN", x));
    }
    
    @Override
    public void setByte(int parameterIndex, byte x) throws SQLException {
        setParameter(parameterIndex, new Parameter("TINYINT", x));
    }
    
    @Override
    public void setShort(int parameterIndex, short x) throws SQLException {
        setParameter(parameterIndex, new Parameter("SMALLINT", x));
    }
    
    @Override
    public void setInt(int parameterIndex, int x) throws SQLException {
        setParameter(parameterIndex, new Parameter("INTEGER", x));
    }
    
    @Override
    public void setLong(int parameterIndex, long x) throws SQLException {
        setParameter(parameterIndex, new Parameter("BIGINT", x));
    }
    
    @Override
    public void setFloat(int parameterIndex, float x) throws SQLException {
        setParameter(parameterIndex, new Parameter("FLOAT", x));
    }
    
    @Override
    public void setDouble(int parameterIndex, double x) throws SQLException {
        setParameter(parameterIndex, new Parameter("DOUBLE", x));
    }
    
    @Override
    public void setBigDecimal(int parameterIndex, BigDecimal x) throws SQLException {
        setParameter(parameterIndex, new Parameter("DECIMAL", x));
    }
    
    @Override
    public void setString(int parameterIndex, String x) throws SQLException {
        setParameter(parameterIndex, new Parameter("VARCHAR", x));
    }
    
    @Override
    public void setBytes(int parameterIndex, byte[] x) throws SQLException {
        setParameter(parameterIndex, new Parameter("BINARY", x));
    }
    
    @Override
    public void setDate(int parameterIndex, Date x) throws SQLException {
        setParameter(parameterIndex, new Parameter("DATE", x));
    }
    
    @Override
    public void setTime(int parameterIndex, Time x) throws SQLException {
        setParameter(parameterIndex, new Parameter("TIME", x));
    }
    
    @Override
    public void setTimestamp(int parameterIndex, Timestamp x) throws SQLException {
        setParameter(parameterIndex, new Parameter("TIMESTAMP", x));
    }
    
    @Override
    public void clearParameters() throws SQLException {
        checkClosed();
        parameters.clear();
    }
    
    private void setParameter(int index, Parameter parameter) throws SQLException {
        if (index < 1) {
            throw new SQLException("Parameter index must be >= 1");
        }
        
        while (parameters.size() < index) {
            parameters.add(null);
        }
        
        parameters.set(index - 1, parameter);
    }
    
    public String getSql() {
        return sql;
    }
    
    public List<Parameter> getParameters() {
        return parameters;
    }
    
    @Override
    public MystiSqlConnection getConnection() throws SQLException {
        return (MystiSqlConnection) super.getConnection();
    }
    
    public static class Parameter {
        private final String type;
        private final Object value;
        
        public Parameter(String type, Object value) {
            this.type = type;
            this.value = value;
        }
        
        public String getType() { return type; }
        public Object getValue() { return value; }
    }
    
    @Override
    public void setAsciiStream(int parameterIndex, InputStream x, int length) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setUnicodeStream(int parameterIndex, InputStream x, int length) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setBinaryStream(int parameterIndex, InputStream x, int length) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setObject(int parameterIndex, Object x, int targetSqlType) throws SQLException {
        if (x == null) {
            setNull(parameterIndex, targetSqlType);
        } else if (x instanceof String) {
            setString(parameterIndex, (String) x);
        } else if (x instanceof Integer) {
            setInt(parameterIndex, (Integer) x);
        } else if (x instanceof Long) {
            setLong(parameterIndex, (Long) x);
        } else if (x instanceof Double) {
            setDouble(parameterIndex, (Double) x);
        } else if (x instanceof Float) {
            setFloat(parameterIndex, (Float) x);
        } else if (x instanceof Boolean) {
            setBoolean(parameterIndex, (Boolean) x);
        } else if (x instanceof BigDecimal) {
            setBigDecimal(parameterIndex, (BigDecimal) x);
        } else if (x instanceof Date) {
            setDate(parameterIndex, (Date) x);
        } else if (x instanceof Time) {
            setTime(parameterIndex, (Time) x);
        } else if (x instanceof Timestamp) {
            setTimestamp(parameterIndex, (Timestamp) x);
        } else if (x instanceof byte[]) {
            setBytes(parameterIndex, (byte[]) x);
        } else {
            setString(parameterIndex, x.toString());
        }
    }
    
    @Override
    public void setObject(int parameterIndex, Object x) throws SQLException {
        setObject(parameterIndex, x, Types.OTHER);
    }
    
    @Override
    public void addBatch() throws SQLException {
        throw new SQLFeatureNotSupportedException("Batch operations not supported");
    }
    
    @Override
    public void setCharacterStream(int parameterIndex, Reader reader, int length) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setRef(int parameterIndex, Ref x) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setBlob(int parameterIndex, Blob x) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setClob(int parameterIndex, Clob x) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setArray(int parameterIndex, Array x) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public ResultSetMetaData getMetaData() throws SQLException {
        return null;
    }
    
    @Override
    public void setDate(int parameterIndex, Date x, Calendar cal) throws SQLException {
        setDate(parameterIndex, x);
    }
    
    @Override
    public void setTime(int parameterIndex, Time x, Calendar cal) throws SQLException {
        setTime(parameterIndex, x);
    }
    
    @Override
    public void setTimestamp(int parameterIndex, Timestamp x, Calendar cal) throws SQLException {
        setTimestamp(parameterIndex, x);
    }
    
    @Override
    public void setNull(int parameterIndex, int sqlType, String typeName) throws SQLException {
        setNull(parameterIndex, sqlType);
    }
    
    @Override
    public void setURL(int parameterIndex, URL x) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public ParameterMetaData getParameterMetaData() throws SQLException {
        return null;
    }
    
    @Override
    public void setRowId(int parameterIndex, RowId x) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setNString(int parameterIndex, String value) throws SQLException {
        setString(parameterIndex, value);
    }
    
    @Override
    public void setNCharacterStream(int parameterIndex, Reader value, long length) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setNClob(int parameterIndex, NClob value) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setClob(int parameterIndex, Reader reader, long length) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setBlob(int parameterIndex, InputStream inputStream, long length) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setNClob(int parameterIndex, Reader reader, long length) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setSQLXML(int parameterIndex, SQLXML xmlObject) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setObject(int parameterIndex, Object x, int targetSqlType, int scaleOrLength) throws SQLException {
        setObject(parameterIndex, x, targetSqlType);
    }
    
    @Override
    public void setAsciiStream(int parameterIndex, InputStream x, long length) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setBinaryStream(int parameterIndex, InputStream x, long length) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setCharacterStream(int parameterIndex, Reader reader, long length) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setAsciiStream(int parameterIndex, InputStream x) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setBinaryStream(int parameterIndex, InputStream x) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setCharacterStream(int parameterIndex, Reader reader) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setNCharacterStream(int parameterIndex, Reader value) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setClob(int parameterIndex, Reader reader) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setBlob(int parameterIndex, InputStream inputStream) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
    
    @Override
    public void setNClob(int parameterIndex, Reader reader) throws SQLException {
        throw new SQLFeatureNotSupportedException();
    }
}
