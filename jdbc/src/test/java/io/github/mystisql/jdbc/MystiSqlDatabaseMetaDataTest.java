package io.github.mystisql.jdbc;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import static org.junit.jupiter.api.Assertions.*;

import java.sql.*;

/**
 * Unit tests for MystiSqlDatabaseMetaData
 */
class MystiSqlDatabaseMetaDataTest {

    private MystiSqlConnection connection;
    private MystiSqlDatabaseMetaData metaData;

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
        metaData = new MystiSqlDatabaseMetaData(connection);
    }

    @Test
    @DisplayName("GetDatabaseProductName should return MystiSql Gateway")
    void testGetDatabaseProductName() {
        assertEquals("MystiSql Gateway", metaData.getDatabaseProductName());
    }

    @Test
    @DisplayName("GetDatabaseProductVersion should return 1.0.0")
    void testGetDatabaseProductVersion() {
        assertEquals("1.0.0", metaData.getDatabaseProductVersion());
    }

    @Test
    @DisplayName("GetDriverName should return MystiSql JDBC Driver")
    void testGetDriverName() {
        assertEquals("MystiSql JDBC Driver", metaData.getDriverName());
    }

    @Test
    @DisplayName("GetDriverVersion should return 1.0.0")
    void testGetDriverVersion() {
        assertEquals("1.0.0", metaData.getDriverVersion());
    }

    @Test
    @DisplayName("GetDriverMajorVersion should return 1")
    void testGetDriverMajorVersion() {
        assertEquals(1, metaData.getDriverMajorVersion());
    }

    @Test
    @DisplayName("GetDriverMinorVersion should return 0")
    void testGetDriverMinorVersion() {
        assertEquals(0, metaData.getDriverMinorVersion());
    }

    @Test
    @DisplayName("GetDatabaseMajorVersion should return 1")
    void testGetDatabaseMajorVersion() {
        assertEquals(1, metaData.getDatabaseMajorVersion());
    }

    @Test
    @DisplayName("GetDatabaseMinorVersion should return 0")
    void testGetDatabaseMinorVersion() {
        assertEquals(0, metaData.getDatabaseMinorVersion());
    }

    @Test
    @DisplayName("GetJDBCMajorVersion should return 4")
    void testGetJDBCMajorVersion() {
        assertEquals(4, metaData.getJDBCMajorVersion());
    }

    @Test
    @DisplayName("GetJDBCMinorVersion should return 2")
    void testGetJDBCMinorVersion() {
        assertEquals(2, metaData.getJDBCMinorVersion());
    }

    @Test
    @DisplayName("GetUserName should return connection username")
    void testGetUserName() {
        assertEquals("testuser", metaData.getUserName());
    }

    @Test
    @DisplayName("GetURL should return formatted URL")
    void testGetURL() {
        assertEquals("http://localhost:8080/test-instance", metaData.getURL());
    }

    @Test
    @DisplayName("GetURL should use https when SSL is enabled")
    void testGetURLWithSSL() {
        MystiSqlConnection sslConn = new MystiSqlConnection(
            "secure.example.com",
            443,
            "prod-instance",
            "admin",
            "token",
            60,
            true,
            true,
            20
        );
        MystiSqlDatabaseMetaData sslMeta = new MystiSqlDatabaseMetaData(sslConn);
        
        assertEquals("https://secure.example.com:443/prod-instance", sslMeta.getURL());
    }

    @Test
    @DisplayName("GetConnection should return the parent connection")
    void testGetConnection() {
        assertSame(connection, metaData.getConnection());
    }

    @Test
    @DisplayName("GetIdentifierQuoteString should return backtick")
    void testGetIdentifierQuoteString() {
        assertEquals("`", metaData.getIdentifierQuoteString());
    }

    @Test
    @DisplayName("GetSearchStringEscape should return backslash")
    void testGetSearchStringEscape() {
        assertEquals("\\", metaData.getSearchStringEscape());
    }

    @Test
    @DisplayName("GetCatalogSeparator should return dot")
    void testGetCatalogSeparator() {
        assertEquals(".", metaData.getCatalogSeparator());
    }

    @Test
    @DisplayName("GetSchemaTerm should return schema")
    void testGetSchemaTerm() {
        assertEquals("schema", metaData.getSchemaTerm());
    }

    @Test
    @DisplayName("GetProcedureTerm should return procedure")
    void testGetProcedureTerm() {
        assertEquals("procedure", metaData.getProcedureTerm());
    }

    @Test
    @DisplayName("GetCatalogTerm should return catalog")
    void testGetCatalogTerm() {
        assertEquals("catalog", metaData.getCatalogTerm());
    }

    @Test
    @DisplayName("IsCatalogAtStart should return true")
    void testIsCatalogAtStart() {
        assertTrue(metaData.isCatalogAtStart());
    }

    @Test
    @DisplayName("IsReadOnly should return false")
    void testIsReadOnly() {
        assertFalse(metaData.isReadOnly());
    }

    @Test
    @DisplayName("AllTablesAreSelectable should return true")
    void testAllTablesAreSelectable() {
        assertTrue(metaData.allTablesAreSelectable());
    }

    @Test
    @DisplayName("Nulls sorting should be correct")
    void testNullsSorting() {
        assertFalse(metaData.nullsAreSortedHigh());
        assertTrue(metaData.nullsAreSortedLow());
        assertFalse(metaData.nullsAreSortedAtStart());
        assertTrue(metaData.nullsAreSortedAtEnd());
    }

    @Test
    @DisplayName("SupportsAlterTableWithAddColumn should return true")
    void testSupportsAlterTableWithAddColumn() {
        assertTrue(metaData.supportsAlterTableWithAddColumn());
    }

    @Test
    @DisplayName("SupportsAlterTableWithDropColumn should return true")
    void testSupportsAlterTableWithDropColumn() {
        assertTrue(metaData.supportsAlterTableWithDropColumn());
    }

    @Test
    @DisplayName("SupportsColumnAliasing should return true")
    void testSupportsColumnAliasing() {
        assertTrue(metaData.supportsColumnAliasing());
    }

    @Test
    @DisplayName("SupportsGroupBy should return true")
    void testSupportsGroupBy() {
        assertTrue(metaData.supportsGroupBy());
    }

    @Test
    @DisplayName("SupportsUnion should return true")
    void testSupportsUnion() {
        assertTrue(metaData.supportsUnion());
        assertTrue(metaData.supportsUnionAll());
    }

    @Test
    @DisplayName("SupportsTransactions should return false (Phase 2.5)")
    void testSupportsTransactions() {
        assertFalse(metaData.supportsTransactions());
    }

    @Test
    @DisplayName("SupportsStoredProcedures should return false")
    void testSupportsStoredProcedures() {
        assertFalse(metaData.supportsStoredProcedures());
    }

    @Test
    @DisplayName("SupportsBatchUpdates should return false")
    void testSupportsBatchUpdates() {
        assertFalse(metaData.supportsBatchUpdates());
    }

    @Test
    @DisplayName("SupportsSavepoints should return false")
    void testSupportsSavepoints() {
        assertFalse(metaData.supportsSavepoints());
    }

    @Test
    @DisplayName("GetDefaultTransactionIsolation should return READ_COMMITTED")
    void testGetDefaultTransactionIsolation() {
        assertEquals(Connection.TRANSACTION_READ_COMMITTED, metaData.getDefaultTransactionIsolation());
    }

    @Test
    @DisplayName("SupportsResultSetType should return true for TYPE_FORWARD_ONLY")
    void testSupportsResultSetType() {
        assertTrue(metaData.supportsResultSetType(ResultSet.TYPE_FORWARD_ONLY));
        assertFalse(metaData.supportsResultSetType(ResultSet.TYPE_SCROLL_INSENSITIVE));
        assertFalse(metaData.supportsResultSetType(ResultSet.TYPE_SCROLL_SENSITIVE));
    }

    @Test
    @DisplayName("SupportsResultSetConcurrency should return correct values")
    void testSupportsResultSetConcurrency() {
        assertTrue(metaData.supportsResultSetConcurrency(
            ResultSet.TYPE_FORWARD_ONLY, 
            ResultSet.CONCUR_READ_ONLY
        ));
        assertFalse(metaData.supportsResultSetConcurrency(
            ResultSet.TYPE_FORWARD_ONLY, 
            ResultSet.CONCUR_UPDATABLE
        ));
    }

    @Test
    @DisplayName("GetMaxColumnNameLength should return 64")
    void testGetMaxColumnNameLength() {
        assertEquals(64, metaData.getMaxColumnNameLength());
    }

    @Test
    @DisplayName("GetMaxTableNameLength should return 64")
    void testGetMaxTableNameLength() {
        assertEquals(64, metaData.getMaxTableNameLength());
    }

    @Test
    @DisplayName("GetMaxSchemaNameLength should return 64")
    void testGetMaxSchemaNameLength() {
        assertEquals(64, metaData.getMaxSchemaNameLength());
    }

    @Test
    @DisplayName("GetMaxCatalogNameLength should return 64")
    void testGetMaxCatalogNameLength() {
        assertEquals(64, metaData.getMaxCatalogNameLength());
    }

    @Test
    @DisplayName("GetMaxColumnsInIndex should return 16")
    void testGetMaxColumnsInIndex() {
        assertEquals(16, metaData.getMaxColumnsInIndex());
    }

    @Test
    @DisplayName("GetResultSetHoldability should return HOLD_CURSORS_OVER_COMMIT")
    void testGetResultSetHoldability() {
        assertEquals(ResultSet.HOLD_CURSORS_OVER_COMMIT, metaData.getResultSetHoldability());
    }

    @Test
    @DisplayName("SupportsResultSetHoldability should be correct")
    void testSupportsResultSetHoldability() {
        assertTrue(metaData.supportsResultSetHoldability(ResultSet.HOLD_CURSORS_OVER_COMMIT));
        assertFalse(metaData.supportsResultSetHoldability(ResultSet.CLOSE_CURSORS_AT_COMMIT));
    }

    @Test
    @DisplayName("GetSQLStateType should return sqlStateSQL")
    void testGetSQLStateType() {
        assertEquals(DatabaseMetaData.sqlStateSQL, metaData.getSQLStateType());
    }

    @Test
    @DisplayName("LocatorsUpdateCopy should return true")
    void testLocatorsUpdateCopy() {
        assertTrue(metaData.locatorsUpdateCopy());
    }

    @Test
    @DisplayName("GetRowIdLifetime should return ROWID_UNSUPPORTED")
    void testGetRowIdLifetime() {
        assertEquals(RowIdLifetime.ROWID_UNSUPPORTED, metaData.getRowIdLifetime());
    }

    @Test
    @DisplayName("SupportsGetGeneratedKeys should return false")
    void testSupportsGetGeneratedKeys() {
        assertFalse(metaData.supportsGetGeneratedKeys());
    }

    @Test
    @DisplayName("GetCatalogs should throw SQLException")
    void testGetCatalogs() {
        assertThrows(SQLException.class, () -> metaData.getCatalogs());
    }

    @Test
    @DisplayName("GetSchemas should throw SQLException")
    void testGetSchemas() {
        assertThrows(SQLException.class, () -> metaData.getSchemas());
    }

    @Test
    @DisplayName("GetTables should throw SQLException")
    void testGetTables() {
        assertThrows(SQLException.class, 
            () -> metaData.getTables(null, null, "%", null));
    }

    @Test
    @DisplayName("GetColumns should throw SQLException")
    void testGetColumns() {
        assertThrows(SQLException.class, 
            () -> metaData.getColumns(null, null, "users", "%"));
    }

    @Test
    @DisplayName("GetPrimaryKeys should throw SQLException")
    void testGetPrimaryKeys() {
        assertThrows(SQLException.class, 
            () -> metaData.getPrimaryKeys(null, null, "users"));
    }

    @Test
    @DisplayName("GetIndexInfo should throw SQLException")
    void testGetIndexInfo() {
        assertThrows(SQLException.class, 
            () -> metaData.getIndexInfo(null, null, "users", false, true));
    }

    @Test
    @DisplayName("IsWrapperFor should return false")
    void testIsWrapperFor() {
        assertFalse(metaData.isWrapperFor(DatabaseMetaData.class));
    }

    @Test
    @DisplayName("Unwrap should throw SQLFeatureNotSupportedException")
    void testUnwrap() {
        assertThrows(SQLFeatureNotSupportedException.class, 
            () -> metaData.unwrap(DatabaseMetaData.class));
    }

    @Test
    @DisplayName("SupportsSubqueries should return true")
    void testSupportsSubqueries() {
        assertTrue(metaData.supportsSubqueriesInComparisons());
        assertTrue(metaData.supportsSubqueriesInExists());
        assertTrue(metaData.supportsSubqueriesInIns());
        assertTrue(metaData.supportsSubqueriesInQuantifieds());
        assertTrue(metaData.supportsCorrelatedSubqueries());
    }

    @Test
    @DisplayName("SupportsOuterJoins should return true")
    void testSupportsOuterJoins() {
        assertTrue(metaData.supportsOuterJoins());
        assertTrue(metaData.supportsFullOuterJoins());
        assertTrue(metaData.supportsLimitedOuterJoins());
    }

    @Test
    @DisplayName("SupportsSQLGrammar should return correct values")
    void testSupportsSQLGrammar() {
        assertTrue(metaData.supportsMinimumSQLGrammar());
        assertTrue(metaData.supportsCoreSQLGrammar());
        assertTrue(metaData.supportsExtendedSQLGrammar());
        assertTrue(metaData.supportsANSI92EntryLevelSQL());
        assertFalse(metaData.supportsANSI92IntermediateSQL());
        assertFalse(metaData.supportsANSI92FullSQL());
    }

    @Test
    @DisplayName("SupportsSchemas should return true for most operations")
    void testSupportsSchemas() {
        assertTrue(metaData.supportsSchemasInDataManipulation());
        assertFalse(metaData.supportsSchemasInProcedureCalls());
        assertTrue(metaData.supportsSchemasInTableDefinitions());
        assertTrue(metaData.supportsSchemasInIndexDefinitions());
        assertTrue(metaData.supportsSchemasInPrivilegeDefinitions());
    }

    @Test
    @DisplayName("SupportsCatalogs should return true for most operations")
    void testSupportsCatalogs() {
        assertTrue(metaData.supportsCatalogsInDataManipulation());
        assertFalse(metaData.supportsCatalogsInProcedureCalls());
        assertTrue(metaData.supportsCatalogsInTableDefinitions());
        assertTrue(metaData.supportsCatalogsInIndexDefinitions());
        assertTrue(metaData.supportsCatalogsInPrivilegeDefinitions());
    }

    @Test
    @DisplayName("GetProcedures should return null")
    void testGetProcedures() {
        assertNull(metaData.getProcedures(null, null, "%"));
    }

    @Test
    @DisplayName("GetTypeInfo should return null")
    void testGetTypeInfo() {
        assertNull(metaData.getTypeInfo());
    }

    @Test
    @DisplayName("DoesMaxRowSizeIncludeBlobs should return true")
    void testDoesMaxRowSizeIncludeBlobs() {
        assertTrue(metaData.doesMaxRowSizeIncludeBlobs());
    }

    @Test
    @DisplayName("AllProceduresAreCallable should return false")
    void testAllProceduresAreCallable() {
        assertFalse(metaData.allProceduresAreCallable());
    }

    @Test
    @DisplayName("NullPlusNonNullIsNull should return true")
    void testNullPlusNonNullIsNull() {
        assertTrue(metaData.nullPlusNonNullIsNull());
    }

    @Test
    @DisplayName("SupportsLikeEscapeClause should return true")
    void testSupportsLikeEscapeClause() {
        assertTrue(metaData.supportsLikeEscapeClause());
    }

    @Test
    @DisplayName("SupportsExpressionsInOrderBy should return true")
    void testSupportsExpressionsInOrderBy() {
        assertTrue(metaData.supportsExpressionsInOrderBy());
    }

    @Test
    @DisplayName("SupportsOrderByUnrelated should return true")
    void testSupportsOrderByUnrelated() {
        assertTrue(metaData.supportsOrderByUnrelated());
    }

    @Test
    @DisplayName("SupportsConvert should return false")
    void testSupportsConvert() {
        assertFalse(metaData.supportsConvert());
        assertFalse(metaData.supportsConvert(Types.INTEGER, Types.VARCHAR));
    }

    @Test
    @DisplayName("SupportsMultipleResultSets should return false")
    void testSupportsMultipleResultSets() {
        assertFalse(metaData.supportsMultipleResultSets());
    }

    @Test
    @DisplayName("SupportsNonNullableColumns should return true")
    void testSupportsNonNullableColumns() {
        assertTrue(metaData.supportsNonNullableColumns());
    }

    @Test
    @DisplayName("SupportsPositionedDelete and Update should return false")
    void testSupportsPositionedOperations() {
        assertFalse(metaData.supportsPositionedDelete());
        assertFalse(metaData.supportsPositionedUpdate());
        assertFalse(metaData.supportsSelectForUpdate());
    }

    @Test
    @DisplayName("Visibility methods should return false")
    void testVisibilityMethods() {
        assertFalse(metaData.ownUpdatesAreVisible(ResultSet.TYPE_FORWARD_ONLY));
        assertFalse(metaData.ownDeletesAreVisible(ResultSet.TYPE_FORWARD_ONLY));
        assertFalse(metaData.ownInsertsAreVisible(ResultSet.TYPE_FORWARD_ONLY));
        assertFalse(metaData.othersUpdatesAreVisible(ResultSet.TYPE_FORWARD_ONLY));
        assertFalse(metaData.othersDeletesAreVisible(ResultSet.TYPE_FORWARD_ONLY));
        assertFalse(metaData.othersInsertsAreVisible(ResultSet.TYPE_FORWARD_ONLY));
        assertFalse(metaData.updatesAreDetected(ResultSet.TYPE_FORWARD_ONLY));
        assertFalse(metaData.deletesAreDetected(ResultSet.TYPE_FORWARD_ONLY));
        assertFalse(metaData.insertsAreDetected(ResultSet.TYPE_FORWARD_ONLY));
    }

    @Test
    @DisplayName("SupportsStatementPooling should return false")
    void testSupportsStatementPooling() {
        assertFalse(metaData.supportsStatementPooling());
    }

    @Test
    @DisplayName("GeneratedKeyAlwaysReturned should return false")
    void testGeneratedKeyAlwaysReturned() {
        assertFalse(metaData.generatedKeyAlwaysReturned());
    }
}
