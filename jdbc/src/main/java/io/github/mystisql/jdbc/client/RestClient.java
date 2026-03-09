package io.github.mystisql.jdbc.client;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import io.github.mystisql.jdbc.MystiSqlResultSet;
import okhttp3.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.sql.SQLException;
import java.sql.Types;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.TimeUnit;

/**
 * RESTful API client for MystiSql Gateway.
 */
public class RestClient {
    
    private static final Logger logger = LoggerFactory.getLogger(RestClient.class);
    private static final MediaType JSON = MediaType.get("application/json; charset=utf-8");
    private static final ObjectMapper objectMapper = new ObjectMapper()
        .registerModule(new JavaTimeModule());
    
    private final OkHttpClient httpClient;
    private final String baseUrl;
    private final String token;
    
    /**
     * Create RestClient.
     * 
     * @param baseUrl Gateway base URL (e.g., "http://localhost:8080")
     * @param token Authentication token (nullable)
     * @param timeoutSeconds Request timeout in seconds
     */
    public RestClient(String baseUrl, String token, int timeoutSeconds) {
        this.baseUrl = baseUrl.replaceAll("/$", "");
        this.token = token;
        
        this.httpClient = new OkHttpClient.Builder()
            .connectTimeout(timeoutSeconds, TimeUnit.SECONDS)
            .readTimeout(timeoutSeconds, TimeUnit.SECONDS)
            .writeTimeout(timeoutSeconds, TimeUnit.SECONDS)
            .connectionPool(new ConnectionPool(20, 5, TimeUnit.MINUTES))
            .build();
        
        logger.debug("RestClient initialized: baseUrl={}, timeout={}s", baseUrl, timeoutSeconds);
    }
    
    /**
     * Execute SQL query.
     * 
     * @param instance Database instance name
     * @param sql SQL query
     * @return ResultSet with query results
     * @throws SQLException if query fails
     */
    public MystiSqlResultSet executeQuery(String instance, String sql) throws SQLException {
        QueryRequest request = new QueryRequest();
        request.setInstance(instance);
        request.setQuery(sql);
        return executeQuery(request);
    }
    
    /**
     * Execute SQL query with parameters.
     * 
     * @param request Query request with parameters
     * @return ResultSet with query results
     * @throws SQLException if query fails
     */
    public MystiSqlResultSet executeQuery(QueryRequest request) throws SQLException {
        logger.debug("Executing query on instance: {}", request.getInstance());
        
        try {
            String url = buildUrl("/api/v1/query");
            String json = objectMapper.writeValueAsString(request);
            
            String responseBody = post(url, json);
            ApiResponse<QueryResult> response = objectMapper.readValue(
                responseBody,
                objectMapper.getTypeFactory().constructParametricType(
                    ApiResponse.class, QueryResult.class
                )
            );
            
            if (!response.isSuccess()) {
                throw createSQLException(response.getError());
            }
            
            return convertToResultSet(response.getData());
            
        } catch (IOException e) {
            logger.error("Query execution failed", e);
            throw new SQLException("Query execution failed: " + e.getMessage(), "08000", e);
        }
    }
    
    /**
     * Execute SQL update (INSERT/UPDATE/DELETE).
     * 
     * @param instance Database instance name
     * @param sql SQL statement
     * @return Execution result with affected rows
     * @throws SQLException if execution fails
     */
    public ExecResult executeUpdate(String instance, String sql) throws SQLException {
        QueryRequest request = new QueryRequest();
        request.setInstance(instance);
        request.setQuery(sql);
        return executeUpdate(request);
    }
    
    /**
     * Execute SQL update with parameters.
     * 
     * @param request Query request with parameters
     * @return Execution result with affected rows
     * @throws SQLException if execution fails
     */
    public ExecResult executeUpdate(QueryRequest request) throws SQLException {
        logger.debug("Executing update on instance: {}", request.getInstance());
        
        try {
            String url = buildUrl("/api/v1/exec");
            String json = objectMapper.writeValueAsString(request);
            
            String responseBody = post(url, json);
            ApiResponse<ExecResult> response = objectMapper.readValue(
                responseBody,
                objectMapper.getTypeFactory().constructParametricType(
                    ApiResponse.class, ExecResult.class
                )
            );
            
            if (!response.isSuccess()) {
                throw createSQLException(response.getError());
            }
            
            return response.getData();
            
        } catch (IOException e) {
            logger.error("Update execution failed", e);
            throw new SQLException("Update execution failed: " + e.getMessage(), "08000", e);
        }
    }
    
    /**
     * Execute health check.
     * 
     * @param instance Database instance name
     * @return true if instance is healthy
     * @throws SQLException if health check fails
     */
    public boolean healthCheck(String instance) throws SQLException {
        try {
            String url = buildUrl("/api/v1/instances/" + instance + "/health");
            String responseBody = get(url);
            
            // Simple check - if we get 200 OK, instance is healthy
            return responseBody != null && responseBody.contains("healthy");
            
        } catch (IOException e) {
            logger.error("Health check failed", e);
            return false;
        }
    }
    
    /**
     * Build full URL from path.
     * 
     * @param path API path (e.g., "/api/v1/query")
     * @return Full URL
     */
    public String buildUrl(String path) {
        return baseUrl + path;
    }
    
    /**
     * POST request.
     */
    private String post(String url, String json) throws IOException, SQLException {
        RequestBody body = RequestBody.create(json, JSON);
        Request.Builder requestBuilder = new Request.Builder()
            .url(url)
            .post(body);
        
        // Add authorization header if token is set
        if (token != null && !token.isEmpty()) {
            requestBuilder.addHeader("Authorization", "Bearer " + token);
        }
        
        Request request = requestBuilder.build();
        
        try (Response response = httpClient.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                String errorBody = response.body() != null ? response.body().string() : "Unknown error";
                logger.error("HTTP request failed: {} - {}", response.code(), errorBody);
                throw new SQLException("HTTP request failed with status " + response.code() + ": " + errorBody);
            }
            
            return response.body() != null ? response.body().string() : "{}";
        }
    }
    
    /**
     * GET request.
     */
    private String get(String url) throws IOException, SQLException {
        Request.Builder requestBuilder = new Request.Builder()
            .url(url)
            .get();
        
        if (token != null && !token.isEmpty()) {
            requestBuilder.addHeader("Authorization", "Bearer " + token);
        }
        
        Request request = requestBuilder.build();
        
        try (Response response = httpClient.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new SQLException("HTTP request failed with status " + response.code());
            }
            
            return response.body() != null ? response.body().string() : "{}";
        }
    }
    
    /**
     * Convert QueryResult to ResultSet.
     */
    private MystiSqlResultSet convertToResultSet(QueryResult queryResult) {
        if (queryResult == null || queryResult.getColumns() == null) {
            return new MystiSqlResultSet(new MystiSqlResultSet.Column[0], new Object[0][]);
        }
        
        // Convert columns
        MystiSqlResultSet.Column[] columns = new MystiSqlResultSet.Column[queryResult.getColumns().size()];
        for (int i = 0; i < queryResult.getColumns().size(); i++) {
            QueryResult.ColumnInfo col = queryResult.getColumns().get(i);
            columns[i] = new MystiSqlResultSet.Column(
                col.getName(),
                mapType(col.getType()),
                col.getType()
            );
        }
        
        // Convert rows
        Object[][] rows = new Object[queryResult.getRows() != null ? queryResult.getRows().size() : 0][];
        if (queryResult.getRows() != null) {
            for (int i = 0; i < queryResult.getRows().size(); i++) {
                List<Object> row = queryResult.getRows().get(i);
                rows[i] = row != null ? row.toArray() : new Object[0];
            }
        }
        
        return new MystiSqlResultSet(columns, rows);
    }
    
    /**
     * Map database type to JDBC type.
     */
    private int mapType(String typeName) {
        if (typeName == null) return Types.VARCHAR;
        
        switch (typeName.toUpperCase()) {
            case "TINYINT": return Types.TINYINT;
            case "SMALLINT": return Types.SMALLINT;
            case "INT":
            case "INTEGER": return Types.INTEGER;
            case "BIGINT": return Types.BIGINT;
            case "FLOAT": return Types.FLOAT;
            case "DOUBLE": return Types.DOUBLE;
            case "DECIMAL":
            case "NUMERIC": return Types.DECIMAL;
            case "CHAR": return Types.CHAR;
            case "VARCHAR": return Types.VARCHAR;
            case "TEXT":
            case "LONGTEXT": return Types.LONGVARCHAR;
            case "DATE": return Types.DATE;
            case "TIME": return Types.TIME;
            case "DATETIME":
            case "TIMESTAMP": return Types.TIMESTAMP;
            case "BLOB":
            case "BINARY": return Types.BLOB;
            case "BOOLEAN":
            case "BOOL": return Types.BOOLEAN;
            default: return Types.OTHER;
        }
    }
    
    /**
     * Create SQLException from error response.
     */
    private SQLException createSQLException(ApiResponse.ErrorInfo error) {
        if (error == null) {
            return new SQLException("Unknown error");
        }
        
        String sqlState = mapErrorCodeToSQLState(error.getCode());
        int vendorCode = 0;
        try {
            if (error.getCode() != null) {
                vendorCode = Integer.parseInt(error.getCode().replaceAll("[^0-9]", ""));
            }
        } catch (NumberFormatException e) {
            // Ignore
        }
        return new SQLException(error.getMessage(), sqlState, vendorCode);
    }
    
    /**
     * Map error code to SQLState.
     */
    private String mapErrorCodeToSQLState(String errorCode) {
        if (errorCode == null) return "HY000";
        
        switch (errorCode.toUpperCase()) {
            case "TABLE_NOT_FOUND": return "42S02";
            case "COLUMN_NOT_FOUND": return "42S22";
            case "SYNTAX_ERROR": return "42000";
            case "CONNECTION_FAILED": return "08001";
            case "CONNECTION_TIMEOUT": return "HYT00";
            case "PERMISSION_DENIED": return "42000";
            case "DUPLICATE_ENTRY": return "23000";
            case "FOREIGN_KEY_CONSTRAINT": return "23503";
            default: return "HY000";
        }
    }
    
    /**
     * Close HTTP client and release resources.
     */
    public void close() {
        httpClient.dispatcher().executorService().shutdown();
        httpClient.connectionPool().evictAll();
        logger.debug("RestClient closed");
    }
}
