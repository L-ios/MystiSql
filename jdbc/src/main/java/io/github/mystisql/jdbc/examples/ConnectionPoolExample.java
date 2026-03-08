package io.github.mystisql.jdbc.examples;

import com.zaxxer.hikari.HikariConfig;
import com.zaxxer.hikari.HikariDataSource;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;

/**
 * Example demonstrating HikariCP connection pool usage.
 * 
 * Note: This example requires HikariCP dependency.
 * Add to build.gradle.kts:
 *   implementation("com.zaxxer:HikariCP:5.0.1")
 */
public class ConnectionPoolExample {
    
    public static void main(String[] args) {
        // Configure HikariCP
        HikariConfig config = new HikariConfig();
        
        // JDBC URL
        config.setJdbcUrl("jdbc:mystisql://gateway.example.com:8080/production-mysql");
        
        // Credentials (use token as password)
        config.setUsername("app-user");
        config.setPassword("your-api-token");
        
        // Connection pool settings
        config.setMaximumPoolSize(10);
        config.setMinimumIdle(5);
        config.setConnectionTimeout(30000); // 30 seconds
        config.setIdleTimeout(600000);      // 10 minutes
        config.setMaxLifetime(1800000);     // 30 minutes
        
        // Connection validation
        config.setConnectionTestQuery("SELECT 1");
        config.setValidationTimeout(5000);  // 5 seconds
        
        // Pool name
        config.setPoolName("MystiSqlPool");
        
        // Create data source
        try (HikariDataSource dataSource = new HikariDataSource(config)) {
            
            System.out.println("Connection pool created: " + config.getPoolName());
            
            // Example 1: Simple query
            System.out.println("\nExample 1: Simple query from pool");
            try (Connection conn = dataSource.getConnection()) {
                System.out.println("Got connection from pool");
                
                try (PreparedStatement stmt = conn.prepareStatement("SELECT COUNT(*) FROM users")) {
                    ResultSet rs = stmt.executeQuery();
                    if (rs.next()) {
                        System.out.println("Total users: " + rs.getInt(1));
                    }
                    rs.close();
                }
            }
            
            // Example 2: Multiple concurrent operations
            System.out.println("\nExample 2: Multiple operations");
            for (int i = 0; i < 5; i++) {
                final int threadId = i;
                new Thread(() -> {
                    try (Connection conn = dataSource.getConnection();
                         PreparedStatement stmt = conn.prepareStatement("SELECT ? AS thread_id")) {
                        
                        stmt.setInt(1, threadId);
                        ResultSet rs = stmt.executeQuery();
                        
                        if (rs.next()) {
                            System.out.println("Thread " + threadId + " executed query");
                        }
                        rs.close();
                        
                    } catch (SQLException e) {
                        e.printStackTrace();
                    }
                }).start();
            }
            
            // Wait for threads
            Thread.sleep(2000);
            
            // Pool statistics
            System.out.println("\nPool statistics:");
            System.out.println("  Active connections: " + dataSource.getHikariPoolMXBean().getActiveConnections());
            System.out.println("  Idle connections: " + dataSource.getHikariPoolMXBean().getIdleConnections());
            System.out.println("  Total connections: " + dataSource.getHikariPoolMXBean().getTotalConnections());
            
            System.out.println("\nConnection pool example completed!");
            
        } catch (SQLException e) {
            System.err.println("SQL Error: " + e.getMessage());
            e.printStackTrace();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }
}
