package io.github.mystisql.jdbc.client;

import io.github.mystisql.jdbc.MystiSqlResultSet;
import okhttp3.mockwebserver.MockResponse;
import okhttp3.mockwebserver.MockWebServer;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;
import static org.junit.jupiter.api.Assertions.*;

import java.sql.SQLException;
import java.util.Arrays;

/**
 * Unit tests for RestClient
 */
class RestClientTest {

    private MockWebServer mockServer;
    private RestClient restClient;
    private String baseUrl;

    @BeforeEach
    void setUp() throws Exception {
        mockServer = new MockWebServer();
        mockServer.start();
        baseUrl = "http://localhost:" + mockServer.getPort();
        restClient = new RestClient(baseUrl, null, 30);
    }

    @AfterEach
    void tearDown() throws Exception {
        mockServer.shutdown();
    }

    @Test
    @DisplayName("Execute query should return result set")
    void testExecuteQuery() throws Exception {
        String responseBody = "{" +
            "\"success\": true," +
            "\"data\": {" +
            "\"columns\": [{\"name\": \"id\", \"type\": \"INT\"}, {\"name\": \"name\", \"type\": \"VARCHAR\"}]," +
            "\"rows\": [[1, \"Alice\"], [2, \"Bob\"]]," +
            "\"rowCount\": 2" +
            "}," +
            "\"executionTime\": 5000000" +
            "}";
        
        mockServer.enqueue(new MockResponse()
            .setBody(responseBody)
            .addHeader("Content-Type", "application/json"));

        MystiSqlResultSet result = restClient.executeQuery("production-mysql", "SELECT * FROM users");

        assertNotNull(result);
        assertEquals(2, result.getMetaData().getColumnCount());
        assertTrue(result.next());
        assertEquals(1, result.getInt("id"));
        assertEquals("Alice", result.getString("name"));
    }

    @Test
    @DisplayName("Execute query with parameters")
    void testExecuteQueryWithParameters() throws Exception {
        String responseBody = "{" +
            "\"success\": true," +
            "\"data\": {" +
            "\"columns\": [{\"name\": \"id\", \"type\": \"INT\"}, {\"name\": \"name\", \"type\": \"VARCHAR\"}]," +
            "\"rows\": [[1, \"Alice\"]]," +
            "\"rowCount\": 1" +
            "}" +
            "}";
        
        mockServer.enqueue(new MockResponse()
            .setBody(responseBody)
            .addHeader("Content-Type", "application/json"));

        QueryRequest request = new QueryRequest();
        request.setInstance("production-mysql");
        request.setQuery("SELECT * FROM users WHERE name = ?");
        request.setParameters(Arrays.asList(
            new QueryParameter("VARCHAR", "Alice")
        ));

        MystiSqlResultSet result = restClient.executeQuery(request);

        assertNotNull(result);
        assertTrue(result.next());
        assertEquals("Alice", result.getString("name"));
    }

    @Test
    @DisplayName("Execute update should return affected rows")
    void testExecuteUpdate() throws Exception {
        String responseBody = "{" +
            "\"success\": true," +
            "\"data\": {" +
            "\"rowsAffected\": 5," +
            "\"lastInsertId\": 123" +
            "}" +
            "}";
        
        mockServer.enqueue(new MockResponse()
            .setBody(responseBody)
            .addHeader("Content-Type", "application/json"));

        ExecResult result = restClient.executeUpdate("production-mysql", "UPDATE users SET active = true");

        assertNotNull(result);
        assertEquals(5L, result.getRowsAffected().longValue());
        assertEquals(123L, result.getLastInsertId().longValue());
    }

    @Test
    @DisplayName("Should send authorization header")
    void testAuthorizationHeader() throws Exception {
        RestClient authClient = new RestClient(baseUrl, "test-token-123", 30);
        
        String responseBody = "{\"success\": true, \"data\": {\"columns\": [], \"rows\": [], \"rowCount\": 0}}";
        mockServer.enqueue(new MockResponse()
            .setBody(responseBody)
            .addHeader("Content-Type", "application/json"));

        authClient.executeQuery("instance", "SELECT 1");

        var recordedRequest = mockServer.takeRequest();
        assertEquals("Bearer test-token-123", recordedRequest.getHeader("Authorization"));
    }

    @Test
    @DisplayName("Should handle HTTP error")
    void testHttpError() {
        mockServer.enqueue(new MockResponse()
            .setResponseCode(500)
            .setBody("{\"error\": \"Internal server error\", \"code\": \"INTERNAL_ERROR\"}")
            .addHeader("Content-Type", "application/json"));

        assertThrows(SQLException.class, () -> 
            restClient.executeQuery("instance", "SELECT 1")
        );
    }

    @Test
    @DisplayName("Should handle Gateway error response")
    void testGatewayError() {
        String responseBody = "{" +
            "\"success\": false," +
            "\"error\": {" +
            "\"code\": \"TABLE_NOT_FOUND\"," +
            "\"message\": \"Table 'users' does not exist\"" +
            "}" +
            "}";
        
        mockServer.enqueue(new MockResponse()
            .setBody(responseBody)
            .addHeader("Content-Type", "application/json"));

        SQLException exception = assertThrows(SQLException.class, () -> 
            restClient.executeQuery("instance", "SELECT * FROM users")
        );
        
        assertTrue(exception.getMessage().contains("Table 'users' does not exist"));
        assertEquals("42S02", exception.getSQLState());
    }

    @Test
    @DisplayName("Should build correct URL")
    void testBuildUrl() {
        assertEquals(baseUrl + "/api/v1/query", restClient.buildUrl("/api/v1/query"));
        assertEquals(baseUrl + "/api/v1/exec", restClient.buildUrl("/api/v1/exec"));
    }

    @Test
    @DisplayName("Close should cleanup resources")
    void testClose() {
        RestClient client = new RestClient(baseUrl, null, 30);
        assertDoesNotThrow(() -> client.close());
    }
}
