package io.github.mystisql.jdbc.client;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import io.github.mystisql.jdbc.MystiSqlResultSet;
import org.java_websocket.client.WebSocketClient;
import org.java_websocket.handshake.ServerHandshake;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.URI;
import java.sql.SQLException;
import java.sql.Types;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicInteger;

/**
 * WebSocket 传输实现，通过 WebSocket 与 MystiSql Gateway 通信。
 */
public class WebSocketTransport implements Transport {
    
    private static final Logger logger = LoggerFactory.getLogger(WebSocketTransport.class);
    private static final ObjectMapper objectMapper = new ObjectMapper()
        .registerModule(new JavaTimeModule());
    
    private static final int MAX_RECONNECT_ATTEMPTS = 3;
    private static final long[] RECONNECT_INTERVALS = {1000, 2000, 4000};
    private static final long HEARTBEAT_INTERVAL = 30000;
    private static final long HEARTBEAT_TIMEOUT = 10000;
    
    private final String wsUrl;
    private final String token;
    private final int timeout;
    private final AtomicBoolean closed = new AtomicBoolean(false);
    private final AtomicBoolean connected = new AtomicBoolean(false);
    private final Map<String, CompletableFuture<WsMessage>> pendingRequests = new ConcurrentHashMap<>();
    private final AtomicInteger reconnectAttempts = new AtomicInteger(0);
    
    private volatile WebSocketClient wsClient;
    private volatile Thread heartbeatThread;
    
    public WebSocketTransport(String baseUrl, String token, int timeoutSeconds) {
        String url = baseUrl.replaceFirst("^http://", "ws://")
                            .replaceFirst("^https://", "wss://");
        if (token != null && !token.isEmpty()) {
            this.wsUrl = url + "/ws?token=" + token;
        } else {
            this.wsUrl = url + "/ws";
        }
        this.token = token;
        this.timeout = timeoutSeconds;
        
        logger.debug("WebSocketTransport initialized: wsUrl={}, timeout={}s", 
            wsUrl.replaceAll("token=[^&]*", "token=***"), timeoutSeconds);
    }
    
    private synchronized void ensureConnected() throws SQLException {
        if (closed.get()) {
            throw new SQLException("Transport is closed");
        }
        
        if (connected.get() && wsClient != null) {
            return;
        }
        
        connect();
    }
    
    private void connect() throws SQLException {
        int attempts = 0;
        while (attempts < MAX_RECONNECT_ATTEMPTS) {
            try {
                doConnect();
                connected.set(true);
                reconnectAttempts.set(0);
                startHeartbeat();
                logger.debug("WebSocket connected successfully");
                return;
            } catch (Exception e) {
                attempts++;
                logger.warn("WebSocket connection attempt {} failed: {}", attempts, e.getMessage());
                if (attempts < MAX_RECONNECT_ATTEMPTS) {
                    try {
                        Thread.sleep(RECONNECT_INTERVALS[attempts - 1]);
                    } catch (InterruptedException ie) {
                        Thread.currentThread().interrupt();
                        throw new SQLException("Connection interrupted", ie);
                    }
                }
            }
        }
        
        throw new SQLException("Failed to connect to WebSocket after " + MAX_RECONNECT_ATTEMPTS + " attempts");
    }
    
    private void doConnect() throws Exception {
        CompletableFuture<Void> connectFuture = new CompletableFuture<>();
        
        wsClient = new WebSocketClient(URI.create(wsUrl)) {
            @Override
            public void onOpen(ServerHandshake handshake) {
                logger.debug("WebSocket connection opened");
                connectFuture.complete(null);
            }
            
            @Override
            public void onMessage(String message) {
                handleMessage(message);
            }
            
            @Override
            public void onClose(int code, String reason, boolean remote) {
                logger.debug("WebSocket closed: code={}, reason={}, remote={}", code, reason, remote);
                connected.set(false);
                stopHeartbeat();
                pendingRequests.values().forEach(f -> 
                    f.completeExceptionally(new SQLException("Connection closed")));
                pendingRequests.clear();
            }
            
            @Override
            public void onError(Exception ex) {
                logger.error("WebSocket error", ex);
                if (!connectFuture.isDone()) {
                    connectFuture.completeExceptionally(ex);
                }
            }
        };
        
        wsClient.connect();
        connectFuture.get(timeout, TimeUnit.SECONDS);
    }
    
    private void handleMessage(String message) {
        try {
            WsMessage wsMessage = objectMapper.readValue(message, WsMessage.class);
            String requestId = wsMessage.getRequestId();
            
            if (requestId != null) {
                CompletableFuture<WsMessage> future = pendingRequests.remove(requestId);
                if (future != null) {
                    future.complete(wsMessage);
                }
            } else if ("pong".equals(wsMessage.getType())) {
                logger.debug("Received pong");
            }
        } catch (Exception e) {
            logger.error("Failed to handle WebSocket message", e);
        }
    }
    
    private void startHeartbeat() {
        stopHeartbeat();
        heartbeatThread = new Thread(() -> {
            while (!closed.get() && connected.get() && !Thread.currentThread().isInterrupted()) {
                try {
                    Thread.sleep(HEARTBEAT_INTERVAL);
                    if (connected.get() && !closed.get()) {
                        sendPing();
                    }
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    break;
                } catch (Exception e) {
                    logger.warn("Heartbeat failed", e);
                }
            }
        }, "ws-heartbeat");
        heartbeatThread.setDaemon(true);
        heartbeatThread.start();
    }
    
    private void stopHeartbeat() {
        if (heartbeatThread != null) {
            heartbeatThread.interrupt();
            heartbeatThread = null;
        }
    }
    
    private void sendPing() throws SQLException {
        try {
            String pingMsg = objectMapper.writeValueAsString(Map.of("action", "ping"));
            wsClient.send(pingMsg);
            logger.debug("Sent ping");
        } catch (Exception e) {
            logger.warn("Failed to send ping", e);
        }
    }
    
    private WsMessage sendRequest(String instance, String sql, String action) throws SQLException {
        ensureConnected();
        
        String requestId = UUID.randomUUID().toString();
        CompletableFuture<WsMessage> future = new CompletableFuture<>();
        pendingRequests.put(requestId, future);
        
        try {
            Map<String, Object> request = new java.util.HashMap<>();
            request.put("action", action);
            request.put("instance", instance);
            request.put("query", sql);
            request.put("requestId", requestId);
            
            String json = objectMapper.writeValueAsString(request);
            wsClient.send(json);
            
            return future.get(timeout, TimeUnit.SECONDS);
        } catch (Exception e) {
            pendingRequests.remove(requestId);
            throw new SQLException("WebSocket request failed: " + e.getMessage(), e);
        }
    }
    
    @Override
    public MystiSqlResultSet executeQuery(String instance, String sql) throws SQLException {
        QueryRequest request = new QueryRequest();
        request.setInstance(instance);
        request.setQuery(sql);
        return executeQuery(request);
    }
    
    @Override
    public MystiSqlResultSet executeQuery(QueryRequest request) throws SQLException {
        logger.debug("Executing query via WebSocket: {}", request.getInstance());
        
        WsMessage response = sendRequest(request.getInstance(), request.getQuery(), "query");
        
        if (!"query_result".equals(response.getType())) {
            throw createSQLException(response);
        }
        
        return convertToResultSet(response);
    }
    
    @Override
    public ExecResult executeUpdate(String instance, String sql) throws SQLException {
        QueryRequest request = new QueryRequest();
        request.setInstance(instance);
        request.setQuery(sql);
        return executeUpdate(request);
    }
    
    @Override
    public ExecResult executeUpdate(QueryRequest request) throws SQLException {
        logger.debug("Executing update via WebSocket: {}", request.getInstance());
        
        WsMessage response = sendRequest(request.getInstance(), request.getQuery(), "query");
        
        if (!"query_result".equals(response.getType())) {
            throw createSQLException(response);
        }
        
        ExecResult result = new ExecResult();
        result.setRowsAffected(response.getRowCount() != null ? response.getRowCount().longValue() : 0L);
        result.setLastInsertId(response.getLastInsertId());
        return result;
    }
    
    @Override
    public boolean healthCheck(String instance) throws SQLException {
        try {
            ensureConnected();
            sendPing();
            return connected.get();
        } catch (Exception e) {
            return false;
        }
    }
    
    private MystiSqlResultSet convertToResultSet(WsMessage response) {
        if (response.getColumns() == null || response.getColumns().isEmpty()) {
            return new MystiSqlResultSet(new MystiSqlResultSet.Column[0], new Object[0][]);
        }
        
        MystiSqlResultSet.Column[] columns = new MystiSqlResultSet.Column[response.getColumns().size()];
        for (int i = 0; i < response.getColumns().size(); i++) {
            WsMessage.ColumnInfo col = response.getColumns().get(i);
            columns[i] = new MystiSqlResultSet.Column(
                col.getName(),
                mapType(col.getType()),
                col.getType()
            );
        }
        
        Object[][] rows = new Object[response.getRows() != null ? response.getRows().size() : 0][];
        if (response.getRows() != null) {
            for (int i = 0; i < response.getRows().size(); i++) {
                Map<String, Object> row = response.getRows().get(i);
                Object[] ordered = new Object[columns.length];
                for (int j = 0; j < columns.length; j++) {
                    ordered[j] = row != null ? row.get(columns[j].getName()) : null;
                }
                rows[i] = ordered;
            }
        }
        
        return new MystiSqlResultSet(columns, rows);
    }
    
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
    
    private SQLException createSQLException(WsMessage error) {
        if (error == null) {
            return new SQLException("Unknown error");
        }
        
        String sqlState = mapErrorCodeToSQLState(error.getCode());
        return new SQLException(error.getMessage() != null ? error.getMessage() : "Unknown error", sqlState);
    }
    
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
    
    @Override
    public String getTransportType() {
        return "websocket";
    }
    
    @Override
    public void close() {
        if (closed.compareAndSet(false, true)) {
            stopHeartbeat();
            pendingRequests.values().forEach(f -> 
                f.completeExceptionally(new SQLException("Transport closed")));
            pendingRequests.clear();
            
            if (wsClient != null) {
                try {
                    wsClient.close();
                } catch (Exception e) {
                    logger.debug("Error closing WebSocket", e);
                }
            }
            logger.debug("WebSocketTransport closed");
        }
    }
    
    private static class WsMessage {
        private String requestId;
        private String type;
        private List<ColumnInfo> columns;
        private List<Map<String, Object>> rows;
        private Integer rowCount;
        private Long rowsAffected;
        private Long lastInsertId;
        private String code;
        private String message;
        
        public String getRequestId() { return requestId; }
        public void setRequestId(String requestId) { this.requestId = requestId; }
        public String getType() { return type; }
        public void setType(String type) { this.type = type; }
        public List<ColumnInfo> getColumns() { return columns; }
        public void setColumns(List<ColumnInfo> columns) { this.columns = columns; }
        public List<Map<String, Object>> getRows() { return rows; }
        public void setRows(List<Map<String, Object>> rows) { this.rows = rows; }
        public Integer getRowCount() { return rowCount; }
        public void setRowCount(Integer rowCount) { this.rowCount = rowCount; }
        public Long getRowsAffected() { return rowsAffected; }
        public void setRowsAffected(Long rowsAffected) { this.rowsAffected = rowsAffected; }
        public Long getLastInsertId() { return lastInsertId; }
        public void setLastInsertId(Long lastInsertId) { this.lastInsertId = lastInsertId; }
        public String getCode() { return code; }
        public void setCode(String code) { this.code = code; }
        public String getMessage() { return message; }
        public void setMessage(String message) { this.message = message; }
        
        public static class ColumnInfo {
            private String name;
            private String type;
            public String getName() { return name; }
            public void setName(String name) { this.name = name; }
            public String getType() { return type; }
            public void setType(String type) { this.type = type; }
        }
        
    }
}
