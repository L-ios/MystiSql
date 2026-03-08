package io.github.mystisql.jdbc;

import java.sql.*;
import java.util.ArrayList;
import java.util.List;

/**
 * MystiSql JDBC PreparedStatement implementation.
 */
public class MystiSqlPreparedStatement extends MystiSqlStatement implements PreparedStatement {
    
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
        // TODO: Implement HTTP call with parameters
        throw new SQLException("Not implemented yet");
    }
    
    @Override
    public int executeUpdate() throws SQLException {
        checkClosed();
        // TODO: Implement HTTP call with parameters
        throw new SQLException("Not implemented yet");
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
    public void setBigDecimal(int parameterIndex, java.math.BigDecimal x) throws SQLException {
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
    public void setDate(int parameterIndex, java.sql.Date x) throws SQLException {
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
    
    @Override
    public ResultSet getGeneratedKeys() throws SQLException {
        // TODO: Implement
        throw new SQLException("Not implemented yet");
    }
    
    private void setParameter(int index, Parameter parameter) throws SQLException {
        if (index < 1) {
            throw new SQLException("Parameter index must be >= 1");
        }
        
        // Expand list if needed
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
    
    /**
     * Parameter holder class.
     */
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
    
    // Placeholder implementations for remaining methods
    @Override public void setAsciiStream(int parameterIndex, java.io.InputStream x, int length) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setUnicodeStream(int parameterIndex, java.io.InputStream x, int length) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setBinaryStream(int parameterIndex, java.io.InputStream x, int length) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setObject(int parameterIndex, Object x, int targetSqlType) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setObject(int parameterIndex, Object x) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public boolean execute() throws SQLException { throw new SQLException("Not implemented yet"); }
    @Override public void addBatch() throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setCharacterStream(int parameterIndex, java.io.Reader reader, int length) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setRef(int parameterIndex, Ref x) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setBlob(int parameterIndex, Blob x) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setClob(int parameterIndex, Clob x) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setArray(int parameterIndex, Array x) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public ResultSetMetaData getMetaData() throws SQLException { return null; }
    @Override public void setDate(int parameterIndex, java.sql.Date x, Calendar cal) throws SQLException { setDate(parameterIndex, x); }
    @Override public void setTime(int parameterIndex, Time x, Calendar cal) throws SQLException { setTime(parameterIndex, x); }
    @Override public void setTimestamp(int parameterIndex, Timestamp x, Calendar cal) throws SQLException { setTimestamp(parameterIndex, x); }
    @Override public void setNull(int parameterIndex, int sqlType, String typeName) throws SQLException { setNull(parameterIndex, sqlType); }
    @Override public void setURL(int parameterIndex, java.net.URL x) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public ParameterMetaData getParameterMetaData() throws SQLException { return null; }
    @Override public void setRowId(int parameterIndex, RowId x) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setNString(int parameterIndex, String value) throws SQLException { setString(parameterIndex, value); }
    @Override public void setNCharacterStream(int parameterIndex, java.io.Reader value, long length) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setNClob(int parameterIndex, NClob value) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setClob(int parameterIndex, java.io.Reader reader, long length) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setBlob(int parameterIndex, java.io.InputStream inputStream, long length) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setNClob(int parameterIndex, java.io.Reader reader, long length) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setSQLXML(int parameterIndex, SQLXML xmlObject) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setObject(int parameterIndex, Object x, int targetSqlType, int scaleOrLength) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setAsciiStream(int parameterIndex, java.io.InputStream x, long length) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setBinaryStream(int parameterIndex, java.io.InputStream x, long length) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setCharacterStream(int parameterIndex, java.io.Reader reader, long length) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setAsciiStream(int parameterIndex, java.io.InputStream x) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setBinaryStream(int parameterIndex, java.io.InputStream x) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setCharacterStream(int parameterIndex, java.io.Reader reader) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setNCharacterStream(int parameterIndex, java.io.Reader value) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setClob(int parameterIndex, java.io.Reader reader) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setBlob(int parameterIndex, java.io.InputStream inputStream) throws SQLException { throw new SQLFeatureNotSupportedException(); }
    @Override public void setNClob(int parameterIndex, java.io.Reader reader) throws SQLException { throw new SQLFeatureNotSupportedException(); }
}
