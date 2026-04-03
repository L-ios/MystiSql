package io.github.mystisql.jdbc.client;

import io.github.mystisql.jdbc.MystiSqlResultSet;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Disabled;
import static org.junit.jupiter.api.Assertions.*;

import java.sql.SQLException;

/**
 * Contract tests for Transport interface implementations.
 * All Transport implementations should pass these tests.
 */
interface TransportContractTest {

    Transport createTransport(String baseUrl, String token, int timeoutSeconds) throws Exception;
    void shutdownTransport(Transport transport) throws Exception;

    @Test
    @DisplayName("getTransportType should return non-null string")
    default void testGetTransportType() throws Exception {
        Transport transport = createTransport("http://localhost:8080", null, 30);
        try {
            String type = transport.getTransportType();
            assertNotNull(type);
            assertFalse(type.isEmpty());
        } finally {
            shutdownTransport(transport);
        }
    }

    @Test
    @DisplayName("close should not throw exception")
    default void testCloseDoesNotThrow() throws Exception {
        Transport transport = createTransport("http://localhost:8080", null, 30);
        assertDoesNotThrow(() -> shutdownTransport(transport));
    }

    @Test
    @Disabled("Requires running WebSocket server")
    @DisplayName("executeQuery should return non-null ResultSet on success")
    default void testExecuteQueryReturnsResultSet() throws Exception {
        Transport transport = createTransport("http://localhost:8080", null, 30);
        try {
            QueryRequest request = new QueryRequest();
            request.setInstance("test-instance");
            request.setQuery("SELECT 1 AS id, 'test' AS name");

            MystiSqlResultSet result = transport.executeQuery(request);
            assertNotNull(result);
        } finally {
            shutdownTransport(transport);
        }
    }

    @Test
    @Disabled("Requires running WebSocket server")
    @DisplayName("executeUpdate should return non-null ExecResult on success")
    default void testExecuteUpdateReturnsExecResult() throws Exception {
        Transport transport = createTransport("http://localhost:8080", null, 30);
        try {
            QueryRequest request = new QueryRequest();
            request.setInstance("test-instance");
            request.setQuery("UPDATE test SET value = 1 WHERE id = 1");

            ExecResult result = transport.executeUpdate(request);
            assertNotNull(result);
        } finally {
            shutdownTransport(transport);
        }
    }

    @Test
    @Disabled("Requires running WebSocket server")
    @DisplayName("healthCheck should return boolean")
    default void testHealthCheckReturnsBoolean() throws Exception {
        Transport transport = createTransport("http://localhost:8080", null, 30);
        try {
            boolean healthy = transport.healthCheck("test-instance");
            assertDoesNotThrow(() -> transport.healthCheck("test-instance"));
        } finally {
            shutdownTransport(transport);
        }
    }

    @Test
    @DisplayName("Transport should implement AutoCloseable")
    default void testImplementsAutoCloseable() throws Exception {
        Transport transport = createTransport("http://localhost:8080", null, 30);
        assertTrue(transport instanceof AutoCloseable);
        shutdownTransport(transport);
    }
}
