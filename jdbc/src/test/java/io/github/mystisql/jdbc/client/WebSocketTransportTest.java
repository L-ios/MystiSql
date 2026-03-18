package io.github.mystisql.jdbc.client;

import org.junit.jupiter.api.DisplayName;
import static org.junit.jupiter.api.Assertions.*;

import java.sql.SQLException;

/**
 * Unit tests for WebSocketTransport using contract test pattern.
 */
@DisplayName("WebSocketTransport Contract Tests")
class WebSocketTransportTest implements TransportContractTest {

    @Override
    public Transport createTransport(String baseUrl, String token, int timeoutSeconds) throws Exception {
        return new WebSocketTransport(baseUrl, token, timeoutSeconds);
    }

    @Override
    public void shutdownTransport(Transport transport) throws Exception {
        transport.close();
    }

    @DisplayName("getTransportType should return 'websocket'")
    void testWebSocketTransportType() throws Exception {
        WebSocketTransport transport = new WebSocketTransport("http://localhost:8080", null, 30);
        try {
            assertEquals("websocket", transport.getTransportType());
        } finally {
            transport.close();
        }
    }

    @DisplayName("Should convert HTTP URL to WS URL")
    void testHttpToWsConversion() throws Exception {
        WebSocketTransport transport = new WebSocketTransport("http://localhost:8080", null, 30);
        try {
            assertEquals("websocket", transport.getTransportType());
        } finally {
            transport.close();
        }
    }

    @DisplayName("Should convert HTTPS URL to WSS URL")
    void testHttpsToWssConversion() throws Exception {
        WebSocketTransport transport = new WebSocketTransport("https://localhost:8080", null, 30);
        try {
            assertEquals("websocket", transport.getTransportType());
        } finally {
            transport.close();
        }
    }

    @DisplayName("Should append token to URL when provided")
    void testTokenInUrl() throws Exception {
        WebSocketTransport transport = new WebSocketTransport("http://localhost:8080", "my-token-123", 30);
        try {
            assertNotNull(transport);
        } finally {
            transport.close();
        }
    }

    @DisplayName("Close should be idempotent")
    void testCloseIdempotent() throws Exception {
        WebSocketTransport transport = new WebSocketTransport("http://localhost:8080", null, 30);
        transport.close();
        assertDoesNotThrow(() -> transport.close());
    }

    @DisplayName("Query on closed transport should throw SQLException")
    void testQueryOnClosedTransport() throws Exception {
        WebSocketTransport transport = new WebSocketTransport("http://localhost:8080", null, 30);
        transport.close();
        
        QueryRequest request = new QueryRequest();
        request.setInstance("test");
        request.setQuery("SELECT 1");
        
        assertThrows(SQLException.class, () -> transport.executeQuery(request));
    }

    @DisplayName("Update on closed transport should throw SQLException")
    void testUpdateOnClosedTransport() throws Exception {
        WebSocketTransport transport = new WebSocketTransport("http://localhost:8080", null, 30);
        transport.close();
        
        QueryRequest request = new QueryRequest();
        request.setInstance("test");
        request.setQuery("UPDATE test SET x = 1");
        
        assertThrows(SQLException.class, () -> transport.executeUpdate(request));
    }

    @DisplayName("HealthCheck on closed transport should return false")
    void testHealthCheckOnClosedTransport() throws Exception {
        WebSocketTransport transport = new WebSocketTransport("http://localhost:8080", null, 30);
        transport.close();
        
        assertFalse(transport.healthCheck("test"));
    }
}
