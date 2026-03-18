package io.github.mystisql.jdbc.client;

import io.github.mystisql.jdbc.MystiSqlResultSet;

import java.sql.SQLException;

/**
 * 传输层接口，抽象 HTTP 和 WebSocket 通信方式。
 * 
 * <p>支持两种传输实现：</p>
 * <ul>
 *   <li>{@link RestClient} - HTTP RESTful API 传输</li>
 *   <li>{@link WebSocketTransport} - WebSocket 实时双向通信</li>
 * </ul>
 */
public interface Transport extends AutoCloseable {
    
    /**
     * 执行 SQL 查询（SELECT）。
     *
     * @param instance 数据库实例名称
     * @param sql SQL 查询语句
     * @return 查询结果集
     * @throws SQLException 如果查询失败
     */
    MystiSqlResultSet executeQuery(String instance, String sql) throws SQLException;
    
    /**
     * 执行带参数的 SQL 查询。
     *
     * @param request 查询请求（包含实例名、SQL、参数）
     * @return 查询结果集
     * @throws SQLException 如果查询失败
     */
    MystiSqlResultSet executeQuery(QueryRequest request) throws SQLException;
    
    /**
     * 执行 SQL 更新（INSERT/UPDATE/DELETE）。
     *
     * @param instance 数据库实例名称
     * @param sql SQL 更新语句
     * @return 执行结果（包含受影响行数）
     * @throws SQLException 如果执行失败
     */
    ExecResult executeUpdate(String instance, String sql) throws SQLException;
    
    /**
     * 执行带参数的 SQL 更新。
     *
     * @param request 查询请求（包含实例名、SQL、参数）
     * @return 执行结果（包含受影响行数）
     * @throws SQLException 如果执行失败
     */
    ExecResult executeUpdate(QueryRequest request) throws SQLException;
    
    /**
     * 健康检查，验证连接是否有效。
     *
     * @param instance 数据库实例名称
     * @return true 如果实例健康
     * @throws SQLException 如果检查失败
     */
    boolean healthCheck(String instance) throws SQLException;
    
    /**
     * 获取传输类型名称。
     *
     * @return 传输类型（"http" 或 "websocket"）
     */
    String getTransportType();
    
    /**
     * 关闭传输层并释放资源。
     */
    @Override
    void close();
}
