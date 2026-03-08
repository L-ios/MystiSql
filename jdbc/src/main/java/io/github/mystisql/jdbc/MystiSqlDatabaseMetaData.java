package io.github.mystisql.jdbc;

import io.github.mystisql.jdbc.client.RestClient;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.sql.*;
import java.util.ArrayList;
import java.util.List;

public class MystiSqlDatabaseMetaData implements DatabaseMetaData {
    
    private static final Logger logger = LoggerFactory.getLogger(MystiSqlDatabaseMetaData.class);
    
    private final MystiSqlConnection connection;
    
    public MystiSqlDatabaseMetaData(MystiSqlConnection connection) {
        this.connection = connection;
    }
    
    @Override
    public String getDatabaseProductName() { 
        return "MystiSql Gateway"; 
    }
    
    @Override
    public String getDatabaseProductVersion() { 
        return "1.0.0"; 
    }
    
    @Override
    public String getDriverName() { 
        return "MystiSql JDBC Driver"; 
    }
    
    @Override
    public String getDriverVersion() { 
        return "1.0.0"; 
    }
    
    @Override
    public int getDriverMajorVersion() { 
        return 1; 
    }
    
    @Override
    public int getDriverMinorVersion() { 
        return 0; 
    }
    
    @Override
    public String getUserName() { 
        return connection.getUsername(); 
    }
    
    @Override
    public String getURL() { 
        return String.format("%s://%s:%d/%s", 
            connection.isSsl() ? "https" : "http", 
            connection.getHost(), 
            connection.getPort(), 
            connection.getInstanceName()); 
    }
    
    @Override
    public Connection getConnection() { 
        return connection; 
    }
    
    @Override
    public ResultSet getCatalogs() throws SQLException {
        logger.debug("getCatalogs()");
        
        MystiSqlResultSet.Column[] columns = {
            new MystiSqlResultSet.Column("TABLE_CAT", Types.VARCHAR, "VARCHAR")
        };
        
        try {
            RestClient client = connection.getRestClient();
            MystiSqlResultSet rs = client.executeQuery(
                connection.getInstanceName(), 
                "SELECT SCHEMA_NAME AS TABLE_CAT FROM INFORMATION_SCHEMA.SCHEMATA"
            );
            return rs;
        } catch (SQLException e) {
            logger.warn("Failed to query catalogs, returning empty result", e);
            return new MystiSqlResultSet(columns, new Object[0][]);
        }
    }
    
    @Override
    public ResultSet getSchemas() throws SQLException {
        return getSchemas(null, null);
    }
    
    @Override
    public ResultSet getSchemas(String catalog, String schemaPattern) throws SQLException {
        logger.debug("getSchemas(catalog={}, schemaPattern={})", catalog, schemaPattern);
        
        MystiSqlResultSet.Column[] columns = {
            new MystiSqlResultSet.Column("TABLE_SCHEM", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TABLE_CATALOG", Types.VARCHAR, "VARCHAR")
        };
        
        try {
            StringBuilder sql = new StringBuilder(
                "SELECT SCHEMA_NAME AS TABLE_SCHEM, SCHEMA_NAME AS TABLE_CATALOG " +
                "FROM INFORMATION_SCHEMA.SCHEMATA WHERE 1=1"
            );
            
            if (schemaPattern != null && !schemaPattern.isEmpty()) {
                sql.append(" AND SCHEMA_NAME LIKE '").append(escapePattern(schemaPattern)).append("'");
            }
            
            RestClient client = connection.getRestClient();
            return client.executeQuery(connection.getInstanceName(), sql.toString());
        } catch (SQLException e) {
            logger.warn("Failed to query schemas, returning empty result", e);
            return new MystiSqlResultSet(columns, new Object[0][]);
        }
    }
    
    @Override
    public ResultSet getTables(String catalog, String schemaPattern, String tableNamePattern, String[] types) 
            throws SQLException {
        logger.debug("getTables(catalog={}, schema={}, table={}, types={})", 
            catalog, schemaPattern, tableNamePattern, types);
        
        MystiSqlResultSet.Column[] columns = {
            new MystiSqlResultSet.Column("TABLE_CAT", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TABLE_SCHEM", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TABLE_NAME", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TABLE_TYPE", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("REMARKS", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TYPE_CAT", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TYPE_SCHEM", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TYPE_NAME", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("SELF_REFERENCING_COL_NAME", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("REF_GENERATION", Types.VARCHAR, "VARCHAR")
        };
        
        try {
            StringBuilder sql = new StringBuilder(
                "SELECT " +
                "  TABLE_SCHEMA AS TABLE_CAT, " +
                "  TABLE_SCHEMA AS TABLE_SCHEM, " +
                "  TABLE_NAME, " +
                "  TABLE_TYPE, " +
                "  TABLE_COMMENT AS REMARKS, " +
                "  NULL AS TYPE_CAT, " +
                "  NULL AS TYPE_SCHEM, " +
                "  NULL AS TYPE_NAME, " +
                "  NULL AS SELF_REFERENCING_COL_NAME, " +
                "  NULL AS REF_GENERATION " +
                "FROM INFORMATION_SCHEMA.TABLES WHERE 1=1"
            );
            
            if (schemaPattern != null && !schemaPattern.isEmpty()) {
                sql.append(" AND TABLE_SCHEMA LIKE '").append(escapePattern(schemaPattern)).append("'");
            }
            
            if (tableNamePattern != null && !tableNamePattern.isEmpty()) {
                sql.append(" AND TABLE_NAME LIKE '").append(escapePattern(tableNamePattern)).append("'");
            }
            
            if (types != null && types.length > 0) {
                sql.append(" AND TABLE_TYPE IN (");
                for (int i = 0; i < types.length; i++) {
                    if (i > 0) sql.append(", ");
                    sql.append("'").append(types[i]).append("'");
                }
                sql.append(")");
            }
            
            sql.append(" ORDER BY TABLE_TYPE, TABLE_SCHEMA, TABLE_NAME");
            
            RestClient client = connection.getRestClient();
            return client.executeQuery(connection.getInstanceName(), sql.toString());
        } catch (SQLException e) {
            logger.warn("Failed to query tables, returning empty result", e);
            return new MystiSqlResultSet(columns, new Object[0][]);
        }
    }
    
    @Override
    public ResultSet getColumns(String catalog, String schemaPattern, String tableNamePattern, String columnNamePattern) 
            throws SQLException {
        logger.debug("getColumns(catalog={}, schema={}, table={}, column={})", 
            catalog, schemaPattern, tableNamePattern, columnNamePattern);
        
        MystiSqlResultSet.Column[] columns = {
            new MystiSqlResultSet.Column("TABLE_CAT", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TABLE_SCHEM", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TABLE_NAME", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("COLUMN_NAME", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("DATA_TYPE", Types.INTEGER, "INT"),
            new MystiSqlResultSet.Column("TYPE_NAME", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("COLUMN_SIZE", Types.INTEGER, "INT"),
            new MystiSqlResultSet.Column("BUFFER_LENGTH", Types.INTEGER, "INT"),
            new MystiSqlResultSet.Column("DECIMAL_DIGITS", Types.INTEGER, "INT"),
            new MystiSqlResultSet.Column("NUM_PREC_RADIX", Types.INTEGER, "INT"),
            new MystiSqlResultSet.Column("NULLABLE", Types.INTEGER, "INT"),
            new MystiSqlResultSet.Column("REMARKS", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("COLUMN_DEF", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("SQL_DATA_TYPE", Types.INTEGER, "INT"),
            new MystiSqlResultSet.Column("SQL_DATETIME_SUB", Types.INTEGER, "INT"),
            new MystiSqlResultSet.Column("CHAR_OCTET_LENGTH", Types.INTEGER, "INT"),
            new MystiSqlResultSet.Column("ORDINAL_POSITION", Types.INTEGER, "INT"),
            new MystiSqlResultSet.Column("IS_NULLABLE", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("SCOPE_CATALOG", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("SCOPE_SCHEMA", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("SCOPE_TABLE", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("SOURCE_DATA_TYPE", Types.SMALLINT, "SMALLINT"),
            new MystiSqlResultSet.Column("IS_AUTOINCREMENT", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("IS_GENERATEDCOLUMN", Types.VARCHAR, "VARCHAR")
        };
        
        try {
            StringBuilder sql = new StringBuilder(
                "SELECT " +
                "  TABLE_SCHEMA AS TABLE_CAT, " +
                "  TABLE_SCHEMA AS TABLE_SCHEM, " +
                "  TABLE_NAME, " +
                "  COLUMN_NAME, " +
                "  CASE UPPER(DATA_TYPE) " +
                "    WHEN 'TINYINT' THEN " + Types.TINYINT + " " +
                "    WHEN 'SMALLINT' THEN " + Types.SMALLINT + " " +
                "    WHEN 'INT' THEN " + Types.INTEGER + " " +
                "    WHEN 'INTEGER' THEN " + Types.INTEGER + " " +
                "    WHEN 'BIGINT' THEN " + Types.BIGINT + " " +
                "    WHEN 'FLOAT' THEN " + Types.FLOAT + " " +
                "    WHEN 'DOUBLE' THEN " + Types.DOUBLE + " " +
                "    WHEN 'DECIMAL' THEN " + Types.DECIMAL + " " +
                "    WHEN 'NUMERIC' THEN " + Types.NUMERIC + " " +
                "    WHEN 'CHAR' THEN " + Types.CHAR + " " +
                "    WHEN 'VARCHAR' THEN " + Types.VARCHAR + " " +
                "    WHEN 'TEXT' THEN " + Types.LONGVARCHAR + " " +
                "    WHEN 'LONGTEXT' THEN " + Types.LONGVARCHAR + " " +
                "    WHEN 'DATE' THEN " + Types.DATE + " " +
                "    WHEN 'TIME' THEN " + Types.TIME + " " +
                "    WHEN 'DATETIME' THEN " + Types.TIMESTAMP + " " +
                "    WHEN 'TIMESTAMP' THEN " + Types.TIMESTAMP + " " +
                "    WHEN 'BLOB' THEN " + Types.BLOB + " " +
                "    WHEN 'BINARY' THEN " + Types.BINARY + " " +
                "    WHEN 'BOOLEAN' THEN " + Types.BOOLEAN + " " +
                "    ELSE " + Types.OTHER + " " +
                "  END AS DATA_TYPE, " +
                "  DATA_TYPE AS TYPE_NAME, " +
                "  CHARACTER_MAXIMUM_LENGTH AS COLUMN_SIZE, " +
                "  NULL AS BUFFER_LENGTH, " +
                "  NUMERIC_SCALE AS DECIMAL_DIGITS, " +
                "  10 AS NUM_PREC_RADIX, " +
                "  CASE WHEN IS_NULLABLE = 'YES' THEN " + DatabaseMetaData.columnNullable + 
                       " ELSE " + DatabaseMetaData.columnNoNulls + " END AS NULLABLE, " +
                "  COLUMN_COMMENT AS REMARKS, " +
                "  COLUMN_DEFAULT AS COLUMN_DEF, " +
                "  NULL AS SQL_DATA_TYPE, " +
                "  NULL AS SQL_DATETIME_SUB, " +
                "  CHARACTER_MAXIMUM_LENGTH AS CHAR_OCTET_LENGTH, " +
                "  ORDINAL_POSITION, " +
                "  IS_NULLABLE, " +
                "  NULL AS SCOPE_CATALOG, " +
                "  NULL AS SCOPE_SCHEMA, " +
                "  NULL AS SCOPE_TABLE, " +
                "  NULL AS SOURCE_DATA_TYPE, " +
                "  CASE WHEN EXTRA LIKE '%auto_increment%' THEN 'YES' ELSE 'NO' END AS IS_AUTOINCREMENT, " +
                "  'NO' AS IS_GENERATEDCOLUMN " +
                "FROM INFORMATION_SCHEMA.COLUMNS WHERE 1=1"
            );
            
            if (schemaPattern != null && !schemaPattern.isEmpty()) {
                sql.append(" AND TABLE_SCHEMA LIKE '").append(escapePattern(schemaPattern)).append("'");
            }
            
            if (tableNamePattern != null && !tableNamePattern.isEmpty()) {
                sql.append(" AND TABLE_NAME LIKE '").append(escapePattern(tableNamePattern)).append("'");
            }
            
            if (columnNamePattern != null && !columnNamePattern.isEmpty()) {
                sql.append(" AND COLUMN_NAME LIKE '").append(escapePattern(columnNamePattern)).append("'");
            }
            
            sql.append(" ORDER BY TABLE_SCHEMA, TABLE_NAME, ORDINAL_POSITION");
            
            RestClient client = connection.getRestClient();
            return client.executeQuery(connection.getInstanceName(), sql.toString());
        } catch (SQLException e) {
            logger.warn("Failed to query columns, returning empty result", e);
            return new MystiSqlResultSet(columns, new Object[0][]);
        }
    }
    
    @Override
    public ResultSet getPrimaryKeys(String catalog, String schema, String table) throws SQLException {
        logger.debug("getPrimaryKeys(catalog={}, schema={}, table={})", catalog, schema, table);
        
        MystiSqlResultSet.Column[] columns = {
            new MystiSqlResultSet.Column("TABLE_CAT", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TABLE_SCHEM", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TABLE_NAME", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("COLUMN_NAME", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("KEY_SEQ", Types.SMALLINT, "SMALLINT"),
            new MystiSqlResultSet.Column("PK_NAME", Types.VARCHAR, "VARCHAR")
        };
        
        try {
            StringBuilder sql = new StringBuilder(
                "SELECT " +
                "  TABLE_SCHEMA AS TABLE_CAT, " +
                "  TABLE_SCHEMA AS TABLE_SCHEM, " +
                "  TABLE_NAME, " +
                "  COLUMN_NAME, " +
                "  ORDINAL_POSITION AS KEY_SEQ, " +
                "  CONSTRAINT_NAME AS PK_NAME " +
                "FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE " +
                "WHERE CONSTRAINT_NAME = 'PRIMARY'"
            );
            
            if (schema != null && !schema.isEmpty()) {
                sql.append(" AND TABLE_SCHEMA LIKE '").append(escapePattern(schema)).append("'");
            }
            
            if (table != null && !table.isEmpty()) {
                sql.append(" AND TABLE_NAME LIKE '").append(escapePattern(table)).append("'");
            }
            
            sql.append(" ORDER BY TABLE_SCHEMA, TABLE_NAME, ORDINAL_POSITION");
            
            RestClient client = connection.getRestClient();
            return client.executeQuery(connection.getInstanceName(), sql.toString());
        } catch (SQLException e) {
            logger.warn("Failed to query primary keys, returning empty result", e);
            return new MystiSqlResultSet(columns, new Object[0][]);
        }
    }
    
    @Override
    public ResultSet getIndexInfo(String catalog, String schema, String table, boolean unique, boolean approximate) 
            throws SQLException {
        logger.debug("getIndexInfo(catalog={}, schema={}, table={}, unique={})", 
            catalog, schema, table, unique);
        
        MystiSqlResultSet.Column[] columns = {
            new MystiSqlResultSet.Column("TABLE_CAT", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TABLE_SCHEM", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TABLE_NAME", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("NON_UNIQUE", Types.BOOLEAN, "BOOLEAN"),
            new MystiSqlResultSet.Column("INDEX_QUALIFIER", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("INDEX_NAME", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("TYPE", Types.SMALLINT, "SMALLINT"),
            new MystiSqlResultSet.Column("ORDINAL_POSITION", Types.SMALLINT, "SMALLINT"),
            new MystiSqlResultSet.Column("COLUMN_NAME", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("ASC_OR_DESC", Types.VARCHAR, "VARCHAR"),
            new MystiSqlResultSet.Column("CARDINALITY", Types.BIGINT, "BIGINT"),
            new MystiSqlResultSet.Column("PAGES", Types.BIGINT, "BIGINT"),
            new MystiSqlResultSet.Column("FILTER_CONDITION", Types.VARCHAR, "VARCHAR")
        };
        
        try {
            StringBuilder sql = new StringBuilder(
                "SELECT " +
                "  TABLE_SCHEMA AS TABLE_CAT, " +
                "  TABLE_SCHEMA AS TABLE_SCHEM, " +
                "  TABLE_NAME, " +
                "  NOT NON_UNIQUE AS NON_UNIQUE, " +
                "  TABLE_SCHEMA AS INDEX_QUALIFIER, " +
                "  INDEX_NAME, " +
                "  " + DatabaseMetaData.tableIndexOther + " AS TYPE, " +
                "  SEQ_IN_INDEX AS ORDINAL_POSITION, " +
                "  COLUMN_NAME, " +
                "  CASE WHEN COLLATION = 'A' THEN 'A' ELSE 'D' END AS ASC_OR_DESC, " +
                "  CARDINALITY, " +
                "  NULL AS PAGES, " +
                "  NULL AS FILTER_CONDITION " +
                "FROM INFORMATION_SCHEMA.STATISTICS WHERE 1=1"
            );
            
            if (schema != null && !schema.isEmpty()) {
                sql.append(" AND TABLE_SCHEMA LIKE '").append(escapePattern(schema)).append("'");
            }
            
            if (table != null && !table.isEmpty()) {
                sql.append(" AND TABLE_NAME LIKE '").append(escapePattern(table)).append("'");
            }
            
            if (unique) {
                sql.append(" AND NON_UNIQUE = 0");
            }
            
            sql.append(" ORDER BY TABLE_SCHEMA, TABLE_NAME, NON_UNIQUE, INDEX_NAME, SEQ_IN_INDEX");
            
            RestClient client = connection.getRestClient();
            return client.executeQuery(connection.getInstanceName(), sql.toString());
        } catch (SQLException e) {
            logger.warn("Failed to query index info, returning empty result", e);
            return new MystiSqlResultSet(columns, new Object[0][]);
        }
    }
    
    private String escapePattern(String pattern) {
        if (pattern == null) return null;
        return pattern.replace("'", "''");
    }
    
    @Override public boolean allProceduresAreCallable() { return false; }
    @Override public boolean allTablesAreSelectable() { return true; }
    @Override public boolean isReadOnly() { return false; }
    @Override public boolean nullsAreSortedHigh() { return false; }
    @Override public boolean nullsAreSortedLow() { return true; }
    @Override public boolean nullsAreSortedAtStart() { return false; }
    @Override public boolean nullsAreSortedAtEnd() { return true; }
    @Override public String getIdentifierQuoteString() { return "`"; }
    @Override public String getSQLKeywords() { return ""; }
    @Override public String getNumericFunctions() { return ""; }
    @Override public String getStringFunctions() { return ""; }
    @Override public String getSystemFunctions() { return ""; }
    @Override public String getTimeDateFunctions() { return ""; }
    @Override public String getSearchStringEscape() { return "\\"; }
    @Override public String getExtraNameCharacters() { return ""; }
    @Override public boolean supportsAlterTableWithAddColumn() { return true; }
    @Override public boolean supportsAlterTableWithDropColumn() { return true; }
    @Override public boolean supportsColumnAliasing() { return true; }
    @Override public boolean nullPlusNonNullIsNull() { return true; }
    @Override public boolean supportsConvert() { return false; }
    @Override public boolean supportsConvert(int fromType, int toType) { return false; }
    @Override public boolean supportsTableCorrelationNames() { return true; }
    @Override public boolean supportsDifferentTableCorrelationNames() { return false; }
    @Override public boolean supportsExpressionsInOrderBy() { return true; }
    @Override public boolean supportsOrderByUnrelated() { return true; }
    @Override public boolean supportsGroupBy() { return true; }
    @Override public boolean supportsGroupByUnrelated() { return true; }
    @Override public boolean supportsGroupByBeyondSelect() { return true; }
    @Override public boolean supportsLikeEscapeClause() { return true; }
    @Override public boolean supportsMultipleResultSets() { return false; }
    @Override public boolean supportsMultipleTransactions() { return false; }
    @Override public boolean supportsNonNullableColumns() { return true; }
    @Override public boolean supportsMinimumSQLGrammar() { return true; }
    @Override public boolean supportsCoreSQLGrammar() { return true; }
    @Override public boolean supportsExtendedSQLGrammar() { return true; }
    @Override public boolean supportsANSI92EntryLevelSQL() { return true; }
    @Override public boolean supportsANSI92IntermediateSQL() { return false; }
    @Override public boolean supportsANSI92FullSQL() { return false; }
    @Override public boolean supportsIntegrityEnhancementFacility() { return false; }
    @Override public boolean supportsOuterJoins() { return true; }
    @Override public boolean supportsFullOuterJoins() { return true; }
    @Override public boolean supportsLimitedOuterJoins() { return true; }
    @Override public String getSchemaTerm() { return "schema"; }
    @Override public String getProcedureTerm() { return "procedure"; }
    @Override public String getCatalogTerm() { return "catalog"; }
    @Override public boolean isCatalogAtStart() { return true; }
    @Override public String getCatalogSeparator() { return "."; }
    @Override public boolean supportsSchemasInDataManipulation() { return true; }
    @Override public boolean supportsSchemasInProcedureCalls() { return false; }
    @Override public boolean supportsSchemasInTableDefinitions() { return true; }
    @Override public boolean supportsSchemasInIndexDefinitions() { return true; }
    @Override public boolean supportsSchemasInPrivilegeDefinitions() { return true; }
    @Override public boolean supportsCatalogsInDataManipulation() { return true; }
    @Override public boolean supportsCatalogsInProcedureCalls() { return false; }
    @Override public boolean supportsCatalogsInTableDefinitions() { return true; }
    @Override public boolean supportsCatalogsInIndexDefinitions() { return true; }
    @Override public boolean supportsCatalogsInPrivilegeDefinitions() { return true; }
    @Override public boolean supportsPositionedDelete() { return false; }
    @Override public boolean supportsPositionedUpdate() { return false; }
    @Override public boolean supportsSelectForUpdate() { return false; }
    @Override public boolean supportsStoredProcedures() { return false; }
    @Override public boolean supportsSubqueriesInComparisons() { return true; }
    @Override public boolean supportsSubqueriesInExists() { return true; }
    @Override public boolean supportsSubqueriesInIns() { return true; }
    @Override public boolean supportsSubqueriesInQuantifieds() { return true; }
    @Override public boolean supportsCorrelatedSubqueries() { return true; }
    @Override public boolean supportsUnion() { return true; }
    @Override public boolean supportsUnionAll() { return true; }
    @Override public boolean supportsOpenCursorsAcrossCommit() { return false; }
    @Override public boolean supportsOpenCursorsAcrossRollback() { return false; }
    @Override public boolean supportsOpenStatementsAcrossCommit() { return false; }
    @Override public boolean supportsOpenStatementsAcrossRollback() { return false; }
    @Override public int getMaxBinaryLiteralLength() { return 0; }
    @Override public int getMaxCharLiteralLength() { return 0; }
    @Override public int getMaxColumnNameLength() { return 64; }
    @Override public int getMaxColumnsInGroupBy() { return 0; }
    @Override public int getMaxColumnsInIndex() { return 16; }
    @Override public int getMaxColumnsInOrderBy() { return 0; }
    @Override public int getMaxColumnsInSelect() { return 0; }
    @Override public int getMaxColumnsInTable() { return 0; }
    @Override public int getMaxConnections() { return 0; }
    @Override public int getMaxCursorNameLength() { return 64; }
    @Override public int getMaxIndexLength() { return 0; }
    @Override public int getMaxSchemaNameLength() { return 64; }
    @Override public int getMaxProcedureNameLength() { return 64; }
    @Override public int getMaxCatalogNameLength() { return 64; }
    @Override public int getMaxRowSize() { return 0; }
    @Override public boolean doesMaxRowSizeIncludeBlobs() { return true; }
    @Override public int getMaxStatementLength() { return 0; }
    @Override public int getMaxStatements() { return 0; }
    @Override public int getMaxTableNameLength() { return 64; }
    @Override public int getMaxTablesInSelect() { return 0; }
    @Override public int getMaxUserNameLength() { return 64; }
    @Override public int getDefaultTransactionIsolation() { return Connection.TRANSACTION_READ_COMMITTED; }
    @Override public boolean supportsTransactions() { return false; }
    @Override public boolean supportsTransactionIsolationLevel(int level) { return false; }
    @Override public boolean supportsDataDefinitionAndDataManipulationTransactions() { return false; }
    @Override public boolean supportsDataManipulationTransactionsOnly() { return false; }
    @Override public boolean dataDefinitionCausesTransactionCommit() { return false; }
    @Override public boolean dataDefinitionIgnoredInTransactions() { return false; }
    @Override public ResultSet getProcedures(String catalog, String schemaPattern, String procedureNamePattern) { return null; }
    @Override public ResultSet getProcedureColumns(String catalog, String schemaPattern, String procedureNamePattern, String columnNamePattern) { return null; }
    @Override public ResultSet getColumnPrivileges(String catalog, String schema, String table, String columnNamePattern) { return null; }
    @Override public ResultSet getTablePrivileges(String catalog, String schemaPattern, String tableNamePattern) { return null; }
    @Override public ResultSet getBestRowIdentifier(String catalog, String schema, String table, int scope, boolean nullable) { return null; }
    @Override public ResultSet getVersionColumns(String catalog, String schema, String table) { return null; }
    @Override public ResultSet getImportedKeys(String catalog, String schema, String table) { return null; }
    @Override public ResultSet getExportedKeys(String catalog, String schema, String table) { return null; }
    @Override public ResultSet getCrossReference(String parentCatalog, String parentSchema, String parentTable, String foreignCatalog, String foreignSchema, String foreignTable) { return null; }
    @Override public ResultSet getTypeInfo() { return null; }
    @Override public boolean supportsResultSetType(int type) { return type == ResultSet.TYPE_FORWARD_ONLY; }
    @Override public boolean supportsResultSetConcurrency(int type, int concurrency) { return type == ResultSet.TYPE_FORWARD_ONLY && concurrency == ResultSet.CONCUR_READ_ONLY; }
    @Override public boolean ownUpdatesAreVisible(int type) { return false; }
    @Override public boolean ownDeletesAreVisible(int type) { return false; }
    @Override public boolean ownInsertsAreVisible(int type) { return false; }
    @Override public boolean othersUpdatesAreVisible(int type) { return false; }
    @Override public boolean othersDeletesAreVisible(int type) { return false; }
    @Override public boolean othersInsertsAreVisible(int type) { return false; }
    @Override public boolean updatesAreDetected(int type) { return false; }
    @Override public boolean deletesAreDetected(int type) { return false; }
    @Override public boolean insertsAreDetected(int type) { return false; }
    @Override public boolean supportsBatchUpdates() { return false; }
    @Override public ResultSet getUDTs(String catalog, String schemaPattern, String typeNamePattern, int[] types) { return null; }
    @Override public boolean supportsSavepoints() { return false; }
    @Override public boolean supportsNamedParameters() { return false; }
    @Override public boolean supportsMultipleOpenResults() { return false; }
    @Override public boolean supportsGetGeneratedKeys() { return false; }
    @Override public ResultSet getSuperTypes(String catalog, String schemaPattern, String typeNamePattern) { return null; }
    @Override public ResultSet getSuperTables(String catalog, String schemaPattern, String tableNamePattern) { return null; }
    @Override public ResultSet getAttributes(String catalog, String schemaPattern, String typeNamePattern, String attributeNamePattern) { return null; }
    @Override public boolean supportsResultSetHoldability(int holdability) { return holdability == ResultSet.HOLD_CURSORS_OVER_COMMIT; }
    @Override public int getResultSetHoldability() { return ResultSet.HOLD_CURSORS_OVER_COMMIT; }
    @Override public int getDatabaseMajorVersion() { return 1; }
    @Override public int getDatabaseMinorVersion() { return 0; }
    @Override public int getJDBCMajorVersion() { return 4; }
    @Override public int getJDBCMinorVersion() { return 2; }
    @Override public int getSQLStateType() { return sqlStateSQL; }
    @Override public boolean locatorsUpdateCopy() { return true; }
    @Override public boolean supportsStatementPooling() { return false; }
    @Override public RowIdLifetime getRowIdLifetime() { return RowIdLifetime.ROWID_UNSUPPORTED; }
    @Override public boolean supportsStoredFunctionsUsingCallSyntax() { return false; }
    @Override public boolean autoCommitFailureClosesAllResultSets() { return false; }
    @Override public ResultSet getClientInfoProperties() { return null; }
    @Override public boolean generatedKeyAlwaysReturned() { return false; }
    @Override public ResultSet getPseudoColumns(String catalog, String schemaPattern, String tableNamePattern, String columnNamePattern) { return null; }
    @Override public <T> T unwrap(Class<T> iface) throws SQLException { 
        if (iface.isAssignableFrom(getClass())) {
            return iface.cast(this);
        }
        throw new SQLFeatureNotSupportedException("Not a wrapper for " + iface.getName()); 
    }
    @Override public boolean isWrapperFor(Class<?> iface) { 
        return iface.isAssignableFrom(getClass()); 
    }
    
    @Override public boolean usesLocalFiles() { return false; }
    @Override public boolean usesLocalFilePerTable() { return false; }
    @Override public boolean supportsMixedCaseIdentifiers() { return false; }
    @Override public boolean storesUpperCaseIdentifiers() { return false; }
    @Override public boolean storesLowerCaseIdentifiers() { return true; }
    @Override public boolean storesMixedCaseIdentifiers() { return false; }
    @Override public boolean supportsMixedCaseQuotedIdentifiers() { return true; }
    @Override public boolean storesUpperCaseQuotedIdentifiers() { return false; }
    @Override public boolean storesLowerCaseQuotedIdentifiers() { return false; }
    @Override public boolean storesMixedCaseQuotedIdentifiers() { return true; }
    @Override public ResultSet getTableTypes() throws SQLException {
        MystiSqlResultSet.Column[] columns = {
            new MystiSqlResultSet.Column("TABLE_TYPE", Types.VARCHAR, "VARCHAR")
        };
        Object[][] rows = {{"TABLE"}, {"VIEW"}, {"SYSTEM TABLE"}};
        return new MystiSqlResultSet(columns, rows);
    }
    @Override public ResultSet getFunctions(String catalog, String schemaPattern, String functionNamePattern) { return null; }
    @Override public ResultSet getFunctionColumns(String catalog, String schemaPattern, String functionNamePattern, String columnNamePattern) { return null; }
}
